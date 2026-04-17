'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Sidebar } from '@/components/sidebar';
import { useAuthStore } from '@/store/authStore';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { isAuthenticated, loadAuth } = useAuthStore();

  useEffect(() => {
    loadAuth();
  }, []);

  useEffect(() => {
    if (!isAuthenticated && typeof window !== 'undefined') {
      const token = localStorage.getItem('sk_access_token');
      if (!token) router.push('/login');
    }
  }, [isAuthenticated]);

  return (
    <div className="flex h-screen bg-slate-50">
      <Sidebar />
      <main className="flex-1 overflow-y-auto">
        <div className="p-8">
          {children}
        </div>
      </main>
    </div>
  );
}
