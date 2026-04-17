'use client';

import { useEffect, useState } from 'react';
import { noticeAPI } from '@/services/api';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { formatDate } from '@/lib/utils';

interface Notice {
  id: string; title: string; body: string; category: string;
  is_pinned: boolean; created_by_name: string; created_at: string; read_count: number;
}

const CATEGORIES = ['general', 'maintenance', 'emergency', 'event', 'meeting'];

export default function NoticesPage() {
  const [notices, setNotices] = useState<Notice[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ title: '', body: '', category: 'general', is_pinned: false, broadcast_whatsapp: false });
  const [submitting, setSubmitting] = useState(false);

  const fetchNotices = () => {
    setLoading(true);
    noticeAPI.get('/notices?limit=50')
      .then((res) => setNotices(res.data.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetchNotices(); }, []);

  const handleSubmit = () => {
    if (!form.title.trim() || !form.body.trim()) return;
    setSubmitting(true);
    noticeAPI.post('/notices', form)
      .then(() => {
        setShowForm(false);
        setForm({ title: '', body: '', category: 'general', is_pinned: false, broadcast_whatsapp: false });
        fetchNotices();
      })
      .catch(() => {})
      .finally(() => setSubmitting(false));
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Notices</h1>
          <p className="text-slate-500 mt-1">Broadcast notices to society members</p>
        </div>
        <Button onClick={() => setShowForm(!showForm)}>
          {showForm ? 'Cancel' : 'Create Notice'}
        </Button>
      </div>

      {showForm && (
        <Card>
          <CardHeader><CardTitle>New Notice</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <Input
              label="Title"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              placeholder="Notice title"
            />
            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700">Body</label>
              <textarea
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                rows={4}
                value={form.body}
                onChange={(e) => setForm({ ...form, body: e.target.value })}
                placeholder="Notice content..."
              />
            </div>
            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700">Category</label>
              <select
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                value={form.category}
                onChange={(e) => setForm({ ...form, category: e.target.value })}
              >
                {CATEGORIES.map((c) => (
                  <option key={c} value={c}>{c.charAt(0).toUpperCase() + c.slice(1)}</option>
                ))}
              </select>
            </div>
            <div className="flex gap-6">
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={form.is_pinned} onChange={(e) => setForm({ ...form, is_pinned: e.target.checked })} className="rounded" />
                Pin notice
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={form.broadcast_whatsapp} onChange={(e) => setForm({ ...form, broadcast_whatsapp: e.target.checked })} className="rounded" />
                Broadcast via WhatsApp
              </label>
            </div>
            <Button onClick={handleSubmit} disabled={submitting || !form.title.trim() || !form.body.trim()}>
              {submitting ? 'Publishing...' : 'Publish Notice'}
            </Button>
          </CardContent>
        </Card>
      )}

      {loading ? (
        <p className="text-slate-500">Loading notices...</p>
      ) : notices.length === 0 ? (
        <p className="text-slate-500">No notices yet.</p>
      ) : (
        <div className="grid gap-4">
          {notices.map((n) => (
            <Card key={n.id}>
              <CardContent className="py-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-semibold text-slate-900">{n.title}</h3>
                      {n.is_pinned && <Badge status="emergency" className="text-[10px]" />}
                      <Badge status={n.category} />
                    </div>
                    <p className="text-sm text-slate-600 line-clamp-2">{n.body}</p>
                    <div className="flex gap-4 mt-2 text-xs text-slate-500">
                      <span>By {n.created_by_name}</span>
                      <span>{formatDate(n.created_at)}</span>
                      <span>{n.read_count} reads</span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
