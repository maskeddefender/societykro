import { test, expect } from '@playwright/test';

const AUTH_URL = 'http://localhost:8081/api/v1';
const NOTICE_URL = 'http://localhost:8085/api/v1';

test.describe('Notice Service', () => {
  let token: string;
  let noticeId: string;

  test.beforeAll(async ({ request }) => {
    await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone: '+919876543210' } });
    const res = await request.post(`${AUTH_URL}/auth/otp/verify`, {
      data: { phone: '+919876543210', otp: '000000' },
    });
    token = (await res.json()).data.access_token;
  });

  const headers = () => ({ Authorization: `Bearer ${token}` });

  test('GET /health', async ({ request }) => {
    const res = await request.get('http://localhost:8085/health');
    expect(res.status()).toBe(200);
  });

  test('POST /notices — creates notice', async ({ request }) => {
    const res = await request.post(`${NOTICE_URL}/notices`, {
      headers: headers(),
      data: {
        title: 'Test Notice from Playwright',
        body: 'This is an automated test notice.',
        category: 'general',
        is_pinned: false,
      },
    });
    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.title).toBe('Test Notice from Playwright');
    noticeId = body.data.id;
  });

  test('GET /notices — lists notices', async ({ request }) => {
    const res = await request.get(`${NOTICE_URL}/notices`, { headers: headers() });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data).toBeInstanceOf(Array);
    expect(body.data.length).toBeGreaterThan(0);
  });

  test('GET /notices/:id — returns detail + auto marks read', async ({ request }) => {
    const res = await request.get(`${NOTICE_URL}/notices/${noticeId}`, { headers: headers() });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data.title).toBe('Test Notice from Playwright');
    expect(body.data.read_count).toBeGreaterThanOrEqual(1);
  });

  test('POST /notices/:id/read — marks as read', async ({ request }) => {
    const res = await request.post(`${NOTICE_URL}/notices/${noticeId}/read`, { headers: headers() });
    expect(res.status()).toBe(200);
  });
});
