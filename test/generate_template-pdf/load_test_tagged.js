import http from 'k6/http';
import { check } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const pdfGenerationTime = new Trend('pdf_generation_time');

import { getPayloadOptions, getWeightedPayload } from './payload_generator.js';

// Same shape as load_test.js; stresses PDF/A-style config (pdfaCompliant + taggedPDF) with signing.
export const options = {
    scenarios: {
        smoke: {
            executor: 'constant-vus',
            vus: 1,
            duration: '5s',
            startTime: '0s',
            tags: { test_type: 'smoke' },
        },
        load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '5s', target: 48 },
                { duration: '10s', target: 48 },
                { duration: '5s', target: 0 },
            ],
            startTime: '10s',
            tags: { test_type: 'load' },
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<5000', 'p(99)<7000'],
        http_req_failed: ['rate<0.1'],
        errors: ['rate<0.1'],
    },
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const headers = {
    'Accept': '*/*',
    'Accept-Language': 'en-GB,en-US;q=0.9,en;q=0.8,hi;q=0.7',
    'Content-Type': 'application/json',
    'Origin': BASE_URL,
    'Referer': `${BASE_URL}/gopdfsuit`,
};

const taggedOpts = getPayloadOptions('tagged');

export default function () {
    const url = `${BASE_URL}/api/v1/generate/template-pdf`;
    const payload = JSON.stringify(getWeightedPayload(taggedOpts));

    const startTime = Date.now();
    const response = http.post(url, payload, { headers: headers });
    const duration = Date.now() - startTime;

    pdfGenerationTime.add(duration);

    const checkResult = check(response, {
        'status is 200': (r) => r.status === 200,
        'response has content': (r) => r.body && r.body.length > 0,
        'response time < 5s': (r) => r.timings.duration < 5000,
        'content-type is correct': (r) => {
            const contentType = r.headers['Content-Type'] || r.headers['content-type'] || '';
            return contentType.includes('application/pdf') || contentType.includes('application/json');
        },
    });

    errorRate.add(!checkResult);
}

export function setup() {
    console.log(`Starting TAGGED / PDF-A load test against ${BASE_URL}`);
    console.log('Scenario: pdfaCompliant=true, taggedPDF=true, sign=true');

    const healthCheck = http.get(`${BASE_URL}/gopdfsuit`);
    if (healthCheck.status !== 200) {
        console.warn(`Warning: Server health check returned status ${healthCheck.status}`);
    }

    return { startTime: new Date().toISOString() };
}

export function teardown(data) {
    console.log(`Tagged load test completed. Started at: ${data.startTime}`);
    console.log(`Finished at: ${new Date().toISOString()}`);
}
