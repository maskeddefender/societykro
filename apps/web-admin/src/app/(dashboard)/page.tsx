'use client';

import { useEffect, useState } from 'react';
import { AlertTriangle, CreditCard, Users, Shield } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { StatCard } from '@/components/ui/stat-card';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { complaintAPI, noticeAPI } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { formatDate } from '@/lib/utils';

const PIE_COLORS = ['#f59e0b', '#3b82f6', '#22c55e', '#94a3b8', '#ef4444'];

export default function DashboardPage() {
  const societyId = useAuthStore((s) => s.getSocietyId());
  const [stats, setStats] = useState({ complaints: 0, pendingPayments: 0, members: 0, visitorsToday: 0 });
  const [complaintStats, setComplaintStats] = useState<Record<string, number>>({});
  const [recentComplaints, setRecentComplaints] = useState<any[]>([]);
  const [recentNotices, setRecentNotices] = useState<any[]>([]);

  useEffect(() => {
    if (!societyId) return;

    // Fetch complaint stats
    complaintAPI.get('/complaints/analytics').then((r) => {
      const counts = r.data.data?.counts || {};
      setComplaintStats(counts);
      setStats((s) => ({ ...s, complaints: (counts.open || 0) + (counts.in_progress || 0) }));
    }).catch(() => {});

    // Recent complaints
    complaintAPI.get('/complaints?limit=5').then((r) => {
      setRecentComplaints(r.data.data || []);
    }).catch(() => {});

    // Recent notices
    noticeAPI.get('/notices?limit=5').then((r) => {
      setRecentNotices(r.data.data || []);
    }).catch(() => {});
  }, [societyId]);

  const pieData = Object.entries(complaintStats).map(([name, value]) => ({ name, value }));

  const barData = [
    { month: 'Jan', complaints: 12 },
    { month: 'Feb', complaints: 19 },
    { month: 'Mar', complaints: 8 },
    { month: 'Apr', complaints: stats.complaints },
  ];

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-slate-900">Dashboard</h1>
        <p className="text-sm text-slate-500">Overview of your society</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <StatCard
          title="Open Complaints"
          value={stats.complaints}
          icon={<AlertTriangle size={24} />}
        />
        <StatCard
          title="Pending Payments"
          value={stats.pendingPayments}
          subtitle="This month"
          icon={<CreditCard size={24} />}
        />
        <StatCard
          title="Total Members"
          value={stats.members || '—'}
          icon={<Users size={24} />}
        />
        <StatCard
          title="Visitors Today"
          value={stats.visitorsToday}
          icon={<Shield size={24} />}
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <Card>
          <CardHeader><CardTitle>Complaint Trend</CardTitle></CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={barData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
                  <XAxis dataKey="month" tick={{ fontSize: 12 }} />
                  <YAxis tick={{ fontSize: 12 }} />
                  <Tooltip />
                  <Bar dataKey="complaints" fill="#1e40af" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Complaints by Status</CardTitle></CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center">
              {pieData.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie data={pieData} cx="50%" cy="50%" outerRadius={80} dataKey="value" label={({ name, value }) => `${name}: ${value}`}>
                      {pieData.map((_, i) => <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />)}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              ) : (
                <p className="text-slate-400 text-sm">No complaint data yet</p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader><CardTitle>Recent Complaints</CardTitle></CardHeader>
          <CardContent>
            {recentComplaints.length === 0 ? (
              <p className="text-sm text-slate-400">No complaints yet</p>
            ) : (
              <div className="space-y-3">
                {recentComplaints.map((c: any) => (
                  <div key={c.id} className="flex items-center justify-between py-2 border-b border-slate-50 last:border-0">
                    <div>
                      <p className="text-sm font-medium text-slate-900">{c.title}</p>
                      <p className="text-xs text-slate-500">{c.ticket_number} — {c.category} — {formatDate(c.created_at)}</p>
                    </div>
                    <Badge status={c.status} />
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Recent Notices</CardTitle></CardHeader>
          <CardContent>
            {recentNotices.length === 0 ? (
              <p className="text-sm text-slate-400">No notices yet</p>
            ) : (
              <div className="space-y-3">
                {recentNotices.map((n: any) => (
                  <div key={n.id} className="py-2 border-b border-slate-50 last:border-0">
                    <div className="flex items-center gap-2">
                      {n.is_pinned && <span className="text-xs bg-sky-100 text-sky-800 px-1.5 py-0.5 rounded font-medium">PINNED</span>}
                      <p className="text-sm font-medium text-slate-900">{n.title}</p>
                    </div>
                    <p className="text-xs text-slate-500 mt-1">{n.created_by_name} — {formatDate(n.created_at)}</p>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
