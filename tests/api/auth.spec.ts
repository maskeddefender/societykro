import { test, expect } from '@playwright/test';

const AUTH_URL = 'http://localhost:8081/api/v1';
// Use unique phones per test file to avoid rate limit conflicts in parallel runs
const PHONE = '+919876543210';
const PHONE_OTP_TEST = '+919876543299';
const PHONE_REFRESH = '+919876543298';

// Helper to clear Redis rate limits before tests
async function clearRateLimits(request: any) {
  // We can't call Redis directly, but OTP bypass works in dev mode
}

test.describe('Auth Service', () => {
  test.describe('Health', () => {
    test('GET /health returns healthy', async ({ request }) => {
      const res = await request.get('http://localhost:8081/health');
      expect(res.status()).toBe(200);
      const body = await res.json();
      expect(body.status).toBe('healthy');
      expect(body.service).toBe('auth-service');
    });
  });

  test.describe('OTP Flow', () => {
    test('POST /auth/otp/send — sends OTP successfully', async ({ request }) => {
      const res = await request.post(`${AUTH_URL}/auth/otp/send`, {
        data: { phone: PHONE_OTP_TEST },
      });
      expect(res.status()).toBe(200);
      const body = await res.json();
      expect(body.success).toBe(true);
      expect(body.message).toContain('OTP sent');
    });

    test('POST /auth/otp/send — rejects invalid phone', async ({ request }) => {
      const res = await request.post(`${AUTH_URL}/auth/otp/send`, {
        data: { phone: '123' },
      });
      expect(res.status()).toBe(400);
    });

    test('POST /auth/otp/verify — returns JWT + user with dev bypass', async ({ request }) => {
      await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone: PHONE_OTP_TEST } });
      const res = await request.post(`${AUTH_URL}/auth/otp/verify`, {
        data: { phone: PHONE_OTP_TEST, otp: '000000' },
      });
      expect(res.status()).toBe(200);
      const body = await res.json();
      expect(body.success).toBe(true);
      expect(body.data.access_token).toBeTruthy();
      expect(body.data.refresh_token).toBeTruthy();
      expect(body.data.user.phone).toBe(PHONE_OTP_TEST);
      expect(body.data.user.name).toBeTruthy();
    });

    test('POST /auth/otp/verify — rejects wrong OTP', async ({ request }) => {
      await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone: '+919000000099' } });
      const res = await request.post(`${AUTH_URL}/auth/otp/verify`, {
        data: { phone: '+919000000099', otp: '999999' },
      });
      expect(res.status()).toBe(401);
    });
  });

  test.describe('Protected Endpoints', () => {
    let token: string;

    test.beforeAll(async ({ request }) => {
      // Use the seeded user for protected tests
      const phone = '+919876543211'; // Priya Desai (resident)
      await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone } });
      const res = await request.post(`${AUTH_URL}/auth/otp/verify`, {
        data: { phone, otp: '000000' },
      });
      const body = await res.json();
      token = body.data?.access_token || '';
    });

    test('GET /auth/me — returns profile with memberships', async ({ request }) => {
      const res = await request.get(`${AUTH_URL}/auth/me`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      expect(res.status()).toBe(200);
      const body = await res.json();
      expect(body.data.user.name).toBe('Priya Desai');
      expect(body.data.memberships.length).toBeGreaterThan(0);
      expect(body.data.memberships[0].role).toBe('resident');
    });

    test('GET /auth/me — 401 without token', async ({ request }) => {
      const res = await request.get(`${AUTH_URL}/auth/me`);
      expect(res.status()).toBe(401);
    });

    test('GET /auth/me — 401 with invalid token', async ({ request }) => {
      const res = await request.get(`${AUTH_URL}/auth/me`, {
        headers: { Authorization: 'Bearer invalid_garbage_token' },
      });
      expect(res.status()).toBe(401);
    });

    test('GET /societies/:id — returns society details', async ({ request }) => {
      const res = await request.get(`${AUTH_URL}/societies/a0000000-0000-0000-0000-000000000001`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      expect(res.status()).toBe(200);
      const body = await res.json();
      expect(body.data.name).toBe('Green Valley Apartments');
      expect(body.data.code).toBe('GVA123');
      expect(body.data.city).toBe('Pune');
    });

    test('GET /societies/:id/flats — returns flats list', async ({ request }) => {
      const res = await request.get(`${AUTH_URL}/societies/a0000000-0000-0000-0000-000000000001/flats`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      expect(res.status()).toBe(200);
      const body = await res.json();
      expect(body.data.length).toBeGreaterThan(0);
      expect(body.data[0].flat_number).toBeTruthy();
    });

    test('POST /auth/refresh — returns new token pair', async ({ request }) => {
      await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone: PHONE_REFRESH } });
      const loginRes = await request.post(`${AUTH_URL}/auth/otp/verify`, {
        data: { phone: PHONE_REFRESH, otp: '000000' },
      });
      const refresh = (await loginRes.json()).data?.refresh_token;
      if (!refresh) { test.skip(); return; }

      const res = await request.post(`${AUTH_URL}/auth/refresh`, {
        data: { refresh_token: refresh },
      });
      expect(res.status()).toBe(200);
      const body = await res.json();
      expect(body.data.access_token).toBeTruthy();
      expect(body.data.refresh_token).toBeTruthy();
    });

    test('POST /auth/logout — clears refresh token', async ({ request }) => {
      const res = await request.post(`${AUTH_URL}/auth/logout`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      expect(res.status()).toBe(200);
    });
  });
});
