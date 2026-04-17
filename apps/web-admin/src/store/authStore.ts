import { create } from 'zustand';
import { authAPI } from '../services/api';

interface User {
  id: string;
  phone: string;
  name: string;
  preferred_language: string;
}

interface Membership {
  id: string;
  society_id: string;
  flat_id: string | null;
  role: string;
}

interface AuthState {
  user: User | null;
  memberships: Membership[];
  isAuthenticated: boolean;

  sendOTP: (phone: string) => Promise<boolean>;
  verifyOTP: (phone: string, otp: string) => Promise<boolean>;
  logout: () => void;
  loadAuth: () => void;
  getSocietyId: () => string | null;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  memberships: [],
  isAuthenticated: false,

  sendOTP: async (phone) => {
    try {
      await authAPI.post('/auth/otp/send', { phone });
      return true;
    } catch {
      return false;
    }
  },

  verifyOTP: async (phone, otp) => {
    try {
      const res = await authAPI.post('/auth/otp/verify', { phone, otp });
      const d = res.data.data;

      localStorage.setItem('sk_access_token', d.access_token);
      localStorage.setItem('sk_refresh_token', d.refresh_token);
      localStorage.setItem('sk_user', JSON.stringify(d.user));
      localStorage.setItem('sk_memberships', JSON.stringify(d.memberships || []));

      set({ user: d.user, memberships: d.memberships || [], isAuthenticated: true });
      return true;
    } catch {
      return false;
    }
  },

  logout: () => {
    authAPI.post('/auth/logout').catch(() => {});
    localStorage.clear();
    set({ user: null, memberships: [], isAuthenticated: false });
  },

  loadAuth: () => {
    if (typeof window === 'undefined') return;
    const token = localStorage.getItem('sk_access_token');
    const user = localStorage.getItem('sk_user');
    const memberships = localStorage.getItem('sk_memberships');

    if (token && user) {
      set({
        user: JSON.parse(user),
        memberships: memberships ? JSON.parse(memberships) : [],
        isAuthenticated: true,
      });
    }
  },

  getSocietyId: () => {
    const m = get().memberships;
    return m.length > 0 ? m[0].society_id : null;
  },
}));
