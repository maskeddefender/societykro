import { Platform } from 'react-native';

const DEV_HOST = Platform.OS === 'android' ? '10.0.2.2' : 'localhost';

export const config = {
  authURL: __DEV__ ? `http://${DEV_HOST}:8081/api/v1` : 'https://api.societykro.in/api/v1',
  visitorURL: __DEV__ ? `http://${DEV_HOST}:8083/api/v1` : 'https://api.societykro.in/api/v1',
} as const;
