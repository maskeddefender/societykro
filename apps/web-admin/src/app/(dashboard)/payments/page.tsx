'use client';

import { useEffect, useState } from 'react';
import { paymentAPI } from '@/services/api';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { formatDate, formatCurrency } from '@/lib/utils';

interface Payment {
  id: string; invoice_number: string; flat_number: string; month: string;
  amount: number; status: string; paid_at: string;
}
interface Defaulter { flat_number: string; block: string; months_overdue: number; total_due: number; }

export default function PaymentsPage() {
  const [tab, setTab] = useState<'overview' | 'generate'>('overview');
  const [payments, setPayments] = useState<Payment[]>([]);
  const [defaulters, setDefaulters] = useState<Defaulter[]>([]);
  const [loading, setLoading] = useState(true);
  const [month, setMonth] = useState('');
  const [genResult, setGenResult] = useState<string | null>(null);
  const [generating, setGenerating] = useState(false);
  const [cashNotes, setCashNotes] = useState<Record<string, string>>({});

  useEffect(() => {
    Promise.all([
      paymentAPI.get('/payments?limit=50').then((r) => setPayments(r.data.data || [])).catch(() => {}),
      paymentAPI.get('/payments/defaulters').then((r) => setDefaulters(r.data.data || [])).catch(() => {}),
    ]).finally(() => setLoading(false));
  }, []);

  const generateBills = () => {
    if (!month) return;
    setGenerating(true);
    paymentAPI.post('/payments/generate-bills', { month })
      .then((r) => setGenResult(`Bills generated successfully: ${r.data.data?.count || 0} invoices created.`))
      .catch(() => setGenResult('Failed to generate bills.'))
      .finally(() => setGenerating(false));
  };

  const recordCash = (paymentId: string) => {
    paymentAPI.post(`/payments/${paymentId}/record-cash`, { method: 'cash', notes: cashNotes[paymentId] || '' })
      .then(() => {
        setPayments((prev) => prev.map((p) => p.id === paymentId ? { ...p, status: 'paid' } : p));
      })
      .catch(() => {});
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Payments</h1>
        <p className="text-slate-500 mt-1">Manage maintenance payments and billing</p>
      </div>

      <div className="flex gap-2">
        <Button variant={tab === 'overview' ? 'primary' : 'outline'} size="sm" onClick={() => setTab('overview')}>Overview</Button>
        <Button variant={tab === 'generate' ? 'primary' : 'outline'} size="sm" onClick={() => setTab('generate')}>Generate Bills</Button>
      </div>

      {tab === 'overview' && (
        <>
          {loading ? <p className="text-slate-500">Loading...</p> : (
            <div className="overflow-x-auto bg-white rounded-xl border border-slate-200 shadow-sm">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-100 text-left text-slate-500">
                    <th className="px-4 py-3 font-medium">Invoice#</th>
                    <th className="px-4 py-3 font-medium">Flat</th>
                    <th className="px-4 py-3 font-medium">Month</th>
                    <th className="px-4 py-3 font-medium">Amount</th>
                    <th className="px-4 py-3 font-medium">Status</th>
                    <th className="px-4 py-3 font-medium">Paid Date</th>
                    <th className="px-4 py-3 font-medium">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {payments.map((p) => (
                    <tr key={p.id} className="border-b border-slate-50">
                      <td className="px-4 py-3 font-mono text-xs">{p.invoice_number}</td>
                      <td className="px-4 py-3">{p.flat_number}</td>
                      <td className="px-4 py-3">{p.month}</td>
                      <td className="px-4 py-3 font-medium">{formatCurrency(p.amount)}</td>
                      <td className="px-4 py-3"><Badge status={p.status} /></td>
                      <td className="px-4 py-3 text-slate-500">{p.paid_at ? formatDate(p.paid_at) : '-'}</td>
                      <td className="px-4 py-3">
                        {p.status === 'pending' && (
                          <div className="flex items-center gap-2">
                            <input
                              className="rounded border border-slate-300 px-2 py-1 text-xs w-24"
                              placeholder="Notes"
                              value={cashNotes[p.id] || ''}
                              onChange={(e) => setCashNotes((prev) => ({ ...prev, [p.id]: e.target.value }))}
                            />
                            <Button size="sm" variant="success" onClick={() => recordCash(p.id)}>Cash</Button>
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {defaulters.length > 0 && (
            <Card>
              <CardHeader><CardTitle>Defaulters</CardTitle></CardHeader>
              <CardContent>
                <div className="space-y-2">
                  {defaulters.map((d, i) => (
                    <div key={i} className="flex items-center justify-between text-sm py-2 border-b border-slate-50 last:border-0">
                      <div>
                        <span className="font-medium">{d.flat_number}</span>
                        {d.block && <span className="text-slate-500 ml-2">Block {d.block}</span>}
                      </div>
                      <div className="text-right">
                        <span className="text-red-600 font-medium">{formatCurrency(d.total_due)}</span>
                        <span className="text-slate-500 text-xs ml-2">({d.months_overdue} months)</span>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </>
      )}

      {tab === 'generate' && (
        <Card>
          <CardHeader><CardTitle>Generate Monthly Bills</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-end gap-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Month</label>
                <input
                  type="month"
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm"
                  value={month}
                  onChange={(e) => setMonth(e.target.value)}
                />
              </div>
              <Button onClick={generateBills} disabled={!month || generating}>
                {generating ? 'Generating...' : 'Generate Bills'}
              </Button>
            </div>
            {genResult && <p className="text-sm text-green-700 bg-green-50 rounded-lg px-4 py-2">{genResult}</p>}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
