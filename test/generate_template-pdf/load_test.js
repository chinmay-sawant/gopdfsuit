import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const pdfGenerationTime = new Trend('pdf_generation_time');

import { getWeightedPayload } from './payload_generator.js';

// Load JSON payload from file (executed once at init time)
// const financialDigitalSignaturePayload = JSON.parse(open('../../sampledata/editor/financial_digitalsignature.json'));

// Test configuration
export const options = {
    scenarios: {
        // Smoke test - verify system works under minimal load
        smoke: {
            executor: 'constant-vus',
            vus: 1,
            duration: '5s',
            startTime: '0s',
            tags: { test_type: 'smoke' },
        },
        // Load test - normal expected load
        load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '5s', target: 48 },   // Ramp up to 48 users
                { duration: '10s', target: 48 },  // Stay at 48 users
                { duration: '5s', target: 0 },     // Ramp down to 0
            ],
            startTime: '10s',
            tags: { test_type: 'load' },
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<5000', 'p(99)<7000'],  // 95% < 5s, 99% < 7s
        http_req_failed: ['rate<0.1'],       // Error rate should be below 10%
        errors: ['rate<0.1'],                // Custom error rate below 10%
    },
    summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
};

// Base URL configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Headers for the request
const headers = {
    'Accept': '*/*',
    'Accept-Language': 'en-GB,en-US;q=0.9,en;q=0.8,hi;q=0.7',
    'Content-Type': 'application/json',
    'Origin': BASE_URL,
    'Referer': `${BASE_URL}/gopdfsuit`,
};

export default function () {
    const url = `${BASE_URL}/api/v1/generate/template-pdf`;
    // Generate dynamic payload based on weighted distribution
    const payload = JSON.stringify(getWeightedPayload());

    const startTime = Date.now();
    const response = http.post(url, payload, { headers: headers });
    const duration = Date.now() - startTime;

    // Record custom metrics
    pdfGenerationTime.add(duration);

    // Check response
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

    // Small sleep between requests to simulate realistic user behavior
    // sleep(1);
}

// Setup function - runs once before the test
export function setup() {
    console.log(`Starting load test against ${BASE_URL}`);
    console.log('Testing endpoint: /api/v1/generate/template-pdf');
    console.log('Payload loaded from: random payload');

    // Verify server is reachable
    const healthCheck = http.get(`${BASE_URL}/gopdfsuit`);
    if (healthCheck.status !== 200) {
        console.warn(`Warning: Server health check returned status ${healthCheck.status}`);
    }

    return { startTime: new Date().toISOString() };
}

// Teardown function - runs once after the test
export function teardown(data) {
    console.log(`Load test completed. Started at: ${data.startTime}`);
    console.log(`Finished at: ${new Date().toISOString()}`);
}
