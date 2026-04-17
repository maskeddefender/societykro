import { useEffect } from 'react';
import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useGuardStore } from '../store/guardStore';

export default function RootLayout() {
  const loadAuth = useGuardStore((s) => s.loadAuth);
  useEffect(() => { loadAuth(); }, []);

  return (
    <>
      <StatusBar style="light" />
      <Stack screenOptions={{ headerShown: false }} />
    </>
  );
}
