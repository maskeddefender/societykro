import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { authAPI } from '../services/api';

interface User {
  id: string;
  phone: string;
  name: string;
  preferred_language: string;
  is_senior_citizen: boolean;
}

interface Membership {
  id: string;
  society_id: string;
  flat_id: string | null;
  role: string;
  is_primary_member: boolean;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: User | null;
  memberships: Membership[];
  isAuthenticated: boolean;
  isLoading: boolean;

  // Actions
  sendOTP: (phone: string) => Promise<boolean>;
  verifyOTP: (phone: string, otp: string) => Promise<boolean>;
  refreshTokens: () => Promise<boolean>;
  logout: () => void;
  loadStoredAuth: () => Promise<void>;
  getSocietyId: () => string | null;
  getRole: () => string | null;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  refreshToken: null,
  user: null,
  memberships: [],
  isAuthenticated: false,
  isLoading: true,

  sendOTP: async (phone: string) => {
    try {
      await authAPI.post('/auth/otp/send', { phone });
      return true;
    } catch {
      return false;
    }
  },

  verifyOTP: async (phone: string, otp: string) => {
    try {
      const res = await authAPI.post('/auth/otp/verify', { phone, otp });
      const data = res.data.data;

      set({
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        user: data.user,
        memberships: data.memberships || [],
        isAuthenticated: true,
      });

      // Persist tokens
      await Promise.all([
        AsyncStorage.setItem('@auth_access', data.access_token),
        AsyncStorage.setItem('@auth_refresh', data.refresh_token),
        AsyncStorage.setItem('@auth_user', JSON.stringify(data.user)),
        AsyncStorage.setItem('@auth_memberships', JSON.stringify(data.memberships || [])),
      ]);

      return true;
    } catch {
      return false;
    }
  },

  refreshTokens: async () => {
    try {
      const currentRefresh = get().refreshToken;
      if (!currentRefresh) return false;

      const res = await authAPI.post('/auth/refresh', { refresh_token: currentRefresh });
      const data = res.data.data;

      set({
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
      });

      await Promise.all([
        AsyncStorage.setItem('@auth_access', data.access_token),
        AsyncStorage.setItem('@auth_refresh', data.refresh_token),
      ]);

      return true;
    } catch {
      return false;
    }
  },

  logout: () => {
    // Fire and forget server logout
    const token = get().accessToken;
    if (token) {
      authAPI.post('/auth/logout').catch(() => {});
    }

    set({
      accessToken: null,
      refreshToken: null,
      user: null,
      memberships: [],
      isAuthenticated: false,
    });

    Promise.all([
      AsyncStorage.removeItem('@auth_access'),
      AsyncStorage.removeItem('@auth_refresh'),
      AsyncStorage.removeItem('@auth_user'),
      AsyncStorage.removeItem('@auth_memberships'),
    ]);
  },

  loadStoredAuth: async () => {
    try {
      const [access, refresh, user, memberships] = await Promise.all([
        AsyncStorage.getItem('@auth_access'),
        AsyncStorage.getItem('@auth_refresh'),
        AsyncStorage.getItem('@auth_user'),
        AsyncStorage.getItem('@auth_memberships'),
      ]);

      if (access && refresh && user) {
        set({
          accessToken: access,
          refreshToken: refresh,
          user: JSON.parse(user),
          memberships: memberships ? JSON.parse(memberships) : [],
          isAuthenticated: true,
          isLoading: false,
        });
      } else {
        set({ isLoading: false });
      }
    } catch {
      set({ isLoading: false });
    }
  },

  getSocietyId: () => {
    const m = get().memberships;
    return m.length > 0 ? m[0].society_id : null;
  },

  getRole: () => {
    const m = get().memberships;
    return m.length > 0 ? m[0].role : null;
  },
}));
