'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/authStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function LoginPage() {
  const router = useRouter();
  const { sendOTP, verifyOTP } = useAuthStore();
  const [phone, setPhone] = useState('');
  const [otp, setOtp] = useState('');
  const [step, setStep] = useState<'phone' | 'otp'>('phone');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSendOTP = async () => {
    const formatted = phone.startsWith('+91') ? phone : `+91${phone.replace(/\s/g, '')}`;
    if (formatted.length < 13) {
      setError('Enter a valid 10-digit phone number');
      return;
    }
    setError('');
    setLoading(true);
    const ok = await sendOTP(formatted);
    setLoading(false);
    if (ok) {
      setPhone(formatted);
      setStep('otp');
    } else {
      setError('Failed to send OTP. Try again later.');
    }
  };

  const handleVerifyOTP = async () => {
    if (otp.length !== 6) {
      setError('Enter 6-digit OTP');
      return;
    }
    setError('');
    setLoading(true);
    const ok = await verifyOTP(phone, otp);
    setLoading(false);
    if (ok) {
      router.push('/');
    } else {
      setError('Invalid OTP. Please try again.');
      setOtp('');
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-sky-800 text-white text-xl font-bold">
            SK
          </div>
          <h1 className="text-2xl font-bold text-slate-900">SocietyKro Admin</h1>
          <p className="mt-1 text-sm text-slate-500">Sign in to manage your society</p>
        </div>

        {/* Card */}
        <div className="rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
          {step === 'phone' ? (
            <>
              <Input
                label="Phone Number"
                type="tel"
                placeholder="9876543210"
                value={phone.replace('+91', '')}
                onChange={(e) => setPhone(e.target.value.replace(/[^0-9]/g, ''))}
                maxLength={10}
                autoFocus
              />
              {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
              <Button className="mt-6 w-full" onClick={handleSendOTP} disabled={loading}>
                {loading ? 'Sending...' : 'Send OTP'}
              </Button>
            </>
          ) : (
            <>
              <p className="mb-4 text-sm text-slate-500">OTP sent to <strong>{phone}</strong></p>
              <Input
                label="Enter OTP"
                type="text"
                placeholder="000000"
                value={otp}
                onChange={(e) => setOtp(e.target.value.replace(/[^0-9]/g, ''))}
                maxLength={6}
                autoFocus
                className="text-center text-2xl tracking-[0.5em] font-bold"
              />
              {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
              <Button className="mt-6 w-full" onClick={handleVerifyOTP} disabled={loading}>
                {loading ? 'Verifying...' : 'Verify OTP'}
              </Button>
              <button
                className="mt-4 w-full text-center text-sm text-sky-700 hover:underline"
                onClick={() => { setStep('phone'); setOtp(''); setError(''); }}
              >
                Change phone number
              </button>
            </>
          )}
        </div>

        <p className="mt-6 text-center text-xs text-slate-400">
          Dev bypass: OTP 000000
        </p>
      </div>
    </div>
  );
}
