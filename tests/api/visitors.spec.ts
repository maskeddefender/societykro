import { test, expect } from '@playwright/test';

const AUTH_URL = 'http://localhost:8081/api/v1';
const VISITOR_URL = 'http://localhost:8083/api/v1';

test.describe('Visitor Service', () => {
  let token: string;

  test.beforeAll(async ({ request }) => {
    // Use Priya Desai to avoid rate limit clash with complaint tests using Sharma Ji
    const phone = '+919876543211';
    await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone } });
    const res = await request.post(`${AUTH_URL}/auth/otp/verify`, { data: { phone, otp: '000000' } });
    const body = await res.json();
    token = body.data?.access_token || '';
    if (!token) throw new Error('Failed to get auth token for visitor tests');
  });

  const headers = () => ({ Authorization: `Bearer ${token}` });

  test('GET /health', async ({ request }) => {
    const res = await request.get('http://localhost:8083/health');
    expect(res.status()).toBe(200);
  });

  test('POST /visitors/pre-approve — generates OTP', async ({ request }) => {
    const res = await request.post(`${VISITOR_URL}/visitors/pre-approve`, {
      headers: headers(),
      data: { name: 'Test Visitor', purpose: 'guest', phone: '+919111222333' },
    });
    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.otp).toBeTruthy();
    expect(body.data.otp.length).toBe(6);
    expect(body.data.visitor.name).toBe('Test Visitor');
    expect(body.data.visitor.status).toBe('pending');
  });

  test('POST /visitors/pre-approve — rejects missing name', async ({ request }) => {
    const res = await request.post(`${VISITOR_URL}/visitors/pre-approve`, {
      headers: headers(),
      data: { purpose: 'guest' },
    });
    expect(res.status()).toBe(400);
  });

  test('GET /visitors — lists visitors', async ({ request }) => {
    const res = await request.get(`${VISITOR_URL}/visitors`, { headers: headers() });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data).toBeInstanceOf(Array);
  });

  test('GET /visitors/active — lists checked-in visitors', async ({ request }) => {
    const res = await request.get(`${VISITOR_URL}/visitors/active`, { headers: headers() });
    expect(res.status()).toBe(200);
  });

  test('GET /visitors — 401 without auth', async ({ request }) => {
    const res = await request.get(`${VISITOR_URL}/visitors`);
    expect(res.status()).toBe(401);
  });
});
