import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const pdfGenerationTime = new Trend('pdf_generation_time');
const successfulRequests = new Counter('successful_requests');

import { getWeightedPayload } from './payload_generator.js';

// Load JSON payload from file (executed once at init time)
// const financialDigitalSignaturePayload = JSON.parse(open('../../sampledata/editor/financial_digitalsignature.json'));

// Soak test configuration - sustained load over extended period
export const options = {
    scenarios: {
        soak: {
            executor: 'constant-vus',
            vus: 10,
            duration: '2m',  // Reduced for reporting purposes
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<5000', 'p(99)<8000'],
        http_req_failed: ['rate<0.05'],
        errors: ['rate<0.05'],
    },
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const headers = {
    'Content-Type': 'application/json',
    'Accept': '*/*',
};

export default function () {
    const url = `${BASE_URL}/api/v1/generate/template-pdf`;
    const payload = JSON.stringify(getWeightedPayload());

    const startTime = Date.now();
    const response = http.post(url, payload, { headers });
    const duration = Date.now() - startTime;

    pdfGenerationTime.add(duration);

    const checkResult = check(response, {
        'status is 200': (r) => r.status === 200,
        'response has content': (r) => r.body && r.body.length > 0,
        'response time < 5s': (r) => r.timings.duration < 5000,
    });

    if (checkResult) {
        successfulRequests.add(1);
    }
    
    errorRate.add(!checkResult);

    // Slightly longer sleep for soak test
    sleep(2);
}

export function setup() {
    console.log('Starting soak test - this will run for 30 minutes');
    console.log(`Target: ${BASE_URL}/api/v1/generate/template-pdf`);
    console.log('Payload loaded from: random payload');
    return { startTime: new Date().toISOString() };
}

export function teardown(data) {
    console.log(`Soak test completed. Duration: ${data.startTime} to ${new Date().toISOString()}`);
}
