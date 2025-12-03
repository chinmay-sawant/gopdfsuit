import http from 'k6/http';
import { check, sleep } from 'k6';

// Load JSON payload from file (executed once at init time)
const amazonReceiptPayload = JSON.parse(open('../sampledata/amazon/amazon_receipt.json'));

// Simple smoke test configuration
export const options = {
    vus: 1,
    duration: '10s',
    thresholds: {
        http_req_duration: ['p(95)<3000'],
        http_req_failed: ['rate<0.05'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const headers = {
    'Content-Type': 'application/json',
    'Accept': '*/*',
};

export default function () {
    const url = `${BASE_URL}/api/v1/generate/template-pdf`;
    const response = http.post(url, JSON.stringify(amazonReceiptPayload), { headers });

    check(response, {
        'status is 200': (r) => r.status === 200,
        'response has content': (r) => r.body && r.body.length > 0,
    });

    sleep(0.5);
}
