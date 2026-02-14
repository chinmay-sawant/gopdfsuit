import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const pdfGenerationTime = new Trend('pdf_generation_time');
const successfulRequests = new Counter('successful_requests');
const failedRequests = new Counter('failed_requests');

import { getWeightedPayload } from './payload_generator.js';

// Load JSON payload from file (executed once at init time)
// const financialDigitalSignaturePayload = JSON.parse(open('../../sampledata/editor/financial_digitalsignature.json'));

// Spike test configuration - sudden increase in load
export const options = {
    scenarios: {
        spike: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '10s', target: 5 },    // Warm up
                { duration: '10s', target: 5 },    // Stay at normal load
                { duration: '10s', target: 100 },  // Spike to 100 users
                { duration: '30s', target: 100 },  // Stay at spike
                { duration: '10s', target: 5 },    // Scale down quickly
                { duration: '20s', target: 5 },    // Recovery period
                { duration: '10s', target: 0 },    // Ramp down to 0
            ],
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<10000', 'p(99)<15000'],  // More lenient during spike
        http_req_failed: ['rate<0.3'],        // Allow higher error rate during spike
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
    });

    if (checkResult) {
        successfulRequests.add(1);
    } else {
        failedRequests.add(1);
    }
    
    errorRate.add(!checkResult);

    sleep(0.5);
}

export function handleSummary(data) {
    const summary = `
========== SPIKE TEST SUMMARY ==========
Total requests: ${data.metrics.http_reqs.values.count}
Throughput: ${(data.metrics.http_reqs.values.count / data.state.testRunDurationMs * 1000).toFixed(2)} req/s
Failed requests: ${data.metrics.http_req_failed.values.passes}
Average response time: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms
95th percentile: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms
99th percentile: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms
Max response time: ${data.metrics.http_req_duration.values.max.toFixed(2)}ms
==========================================
`;
    console.log(summary);
    
    return {
        'stdout': summary,
        'current_test_results.txt': summary,
        'summary.json': JSON.stringify(data),
    };
}
