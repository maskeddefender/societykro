import { test, expect } from '@playwright/test';

const AUTH_URL = 'http://localhost:8081/api/v1';
const VENDOR_URL = 'http://localhost:8086/api/v1';
const VOICE_URL = 'http://localhost:8090/api/v1';

test.describe('Vendor Service', () => {
  let adminToken: string;

  test.beforeAll(async ({ request }) => {
    await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone: '+919876543212' } });
    const res = await request.post(`${AUTH_URL}/auth/otp/verify`, {
      data: { phone: '+919876543212', otp: '000000' },
    });
    adminToken = (await res.json()).data.access_token;
  });

  test('GET /health', async ({ request }) => {
    const res = await request.get('http://localhost:8086/health');
    expect(res.status()).toBe(200);
  });

  test('GET /vendors — lists seeded vendors', async ({ request }) => {
    const res = await request.get(`${VENDOR_URL}/vendors`, {
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data.length).toBeGreaterThanOrEqual(3);
  });

  test('POST /vendors — creates vendor (admin)', async ({ request }) => {
    const res = await request.post(`${VENDOR_URL}/vendors`, {
      headers: { Authorization: `Bearer ${adminToken}` },
      data: { name: 'Playwright Test Vendor', phone: '+919876599999', category: 'cleaning' },
    });
    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.name).toBe('Playwright Test Vendor');
  });
});

test.describe('Voice Service', () => {
  test('GET /health', async ({ request }) => {
    const res = await request.get('http://localhost:8090/health');
    expect(res.status()).toBe(200);
  });

  test('GET /languages — returns 22+ languages (public)', async ({ request }) => {
    const res = await request.get(`${VOICE_URL}/languages`);
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data.length).toBeGreaterThanOrEqual(22);

    // Check Hindi exists
    const hindi = body.data.find((l: any) => l.code === 'hi');
    expect(hindi).toBeTruthy();
    expect(hindi.name).toBe('Hindi');
  });
});

test.describe('Message Router', () => {
  test('GET /health', async ({ request }) => {
    const res = await request.get('http://localhost:8089/health');
    expect(res.status()).toBe(200);
  });
});

test.describe('Cross-Service Security', () => {
  test('All services reject requests without auth', async ({ request }) => {
    const endpoints = [
      { url: 'http://localhost:8082/api/v1/complaints', method: 'GET' },
      { url: 'http://localhost:8083/api/v1/visitors', method: 'GET' },
      { url: 'http://localhost:8084/api/v1/payments', method: 'GET' },
      { url: 'http://localhost:8085/api/v1/notices', method: 'GET' },
      { url: 'http://localhost:8086/api/v1/vendors', method: 'GET' },
    ];

    for (const ep of endpoints) {
      const res = await request.get(ep.url);
      expect(res.status(), `${ep.url} should be 401`).toBe(401);
    }
  });
});
