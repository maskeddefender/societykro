import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { config } from '../constants/config';
import { useAuthStore } from '../store/authStore';

// Create separate axios instances per service
function createClient(baseURL: string) {
  const client = axios.create({
    baseURL,
    timeout: 15000,
    headers: { 'Content-Type': 'application/json' },
  });

  // Inject JWT token on every request
  client.interceptors.request.use((req: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().accessToken;
    if (token) {
      req.headers.Authorization = `Bearer ${token}`;
    }
    return req;
  });

  // Handle 401 → attempt token refresh → retry
  client.interceptors.response.use(
    (res) => res,
    async (error: AxiosError) => {
      const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

      if (error.response?.status === 401 && !originalRequest._retry) {
        originalRequest._retry = true;

        const refreshed = await useAuthStore.getState().refreshTokens();
        if (refreshed) {
          const newToken = useAuthStore.getState().accessToken;
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return client(originalRequest);
        }

        // Refresh failed — force logout
        useAuthStore.getState().logout();
      }

      return Promise.reject(error);
    }
  );

  return client;
}

export const authAPI = createClient(config.api.authBaseURL);
export const complaintAPI = createClient(config.api.complaintBaseURL);
export const visitorAPI = createClient(config.api.visitorBaseURL);
export const paymentAPI = createClient(config.api.paymentBaseURL);
export const noticeAPI = createClient(config.api.noticeBaseURL);
