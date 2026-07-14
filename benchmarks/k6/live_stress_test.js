import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 50, // Run 50 concurrent virtual users
    duration: '30s', // Test for 30 seconds
};

export default function () {
    const params = {
        headers: {
            'X-API-Key': 'sk_test_1133557799', // From your .env
        },
    };
    
    // Testing the Users route
    const res = http.get('https://api.gogate.rejwanul.dev/api/v1/users', params);
    
    // Validate responses
    check(res, {
        'is status 200': (r) => r.status === 200,
        'is status 429 (rate limited)': (r) => r.status === 429, 
        'is status 503 (circuit breaker)': (r) => r.status === 503,
    });
    
    // Slight sleep to simulate real users, but still aggressive enough to hit the 100 req/sec limit
    sleep(0.1); 
}
