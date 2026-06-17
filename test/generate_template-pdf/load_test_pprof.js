/**
 * Steady-state k6 load for Gin HTTP pprof capture.
 *
 * Env:
 *   BASE_URL          — default http://localhost:8080
 *   PAYLOAD_SCENARIO  — tagged_ecdsa (default) | tagged_rsa | retail_only_signed | retail_active_signed | unsigned
 *   PROFILE_SECONDS   — steady VU duration (default 35)
 *   LOAD_VUS          — concurrent VUs (default 48)
 *   SKIP_SMOKE        — set "1" to skip 1-VU smoke phase
 *   THROUGHPUT_GATE   — min req/s for http_reqs threshold (default 0 = disabled)
 *
 * Makefile targets:
 *   make bench-k6       — 48 VU × 35s (full harness)
 *   make bench-k6-light — 24 VU × 15s, MAX_CONCURRENT=24, GOMAXPROCS=12
 */
import http from 'k6/http';
import { check } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { getPayloadOptions, getWeightedPayloadWithTier } from './payload_generator.js';

const errorRate = new Rate('errors');
const pdfGenerationTime = new Trend('pdf_generation_time');
const hftLatency = new Trend('hft_latency');
const retailLatency = new Trend('retail_latency');
const activeLatency = new Trend('active_latency');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const PAYLOAD_SCENARIO = __ENV.PAYLOAD_SCENARIO || 'tagged_ecdsa';
const PROFILE_SECONDS = parseInt(__ENV.PROFILE_SECONDS || '35', 10);
const LOAD_VUS = parseInt(__ENV.LOAD_VUS || '48', 10);
const SKIP_SMOKE = __ENV.SKIP_SMOKE === '1';
const THROUGHPUT_GATE = parseInt(__ENV.THROUGHPUT_GATE || '0', 10);

const payloadOpts = getPayloadOptions(PAYLOAD_SCENARIO);

const headers = {
    'Accept': '*/*',
    'Content-Type': 'application/json',
    'Origin': BASE_URL,
    'Referer': `${BASE_URL}/gopdfsuit`,
};

const thresholds = {
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.05'],
    http_req_duration: ['p(99)<5000'],
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
                  tags: { test_type: 'pprof_steady' },
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
                  tags: { test_type: 'pprof_steady' },
              },
          },
    thresholds,
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

export default function () {
    const url = `${BASE_URL}/api/v1/generate/template-pdf`;
    const { payload, tier } = getWeightedPayloadWithTier(payloadOpts);
    const reqHeaders = { ...headers, 'X-Payload-Tier': tier };

    const startTime = Date.now();
    const response = http.post(url, JSON.stringify(payload), { headers: reqHeaders });
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
        'response has content': (r) => r.body && r.body.length > 0,
    });
    errorRate.add(!ok);
}

export function setup() {
    console.log(`pprof load: ${LOAD_VUS} VUs × ${PROFILE_SECONDS}s, scenario=${PAYLOAD_SCENARIO}`);
    const health = http.get(`${BASE_URL}/gopdfsuit`);
    if (health.status !== 200) {
        console.warn(`health check status ${health.status}`);
    }
    return { startTime: new Date().toISOString(), payloadScenario: PAYLOAD_SCENARIO };
}

export function teardown(data) {
    console.log(`pprof load done (started ${data.startTime})`);
}