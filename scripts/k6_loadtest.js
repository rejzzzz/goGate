import http from 'k6/http';
import { check } from 'k6';

// 1. Read .env file to get the bypass token (k6 reads this from disk at startup)
let bypassToken = '';
let bypassHeader = 'X-Stress-Test-Token';
let apiKey = '';

try {
    const envFile = open('../.env');
    envFile.split('\n').forEach((line) => {
        if (line.startsWith('STRESS_TEST_BYPASS_TOKEN=')) {
            bypassToken = line.split('=')[1].trim().replace(/['"]/g, '');
        }
        if (line.startsWith('STRESS_TEST_BYPASS_HEADER=')) {
            bypassHeader = line.split('=')[1].trim().replace(/['"]/g, '');
        }
        if (line.startsWith('TEST_API_KEY=')) {
            apiKey = line.split('=')[1].trim().replace(/['"]/g, '');
        }
    });
} catch (e) {
    console.log("Could not open ../.env, continuing without bypass token...");
}

export const options = {
    // Start with 1000, ramp up to 3000 over 10s, hold for 20s
    stages: [
        { duration: '10s', target: 3000 },
        { duration: '20s', target: 3000 },
    ],
};

export default function () {
    const url = __ENV.TARGET_URL || 'https://api.gogate.rejwanul.dev/api/v1/users';
    
    const params = {
        headers: {},
    };

    if (bypassToken) {
        params.headers[bypassHeader] = bypassToken;
    }
    if (apiKey) {
        params.headers['X-API-Key'] = apiKey;
    }

    const res = http.get(url, params);
    
    // Check if the request was successful
    const success = check(res, {
        'status is 200': (r) => r.status === 200,
    });
    
    if (!success) {
        console.log(`Failed! Status: ${res.status}, Body: ${res.body.substring(0, 100)}`);
    }
}
