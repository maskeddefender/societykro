'use client';

import { useEffect, useState } from 'react';
import { vendorAPI } from '@/services/api';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';

interface Vendor {
  id: string; name: string; category: string; phone: string;
  rating: number; total_jobs: number; is_active: boolean;
}

const CATEGORIES = [
  'plumbing', 'electrical', 'lift', 'pest_control', 'security',
  'cleaning', 'gardening', 'painting', 'carpentry', 'other',
];

export default function VendorsPage() {
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: '', phone: '', category: 'plumbing' });
  const [submitting, setSubmitting] = useState(false);

  const fetchVendors = () => {
    setLoading(true);
    vendorAPI.get('/vendors')
      .then((res) => setVendors(res.data.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetchVendors(); }, []);

  const handleSubmit = () => {
    if (!form.name.trim() || !form.phone.trim()) return;
    setSubmitting(true);
    vendorAPI.post('/vendors', form)
      .then(() => {
        setShowForm(false);
        setForm({ name: '', phone: '', category: 'plumbing' });
        fetchVendors();
      })
      .catch(() => {})
      .finally(() => setSubmitting(false));
  };

  const handleDelete = (id: string) => {
    if (!confirm('Delete this vendor?')) return;
    vendorAPI.delete(`/vendors/${id}`)
      .then(() => fetchVendors())
      .catch(() => {});
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Vendors</h1>
          <p className="text-slate-500 mt-1">Manage service vendors for your society</p>
        </div>
        <Button onClick={() => setShowForm(!showForm)}>
          {showForm ? 'Cancel' : 'Add Vendor'}
        </Button>
      </div>

      {showForm && (
        <Card>
          <CardHeader><CardTitle>New Vendor</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <Input
              label="Name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Vendor name"
            />
            <Input
              label="Phone"
              value={form.phone}
              onChange={(e) => setForm({ ...form, phone: e.target.value })}
              placeholder="Phone number"
            />
            <div className="space-y-1">
              <label className="block text-sm font-medium text-slate-700">Category</label>
              <select
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                value={form.category}
                onChange={(e) => setForm({ ...form, category: e.target.value })}
              >
                {CATEGORIES.map((c) => (
                  <option key={c} value={c}>
                    {c.replace(/_/g, ' ').replace(/\b\w/g, (ch) => ch.toUpperCase())}
                  </option>
                ))}
              </select>
            </div>
            <Button onClick={handleSubmit} disabled={submitting || !form.name.trim() || !form.phone.trim()}>
              {submitting ? 'Saving...' : 'Save Vendor'}
            </Button>
          </CardContent>
        </Card>
      )}

      {loading ? (
        <p className="text-slate-500">Loading vendors...</p>
      ) : vendors.length === 0 ? (
        <p className="text-slate-500">No vendors found.</p>
      ) : (
        <div className="overflow-x-auto bg-white rounded-xl border border-slate-200 shadow-sm">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-left text-slate-500">
                <th className="px-4 py-3 font-medium">Name</th>
                <th className="px-4 py-3 font-medium">Category</th>
                <th className="px-4 py-3 font-medium">Phone</th>
                <th className="px-4 py-3 font-medium">Rating</th>
                <th className="px-4 py-3 font-medium">Total Jobs</th>
                <th className="px-4 py-3 font-medium">Active</th>
                <th className="px-4 py-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {vendors.map((v) => (
                <tr key={v.id} className="border-b border-slate-50">
                  <td className="px-4 py-3 font-medium text-slate-900">{v.name}</td>
                  <td className="px-4 py-3 text-slate-600">
                    {v.category.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())}
                  </td>
                  <td className="px-4 py-3">{v.phone}</td>
                  <td className="px-4 py-3">{v.rating ? `${v.rating.toFixed(1)} / 5` : '-'}</td>
                  <td className="px-4 py-3">{v.total_jobs}</td>
                  <td className="px-4 py-3">
                    <span className={`inline-block w-2 h-2 rounded-full ${v.is_active ? 'bg-green-500' : 'bg-slate-300'}`} />
                  </td>
                  <td className="px-4 py-3">
                    <Button size="sm" variant="danger" onClick={() => handleDelete(v.id)}>Delete</Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
