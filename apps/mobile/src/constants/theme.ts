export const colors = {
  primary: '#1A3C5E',
  primaryLight: '#2E75B6',
  primaryDark: '#0F2640',
  secondary: '#2E6DA4',
  accent: '#4A86C8',

  success: '#22C55E',
  warning: '#F59E0B',
  error: '#EF4444',
  info: '#3B82F6',

  background: '#F8FAFC',
  surface: '#FFFFFF',
  surfaceSecondary: '#F1F5F9',
  border: '#E2E8F0',
  borderLight: '#F1F5F9',

  text: '#0F172A',
  textSecondary: '#64748B',
  textTertiary: '#94A3B8',
  textInverse: '#FFFFFF',

  statusOpen: '#F59E0B',
  statusInProgress: '#3B82F6',
  statusResolved: '#22C55E',
  statusClosed: '#94A3B8',
  statusEmergency: '#EF4444',
} as const;

export const spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  xxl: 24,
  xxxl: 32,
} as const;

export const fontSize = {
  xs: 11,
  sm: 13,
  md: 15,
  lg: 17,
  xl: 20,
  xxl: 24,
  xxxl: 32,
  title: 28,
} as const;

export const borderRadius = {
  sm: 6,
  md: 10,
  lg: 14,
  xl: 20,
  full: 9999,
} as const;
