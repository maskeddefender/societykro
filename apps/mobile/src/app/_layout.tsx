import { useEffect } from 'react';
import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAuthStore } from '../store/authStore';
import '../i18n';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 2, staleTime: 30_000 },
  },
});

export default function RootLayout() {
  const loadStoredAuth = useAuthStore((s) => s.loadStoredAuth);

  useEffect(() => {
    loadStoredAuth();
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <StatusBar style="light" />
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="(auth)" />
        <Stack.Screen name="(tabs)" />
        <Stack.Screen name="complaint/[id]" options={{ headerShown: true, title: 'Complaint Detail' }} />
        <Stack.Screen name="complaint/new" options={{ headerShown: true, title: 'Raise Complaint', presentation: 'modal' }} />
        <Stack.Screen name="visitor/[id]" options={{ headerShown: true, title: 'Visitor Detail' }} />
        <Stack.Screen name="notice/[id]" options={{ headerShown: true, title: 'Notice' }} />
      </Stack>
    </QueryClientProvider>
  );
}
