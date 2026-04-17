import axios from 'axios';
import { config } from '../constants/config';
import { useGuardStore } from '../store/guardStore';

function createClient(baseURL: string) {
  const client = axios.create({ baseURL, timeout: 10000 });

  client.interceptors.request.use((req) => {
    const token = useGuardStore.getState().accessToken;
    if (token) req.headers.Authorization = `Bearer ${token}`;
    return req;
  });

  client.interceptors.response.use(
    (res) => res,
    (error) => {
      if (error.response?.status === 401) {
        useGuardStore.getState().logout();
      }
      return Promise.reject(error);
    }
  );

  return client;
}

export const authAPI = createClient(config.authURL);
export const visitorAPI = createClient(config.visitorURL);
