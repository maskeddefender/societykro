import { Platform } from 'react-native';
import Constants from 'expo-constants';

// Dev host resolution:
// - Android emulator: 10.0.2.2 maps to host machine
// - iOS simulator: localhost works
// - Physical device: use machine's LAN IP
// Override via Expo env: EXPO_PUBLIC_API_HOST=192.168.1.100
function getDevHost(): string {
  // Allow override via Expo environment variable
  const override = Constants.expoConfig?.extra?.apiHost
    || process.env.EXPO_PUBLIC_API_HOST;
  if (override) return override;

  // Default per platform
  return Platform.OS === 'android' ? '10.0.2.2' : 'localhost';
}

const DEV_HOST = getDevHost();
const PROD_BASE = 'https://api.societykro.in/api/v1';

function devURL(port: number) {
  return `http://${DEV_HOST}:${port}/api/v1`;
}

export const config = {
  api: {
    authBaseURL:      __DEV__ ? devURL(8081) : PROD_BASE,
    complaintBaseURL: __DEV__ ? devURL(8082) : PROD_BASE,
    visitorBaseURL:   __DEV__ ? devURL(8083) : PROD_BASE,
    paymentBaseURL:   __DEV__ ? devURL(8084) : PROD_BASE,
    noticeBaseURL:    __DEV__ ? devURL(8085) : PROD_BASE,
    vendorBaseURL:    __DEV__ ? devURL(8086) : PROD_BASE,
    voiceBaseURL:     __DEV__ ? devURL(8090) : PROD_BASE,
  },
  otp: {
    length: 6,
    devBypass: '000000',
  },
  pagination: {
    defaultLimit: 20,
  },
} as const;
