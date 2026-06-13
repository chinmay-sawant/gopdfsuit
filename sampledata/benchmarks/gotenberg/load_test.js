/**
 * Gotenberg weighted HTML→PDF load test (no smoke phase).
 * Same env vars as load_test_pprof.js; defaults to 35s × 48 VUs.
 */
import http from 'k6/http';
import { check } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { getPayloadOptions, getWeightedHTMLWithTier } from './html_payload_generator.js';

const errorRate = new Rate('errors');
const pdfGenerationTime = new Trend('pdf_generation_time');

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:3010';
const PAYLOAD_SCENARIO = __ENV.PAYLOAD_SCENARIO || 'tagged_ecdsa';
const PROFILE_SECONDS = parseInt(__ENV.PROFILE_SECONDS || '35', 10);
const LOAD_VUS = parseInt(__ENV.LOAD_VUS || '48', 10);
const SKIP_NETWORK_IDLE = __ENV.SKIP_NETWORK_IDLE !== '0';

const payloadOpts = getPayloadOptions(PAYLOAD_SCENARIO);
const convertURL = `${BASE_URL}/forms/chromium/convert/html`;

export const options = {
    scenarios: {
        steady: {
            executor: 'constant-vus',
            vus: LOAD_VUS,
            duration: `${PROFILE_SECONDS}s`,
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.05'],
        errors: ['rate<0.05'],
    },
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

export default function () {
    const { html } = getWeightedHTMLWithTier(payloadOpts);
    const formData = {
        files: http.file(html, 'index.html', 'text/html'),
        printBackground: 'true',
    };
    if (SKIP_NETWORK_IDLE) {
        formData.skipNetworkIdleEvent = 'true';
    }

    const startTime = Date.now();
    const response = http.post(convertURL, formData);
    pdfGenerationTime.add(Date.now() - startTime);

    const ok = check(response, {
        'status is 200': (r) => r.status === 200,
        'response has content': (r) => r.body && r.body.length > 1000,
    });
    errorRate.add(!ok);
}