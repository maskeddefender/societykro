import { test, expect } from '@playwright/test';

const AUTH_URL = 'http://localhost:8081/api/v1';
const COMPLAINT_URL = 'http://localhost:8082/api/v1';

test.describe('Complaint Service', () => {
  let token: string;
  let complaintId: string;

  test.beforeAll(async ({ request }) => {
    const phone = '+919876543210'; // Sharma Ji
    await request.post(`${AUTH_URL}/auth/otp/send`, { data: { phone } });
    const res = await request.post(`${AUTH_URL}/auth/otp/verify`, { data: { phone, otp: '000000' } });
    const body = await res.json();
    token = body.data?.access_token || '';
    if (!token) throw new Error('Failed to get auth token for complaint tests');
  });

  const headers = () => ({ Authorization: `Bearer ${token}` });

  test('GET /health', async ({ request }) => {
    const res = await request.get('http://localhost:8082/health');
    expect(res.status()).toBe(200);
  });

  test('POST /complaints — creates complaint', async ({ request }) => {
    const res = await request.post(`${COMPLAINT_URL}/complaints`, {
      headers: headers(),
      data: {
        category: 'plumbing',
        title: 'Bathroom tap leaking',
        description: 'Hot water tap in master bathroom has been dripping since yesterday',
        priority: 'normal',
        image_urls: [],
      },
    });
    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.ticket_number).toBeTruthy();
    expect(body.data.category).toBe('plumbing');
    expect(body.data.status).toBe('open');
    complaintId = body.data.id;
  });

  test('POST /complaints — rejects missing fields', async ({ request }) => {
    const res = await request.post(`${COMPLAINT_URL}/complaints`, {
      headers: headers(),
      data: { category: 'water' },
    });
    expect(res.status()).toBe(400);
  });

  test('GET /complaints — lists complaints', async ({ request }) => {
    const res = await request.get(`${COMPLAINT_URL}/complaints`, { headers: headers() });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data).toBeInstanceOf(Array);
    expect(body.data.length).toBeGreaterThan(0);
  });

  test('GET /complaints?status=open — filters by status', async ({ request }) => {
    const res = await request.get(`${COMPLAINT_URL}/complaints?status=open`, { headers: headers() });
    expect(res.status()).toBe(200);
    const body = await res.json();
    body.data.forEach((c: any) => expect(c.status).toBe('open'));
  });

  test('GET /complaints/:id — returns detail', async ({ request }) => {
    const res = await request.get(`${COMPLAINT_URL}/complaints/${complaintId}`, { headers: headers() });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data.title).toBe('Bathroom tap leaking');
    expect(body.data.raised_by_name).toBeTruthy();
  });

  test('POST /complaints/:id/comments — adds comment', async ({ request }) => {
    const res = await request.post(`${COMPLAINT_URL}/complaints/${complaintId}/comments`, {
      headers: headers(),
      data: { comment: 'Please send plumber ASAP' },
    });
    expect(res.status()).toBe(201);
  });

  test('GET /complaints/:id/comments — lists comments', async ({ request }) => {
    const res = await request.get(`${COMPLAINT_URL}/complaints/${complaintId}/comments`, { headers: headers() });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data.length).toBeGreaterThan(0);
  });

  test('GET /complaints/analytics — returns stats', async ({ request }) => {
    const res = await request.get(`${COMPLAINT_URL}/complaints/analytics`, { headers: headers() });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.data.counts).toBeTruthy();
    expect(body.data.counts.open).toBeGreaterThanOrEqual(0);
  });

  test('GET /complaints — 401 without auth', async ({ request }) => {
    const res = await request.get(`${COMPLAINT_URL}/complaints`);
    expect(res.status()).toBe(401);
  });
});
