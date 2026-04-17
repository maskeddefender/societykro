import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { authAPI } from '../services/api';

interface GuardState {
  accessToken: string | null;
  guardName: string | null;
  societyId: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  sendOTP: (phone: string) => Promise<boolean>;
  verifyOTP: (phone: string, otp: string) => Promise<boolean>;
  logout: () => void;
  loadAuth: () => Promise<void>;
}

export const useGuardStore = create<GuardState>((set) => ({
  accessToken: null,
  guardName: null,
  societyId: null,
  isAuthenticated: false,
  isLoading: true,

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
      const membership = d.memberships?.[0];

      // Guard app only allows guard role
      if (membership?.role !== 'guard' && membership?.role !== 'admin' && membership?.role !== 'secretary') {
        return false;
      }

      await Promise.all([
        AsyncStorage.setItem('@guard_token', d.access_token),
        AsyncStorage.setItem('@guard_name', d.user.name),
        AsyncStorage.setItem('@guard_society', membership?.society_id || ''),
      ]);

      set({
        accessToken: d.access_token,
        guardName: d.user.name,
        societyId: membership?.society_id || null,
        isAuthenticated: true,
      });
      return true;
    } catch {
      return false;
    }
  },

  logout: () => {
    set({ accessToken: null, guardName: null, societyId: null, isAuthenticated: false });
    Promise.all([
      AsyncStorage.removeItem('@guard_token'),
      AsyncStorage.removeItem('@guard_name'),
      AsyncStorage.removeItem('@guard_society'),
    ]);
  },

  loadAuth: async () => {
    try {
      const [token, name, society] = await Promise.all([
        AsyncStorage.getItem('@guard_token'),
        AsyncStorage.getItem('@guard_name'),
        AsyncStorage.getItem('@guard_society'),
      ]);
      if (token && name) {
        set({ accessToken: token, guardName: name, societyId: society, isAuthenticated: true, isLoading: false });
      } else {
        set({ isLoading: false });
      }
    } catch {
      set({ isLoading: false });
    }
  },
}));
