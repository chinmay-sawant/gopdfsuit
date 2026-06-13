import http from 'k6/http';
import { check } from 'k6';
import { generateRetailHTML } from './html_payload_generator.js';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:3010';
const SKIP_NETWORK_IDLE = __ENV.SKIP_NETWORK_IDLE !== '0';

export const options = {
    vus: 1,
    duration: '10s',
    thresholds: {
        http_req_failed: ['rate<0.05'],
    },
};

export default function () {
    const formData = {
        files: http.file(generateRetailHTML(), 'index.html', 'text/html'),
        printBackground: 'true',
    };
    if (SKIP_NETWORK_IDLE) {
        formData.skipNetworkIdleEvent = 'true';
    }

    const res = http.post(`${BASE_URL}/forms/chromium/convert/html`, formData);
    check(res, {
        'status is 200': (r) => r.status === 200,
        'pdf bytes': (r) => r.body && r.body.length > 1000,
    });
}