import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 100,
    duration: '30s',
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.01'],
    },
};

let authToken;

export function setup() {
    const loginUrl = 'http://localhost:8080/login';
    const credentials = {
        email: 'handayani@gmail.com',
        password: 'handayani',
    };

    const res = http.post(loginUrl, JSON.stringify(credentials), {
        headers: { 'Content-Type': 'application/json' },
    });

    check(res, { 'Login successfully': (r) => r.status === 200 });

    const responseBody = JSON.parse(res.body);
    authToken = responseBody.token;

    if (!authToken) {
        console.error('Failed get token JWT');
    }

    return { authToken: authToken };
}

export default function (data) {
    if (!data.authToken) {
        console.error('Token JWT Not found, melewati pengujian endpoint terproteksi.');
        return;
    }

    const protectedUrl = 'http://localhost:8080/book/get-books';
    const headers = {
        'Authorization': `Bearer ${data.authToken}`,
    };

    const res = http.get(protectedUrl, { headers: headers });

    check(res, { 'Success access endpoint security': (r) => r.status === 200 });
    sleep(1);
}