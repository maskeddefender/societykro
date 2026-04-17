import { test, expect } from '@playwright/test';

const AUTH_URL = 'http://localhost:8081/api/v1';
const PAYMENT_URL = 'http://localhost:8084/api/v1';

test.describe('Payment Service', () => {
  let residentToken: string;
  let adminToken: string;

  test.beforeAll(async ({ request }) => {
    // Resident token
    await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone: '+919876543210' } });
    const r1 = await request.post(`${AUTH_URL}/auth/otp/verify`, {
      data: { phone: '+919876543210', otp: '000000' },
    });
    residentToken = (await r1.json()).data.access_token;

    // Admin token (secretary)
    await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone: '+919876543212' } });
    const r2 = await request.post(`${AUTH_URL}/auth/otp/verify`, {
      data: { phone: '+919876543212', otp: '000000' },
    });
    adminToken = (await r2.json()).data.access_token;
  });

  test('GET /health', async ({ request }) => {
    const res = await request.get('http://localhost:8084/health');
    expect(res.status()).toBe(200);
  });

  test('POST /payments/generate-bills — admin generates bills', async ({ request }) => {
    const res = await request.post(`${PAYMENT_URL}/payments/generate-bills`, {
      headers: { Authorization: `Bearer ${adminToken}` },
      data: { month: '2026-05' },
    });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data.bills_generated).toBeGreaterThanOrEqual(0);
  });

  test('POST /payments/generate-bills — 403 for resident', async ({ request }) => {
    const res = await request.post(`${PAYMENT_URL}/payments/generate-bills`, {
      headers: { Authorization: `Bearer ${residentToken}` },
      data: { month: '2026-05' },
    });
    expect(res.status()).toBe(403);
  });

  test('GET /payments — lists payments', async ({ request }) => {
    const res = await request.get(`${PAYMENT_URL}/payments`, {
      headers: { Authorization: `Bearer ${residentToken}` },
    });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data).toBeInstanceOf(Array);
  });

  test('GET /payments — 401 without auth', async ({ request }) => {
    const res = await request.get(`${PAYMENT_URL}/payments`);
    expect(res.status()).toBe(401);
  });
});
