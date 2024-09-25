import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { randomString, randomUrl } from './helpers.js'

const errorRate = new Rate('errors');

const BASE_URL = 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '1m', target: 50 },  // Ramp up to 50 users
    { duration: '3m', target: 50 },  // Stay at 50 users for 3 minutes
    { duration: '1m', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'], // 95% of requests should be below 500ms
    'errors': ['rate<0.1'], // Error rate should be below 10%
  },
};

const createShortUrl = () => {
  const longUrl = randomUrl();
  const response = http.post(
    `${BASE_URL}/create`,
    JSON.stringify({ long_url: longUrl }),
    { headers: { 'Content-Type': 'application/json' },
  });
  
  check(response, {
    'Create status is 201': (r) => r.status === 201,
    'Create response has short_url': (r) => r.json('short_url') !== undefined,
  }) || errorRate.add(1);

  if (response.status === 200) {
    shortUrls.push(response.json('short_url'));
  }

  return response;
};

const getLongUrl = () => {
  const urlHash = randomString(randomIntBetween(1,7));
  const response = http.get(`${BASE_URL}/${urlHash}`);
  
  check(response, {
    'Get status is 200 or 404': (r) => r.status === 200 || r.status === 404,
  }) || errorRate.add(1);

  return response;
};

export default function () {
  // Maintain approximately 1:200 ratio of writes to reads
  if (Math.random() < 0.005) { // 0.5% chance to write
    createShortUrl();
  } else {
    getLongUrl();
  }

  sleep(0.1); // Sleep for 100ms between requests
}