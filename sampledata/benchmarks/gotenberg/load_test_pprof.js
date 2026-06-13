/**
 * Steady-state k6 load for Gotenberg Chromium HTML→PDF.
 * Mirrors test/generate_template-pdf/load_test_pprof.js structure.
 *
 * Env:
 *   BASE_URL              — default http://127.0.0.1:3000
 *   PAYLOAD_SCENARIO      — tagged_ecdsa (default) | retail_only_signed | retail_active_signed | unsigned
 *   PROFILE_SECONDS       — steady VU duration (default 35)
 *   LOAD_VUS              — concurrent VUs (default 48)
 *   SKIP_SMOKE            — set "1" to skip 1-VU smoke phase
 *   THROUGHPUT_GATE       — min req/s for http_reqs threshold (default 0 = disabled)
 *   SKIP_NETWORK_IDLE     — set "1" to pass skipNetworkIdleEvent=true (faster, default on)
 */
import http from 'k6/http';
import { check } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { getPayloadOptions, getWeightedHTMLWithTier } from './html_payload_generator.js';

const errorRate = new Rate('errors');
const pdfGenerationTime = new Trend('pdf_generation_time');
const hftLatency = new Trend('hft_latency');
const retailLatency = new Trend('retail_latency');
const activeLatency = new Trend('active_latency');

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:3010';
const PAYLOAD_SCENARIO = __ENV.PAYLOAD_SCENARIO || 'tagged_ecdsa';
const PROFILE_SECONDS = parseInt(__ENV.PROFILE_SECONDS || '35', 10);
const LOAD_VUS = parseInt(__ENV.LOAD_VUS || '48', 10);
const SKIP_SMOKE = __ENV.SKIP_SMOKE === '1';
const THROUGHPUT_GATE = parseInt(__ENV.THROUGHPUT_GATE || '0', 10);
const SKIP_NETWORK_IDLE = __ENV.SKIP_NETWORK_IDLE !== '0';

const payloadOpts = getPayloadOptions(PAYLOAD_SCENARIO);
const convertURL = `${BASE_URL}/forms/chromium/convert/html`;

const thresholds = {
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.05'],
    http_req_duration: ['p(99)<30000'],
};

if (THROUGHPUT_GATE > 0) {
    thresholds.http_reqs = [`rate>${THROUGHPUT_GATE}`];
}

export const options = {
    scenarios: SKIP_SMOKE
        ? {
              steady: {
                  executor: 'constant-vus',
                  vus: LOAD_VUS,
                  duration: `${PROFILE_SECONDS}s`,
                  tags: { test_type: 'gotenberg_steady' },
              },
          }
        : {
              smoke: {
                  executor: 'constant-vus',
                  vus: 1,
                  duration: '3s',
                  tags: { test_type: 'smoke' },
              },
              steady: {
                  executor: 'constant-vus',
                  vus: LOAD_VUS,
                  duration: `${PROFILE_SECONDS}s`,
                  startTime: '5s',
                  tags: { test_type: 'gotenberg_steady' },
              },
          },
    thresholds,
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

function buildFormData(html) {
    const formData = {
        files: http.file(html, 'index.html', 'text/html'),
        paperWidth: '8.27',
        paperHeight: '11.7',
        marginTop: '0.47',
        marginBottom: '0.47',
        marginLeft: '0.47',
        marginRight: '0.47',
        printBackground: 'true',
    };
    if (SKIP_NETWORK_IDLE) {
        formData.skipNetworkIdleEvent = 'true';
    }
    return formData;
}

export default function () {
    const { html, tier } = getWeightedHTMLWithTier(payloadOpts);

    const startTime = Date.now();
    const response = http.post(convertURL, buildFormData(html));
    const elapsed = Date.now() - startTime;

    pdfGenerationTime.add(elapsed);
    if (tier === 'hft') {
        hftLatency.add(elapsed);
    } else if (tier === 'active') {
        activeLatency.add(elapsed);
    } else {
        retailLatency.add(elapsed);
    }

    const ok = check(response, {
        'status is 200': (r) => r.status === 200,
        'response is pdf': (r) => r.headers['Content-Type'] && r.headers['Content-Type'].includes('pdf'),
        'response has content': (r) => r.body && r.body.length > 1000,
    });
    errorRate.add(!ok);
}

export function setup() {
    console.log(
        `gotenberg load: ${LOAD_VUS} VUs × ${PROFILE_SECONDS}s, scenario=${PAYLOAD_SCENARIO}, skipNetworkIdle=${SKIP_NETWORK_IDLE}`
    );
    const health = http.get(`${BASE_URL}/health`);
    if (health.status !== 200) {
        console.warn(`gotenberg health status ${health.status}`);
    }
    return { startTime: new Date().toISOString(), payloadScenario: PAYLOAD_SCENARIO };
}

export function teardown(data) {
    console.log(`gotenberg load done (started ${data.startTime})`);
}