import http from 'k6/http';
import { check, sleep } from 'k6';

const baseUrl = __ENV.K6_BASE_URL || 'http://127.0.0.1:8080';
const username = __ENV.K6_USERNAME || 'developer';
const password = __ENV.K6_PASSWORD || 'developer_dev_password';

export const options = {
  stages: [
    { duration: __ENV.K6_SPIKE_RAMP_UP || '20s', target: Number(__ENV.K6_SPIKE_PEAK_VUS || 40) },
    { duration: __ENV.K6_SPIKE_HOLD || '20s', target: Number(__ENV.K6_SPIKE_PEAK_VUS || 40) },
    { duration: __ENV.K6_SPIKE_RAMP_DOWN || '20s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.03'],
    http_req_duration: ['p(95)<1200'],
    checks: ['rate>0.98'],
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

  const fibRes = http.post(
    `${baseUrl}/api/v1/compute/fibonacci`,
    JSON.stringify({ n: 32 }),
    { headers, timeout: '5s' },
  );
  check(fibRes, {
    'fibonacci status is 200 or 503': (r) => r.status === 200 || r.status === 503,
  });

  const usersRes = http.get(`${baseUrl}/api/v1/users`, {
    headers,
    timeout: '5s',
  });
  check(usersRes, {
    'users status is 200 or 429': (r) => r.status === 200 || r.status === 429,
  });

  sleep(0.1);
}
