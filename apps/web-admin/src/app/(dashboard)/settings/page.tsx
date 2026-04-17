'use client';

import { useEffect, useState } from 'react';
import { authAPI } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { formatCurrency } from '@/lib/utils';

interface Society {
  id: string; name: string; code: string; address: string; city: string;
  total_flats: number; maintenance_amount: number; maintenance_due_day: number;
  late_fee_percent: number; subscription: string;
}

export default function SettingsPage() {
  const { getSocietyId } = useAuthStore();
  const [society, setSociety] = useState<Society | null>(null);
  const [loading, setLoading] = useState(true);
  const [form, setForm] = useState({ maintenance_amount: 0, maintenance_due_day: 1, late_fee_percent: 0 });
  const [saving, setSaving] = useState(false);
  const [saveMsg, setSaveMsg] = useState('');

  useEffect(() => {
    const societyId = getSocietyId();
    if (!societyId) { setLoading(false); return; }
    authAPI.get(`/societies/${societyId}`)
      .then((res) => {
        const s = res.data.data;
        setSociety(s);
        setForm({
          maintenance_amount: s.maintenance_amount || 0,
          maintenance_due_day: s.maintenance_due_day || 1,
          late_fee_percent: s.late_fee_percent || 0,
        });
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [getSocietyId]);

  const handleSave = () => {
    const societyId = getSocietyId();
    if (!societyId) return;
    setSaving(true);
    setSaveMsg('');
    authAPI.put(`/societies/${societyId}`, form)
      .then(() => setSaveMsg('Settings saved successfully.'))
      .catch(() => setSaveMsg('Coming soon -- settings update endpoint is not yet available.'))
      .finally(() => setSaving(false));
  };

  if (loading) return <p className="text-slate-500">Loading settings...</p>;
  if (!society) return <p className="text-slate-500">Society not found. Please check your membership.</p>;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Settings</h1>
        <p className="text-slate-500 mt-1">Society configuration and preferences</p>
      </div>

      <Card>
        <CardHeader><CardTitle>Society Details</CardTitle></CardHeader>
        <CardContent>
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4 text-sm">
            <div><dt className="text-slate-500">Name</dt><dd className="font-medium">{society.name}</dd></div>
            <div><dt className="text-slate-500">Code</dt><dd className="font-mono">{society.code}</dd></div>
            <div><dt className="text-slate-500">Address</dt><dd>{society.address}</dd></div>
            <div><dt className="text-slate-500">City</dt><dd>{society.city}</dd></div>
            <div><dt className="text-slate-500">Total Flats</dt><dd>{society.total_flats}</dd></div>
            <div><dt className="text-slate-500">Subscription</dt><dd className="capitalize">{society.subscription || 'Free'}</dd></div>
            <div><dt className="text-slate-500">Maintenance Amount</dt><dd>{formatCurrency(society.maintenance_amount)}</dd></div>
            <div><dt className="text-slate-500">Due Day</dt><dd>{society.maintenance_due_day}th of every month</dd></div>
            <div><dt className="text-slate-500">Late Fee</dt><dd>{society.late_fee_percent}%</dd></div>
          </dl>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>Edit Maintenance Settings</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <Input
              label="Maintenance Amount"
              type="number"
              value={form.maintenance_amount}
              onChange={(e) => setForm({ ...form, maintenance_amount: Number(e.target.value) })}
            />
            <Input
              label="Due Day (1-28)"
              type="number"
              min={1}
              max={28}
              value={form.maintenance_due_day}
              onChange={(e) => setForm({ ...form, maintenance_due_day: Number(e.target.value) })}
            />
            <Input
              label="Late Fee %"
              type="number"
              min={0}
              max={100}
              step={0.5}
              value={form.late_fee_percent}
              onChange={(e) => setForm({ ...form, late_fee_percent: Number(e.target.value) })}
            />
          </div>
          <div className="flex items-center gap-4">
            <Button onClick={handleSave} disabled={saving}>
              {saving ? 'Saving...' : 'Save Settings'}
            </Button>
            {saveMsg && (
              <p className={`text-sm ${saveMsg.includes('success') ? 'text-green-700' : 'text-amber-700'}`}>
                {saveMsg}
              </p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
