import http from 'k6/http';
import { check, sleep } from 'k6';

// This benchmark pushes the gateway to its limits to test max RPS and P99 latency.
export const options = {
    stages: [
        { duration: '30s', target: 200 },  // Ramp up
        { duration: '2m', target: 1000 },  // Sustained high load
        { duration: '30s', target: 0 },    // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(99)<15'], // 99% of requests must complete below 15ms
        http_req_failed: ['rate<0.01'],  // Less than 1% error rate
    },
};

export default function () {
    // Calling the fast user-service via the gateway
    const res = http.get('http://localhost:8080/api/v1/users');
    check(res, {
        'is status 200': (r) => r.status === 200,
    });
    // Tiny sleep to avoid completely overwhelming the local port pool during test
    sleep(0.01);
}
