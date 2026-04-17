'use client';

import { useEffect, useState } from 'react';
import { visitorAPI } from '@/services/api';
import { Badge } from '@/components/ui/badge';
import { formatDate } from '@/lib/utils';

interface Visitor {
  id: string; name: string; purpose: string; flat_number: string;
  status: string; check_in_time: string; check_out_time: string;
}

const STATUSES = ['all', 'checked_in', 'checked_out', 'approved', 'denied'];

export default function VisitorsPage() {
  const [visitors, setVisitors] = useState<Visitor[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');

  useEffect(() => {
    visitorAPI.get('/visitors?limit=50')
      .then((res) => setVisitors(res.data.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const filtered = visitors.filter((v) => filter === 'all' || v.status === filter);

  const formatTime = (t: string) => {
    if (!t) return '-';
    return new Date(t).toLocaleString('en-IN', {
      day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit',
    });
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Visitor Log</h1>
        <p className="text-slate-500 mt-1">View visitor entries and exits</p>
      </div>

      <div className="flex gap-2 flex-wrap">
        <label className="text-sm font-medium text-slate-700 self-center mr-2">Status:</label>
        <select
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        >
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {s === 'all' ? 'All' : s.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())}
            </option>
          ))}
        </select>
      </div>

      {loading ? (
        <p className="text-slate-500">Loading visitors...</p>
      ) : filtered.length === 0 ? (
        <p className="text-slate-500">No visitors found.</p>
      ) : (
        <div className="overflow-x-auto bg-white rounded-xl border border-slate-200 shadow-sm">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-left text-slate-500">
                <th className="px-4 py-3 font-medium">Name</th>
                <th className="px-4 py-3 font-medium">Purpose</th>
                <th className="px-4 py-3 font-medium">Flat</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Check-in</th>
                <th className="px-4 py-3 font-medium">Check-out</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((v) => (
                <tr key={v.id} className="border-b border-slate-50">
                  <td className="px-4 py-3 font-medium text-slate-900">{v.name}</td>
                  <td className="px-4 py-3 text-slate-600">{v.purpose}</td>
                  <td className="px-4 py-3">{v.flat_number}</td>
                  <td className="px-4 py-3"><Badge status={v.status} /></td>
                  <td className="px-4 py-3 text-slate-500">{formatTime(v.check_in_time)}</td>
                  <td className="px-4 py-3 text-slate-500">{formatTime(v.check_out_time)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
