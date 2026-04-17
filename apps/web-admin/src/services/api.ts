import axios from 'axios';

const DEV_BASE = 'http://localhost';

function createClient(port: number) {
  const client = axios.create({
    baseURL: process.env.NODE_ENV === 'development' ? `${DEV_BASE}:${port}/api/v1` : '/api/v1',
    timeout: 15000,
    headers: { 'Content-Type': 'application/json' },
  });

  client.interceptors.request.use((config) => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('sk_access_token');
      if (token) config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  client.interceptors.response.use(
    (res) => res,
    async (error) => {
      if (error.response?.status === 401 && typeof window !== 'undefined') {
        const refresh = localStorage.getItem('sk_refresh_token');
        if (refresh && !error.config._retry) {
          error.config._retry = true;
          try {
            const res = await authAPI.post('/auth/refresh', { refresh_token: refresh });
            const { access_token, refresh_token } = res.data.data;
            localStorage.setItem('sk_access_token', access_token);
            localStorage.setItem('sk_refresh_token', refresh_token);
            error.config.headers.Authorization = `Bearer ${access_token}`;
            return client(error.config);
          } catch {
            localStorage.clear();
            window.location.href = '/login';
          }
        }
      }
      return Promise.reject(error);
    }
  );

  return client;
}

export const authAPI = createClient(8081);
export const complaintAPI = createClient(8082);
export const visitorAPI = createClient(8083);
export const paymentAPI = createClient(8084);
export const noticeAPI = createClient(8085);
export const vendorAPI = createClient(8086);
