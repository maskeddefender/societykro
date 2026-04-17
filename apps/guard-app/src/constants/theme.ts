// Guard app uses a dark theme for outdoor/night visibility
export const colors = {
  primary: '#0F172A',     // Dark navy
  primaryLight: '#1E3A5F',
  accent: '#3B82F6',      // Blue for actions
  success: '#22C55E',
  danger: '#EF4444',
  warning: '#F59E0B',

  background: '#0F172A',
  surface: '#1E293B',
  surfaceLight: '#334155',
  border: '#475569',

  text: '#F8FAFC',
  textSecondary: '#94A3B8',
  textMuted: '#64748B',
} as const;

export const spacing = { xs: 4, sm: 8, md: 12, lg: 16, xl: 24, xxl: 32 } as const;
export const fontSize = { sm: 14, md: 16, lg: 20, xl: 24, xxl: 32, hero: 48 } as const;
export const borderRadius = { sm: 8, md: 12, lg: 16, xl: 24 } as const;
