'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function RootPage() {
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('sk_access_token');
    if (token) {
      router.replace('/');
    } else {
      router.replace('/login');
    }
  }, []);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-sky-800 border-t-transparent" />
    </div>
  );
}
