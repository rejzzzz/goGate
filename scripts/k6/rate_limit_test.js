import http from 'k6/http';
import { check, sleep } from 'k6';

// This script bombards the rate-limited endpoint to verify Redis handles the load
// and that 429 Too Many Requests are returned accurately.
export const options = {
    vus: 50,
    duration: '30s',
    thresholds: {
        // We actually EXPECT a lot of failures (429s) because we are intentionally breaching the rate limit
        'http_req_status_429': ['rate>0.50'], 
    },
};

export default function () {
    // Add a randomized IP header to simulate different clients if the gateway rate limits by IP
    // Or just hammer it to hit the route's global/IP limit
    const res = http.get('http://localhost:8080/api/v1/users', {
        headers: {
            'X-Forwarded-For': `192.168.1.${Math.floor(Math.random() * 255)}`
        }
    });
    
    check(res, {
        'is status 200 or 429': (r) => r.status === 200 || r.status === 429,
    });
    sleep(0.05);
}
