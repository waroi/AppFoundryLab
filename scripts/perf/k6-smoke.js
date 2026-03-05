import http from 'k6/http';
import { check, sleep } from 'k6';

const baseUrl = __ENV.K6_BASE_URL || 'http://127.0.0.1:8080';
const username = __ENV.K6_USERNAME || 'developer';
const password = __ENV.K6_PASSWORD || 'developer_dev_password';

export const options = {
  vus: Number(__ENV.K6_VUS || 8),
  duration: __ENV.K6_DURATION || '30s',
  thresholds: {
    http_req_failed: ['rate<0.02'],
    http_req_duration: ['p(95)<800'],
    checks: ['rate>0.99'],
  },
};

function getToken() {
  const payload = JSON.stringify({ username, password });
  const res = http.post(`${baseUrl}/api/v1/auth/token`, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '5s',
  });

  check(res, {
    'auth status is 200': (r) => r.status === 200,
    'auth has access token': (r) => {
      try {
        return !!r.json('accessToken');
      } catch (_) {
        return false;
      }
    },
  });

  try {
    return res.json('accessToken') || '';
  } catch (_) {
    return '';
  }
}

export default function () {
  const token = getToken();
  const headers = {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  const usersRes = http.get(`${baseUrl}/api/v1/users`, {
    headers,
    timeout: '5s',
  });
  check(usersRes, {
    'users status is 200': (r) => r.status === 200,
  });

  const healthRes = http.get(`${baseUrl}/health/ready`, { timeout: '5s' });
  check(healthRes, {
    'ready status is 200 or 503': (r) => r.status === 200 || r.status === 503,
  });

  sleep(0.2);
}
