import http from 'k6/http';
import { check, sleep } from 'k6';

// This script tests the circuit breaker by targeting the slow/flaky service-b
export const options = {
    vus: 20,
    duration: '1m',
    thresholds: {
        // We expect some 503s as the circuit breaker opens
        'http_req_status_503': ['rate>0.10'], 
    },
};

export default function () {
    // service-b runs on order-service config in some setups, or you can point this to the endpoint 
    // that fails frequently to trigger the breaker.
    const res = http.get('http://localhost:8080/api/v1/orders');
    
    check(res, {
        'is status 200 or 503': (r) => r.status === 200 || r.status === 503,
    });
    sleep(0.1);
}
