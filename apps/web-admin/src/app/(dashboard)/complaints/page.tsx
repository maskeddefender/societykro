'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { complaintAPI } from '@/services/api';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { formatDate } from '@/lib/utils';

interface Complaint {
  id: string;
  ticket_number: string;
  title: string;
  category: string;
  status: string;
  priority: string;
  raised_by_name: string;
  created_at: string;
}

const STATUSES = ['all', 'open', 'in_progress', 'resolved', 'closed'];

export default function ComplaintsPage() {
  const router = useRouter();
  const [complaints, setComplaints] = useState<Complaint[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');
  const [search, setSearch] = useState('');

  useEffect(() => {
    complaintAPI.get('/complaints?limit=50')
      .then((res) => setComplaints(res.data.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const filtered = complaints.filter((c) => {
    const matchesStatus = filter === 'all' || c.status === filter;
    const matchesSearch = !search || c.title.toLowerCase().includes(search.toLowerCase())
      || c.ticket_number.toLowerCase().includes(search.toLowerCase());
    return matchesStatus && matchesSearch;
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Complaints</h1>
        <p className="text-slate-500 mt-1">Manage and track society complaints</p>
      </div>

      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div className="flex gap-2 flex-wrap">
          {STATUSES.map((s) => (
            <Button
              key={s}
              variant={filter === s ? 'primary' : 'outline'}
              size="sm"
              onClick={() => setFilter(s)}
            >
              {s === 'all' ? 'All' : s.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())}
            </Button>
          ))}
        </div>
        <div className="w-full sm:w-64">
          <Input
            placeholder="Search by title or ticket..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      {loading ? (
        <p className="text-slate-500">Loading complaints...</p>
      ) : filtered.length === 0 ? (
        <p className="text-slate-500">No complaints found.</p>
      ) : (
        <div className="overflow-x-auto bg-white rounded-xl border border-slate-200 shadow-sm">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-left text-slate-500">
                <th className="px-4 py-3 font-medium">Ticket#</th>
                <th className="px-4 py-3 font-medium">Title</th>
                <th className="px-4 py-3 font-medium">Category</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Priority</th>
                <th className="px-4 py-3 font-medium">Raised By</th>
                <th className="px-4 py-3 font-medium">Date</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((c) => (
                <tr
                  key={c.id}
                  onClick={() => router.push(`/complaints/${c.id}`)}
                  className="border-b border-slate-50 hover:bg-slate-50 cursor-pointer transition-colors"
                >
                  <td className="px-4 py-3 font-mono text-xs">{c.ticket_number}</td>
                  <td className="px-4 py-3 font-medium text-slate-900">{c.title}</td>
                  <td className="px-4 py-3 text-slate-600">{c.category}</td>
                  <td className="px-4 py-3"><Badge status={c.status} /></td>
                  <td className="px-4 py-3"><Badge status={c.priority} /></td>
                  <td className="px-4 py-3 text-slate-600">{c.raised_by_name}</td>
                  <td className="px-4 py-3 text-slate-500">{formatDate(c.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
