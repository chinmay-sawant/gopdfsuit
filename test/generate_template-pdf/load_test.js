import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const pdfGenerationTime = new Trend('pdf_generation_time');

// Load JSON payload from file (executed once at init time)
const amazonReceiptPayload = JSON.parse(open('../../sampledata/amazon/amazon_receipt.json'));

// Test configuration
export const options = {
    scenarios: {
        // Smoke test - verify system works under minimal load
        smoke: {
            executor: 'constant-vus',
            vus: 1,
            duration: '30s',
            startTime: '0s',
            tags: { test_type: 'smoke' },
        },
        // Load test - normal expected load
        load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '1m', target: 10 },  // Ramp up to 10 users
                { duration: '3m', target: 10 },  // Stay at 10 users
                { duration: '1m', target: 0 },   // Ramp down to 0
            ],
            startTime: '35s',
            tags: { test_type: 'load' },
        },
        // Stress test - beyond normal load
        stress: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '1m', target: 20 },  // Ramp up to 20 users
                { duration: '2m', target: 20 },  // Stay at 20 users
                { duration: '1m', target: 50 },  // Ramp up to 50 users
                { duration: '2m', target: 50 },  // Stay at 50 users
                { duration: '1m', target: 0 },   // Ramp down
            ],
            startTime: '6m',
            tags: { test_type: 'stress' },
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<5000'],  // 95% of requests should be below 5s
        http_req_failed: ['rate<0.1'],       // Error rate should be below 10%
        errors: ['rate<0.1'],                // Custom error rate below 10%
    },
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
    const payload = JSON.stringify(amazonReceiptPayload);

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
    sleep(1);
}

// Setup function - runs once before the test
export function setup() {
    console.log(`Starting load test against ${BASE_URL}`);
    console.log('Testing endpoint: /api/v1/generate/template-pdf');
    console.log('Payload loaded from: sampledata/amazon/amazon_receipt.json');
    
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
