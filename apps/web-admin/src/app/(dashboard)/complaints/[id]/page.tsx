'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { complaintAPI, vendorAPI } from '@/services/api';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { formatDate } from '@/lib/utils';

interface Complaint {
  id: string; ticket_number: string; title: string; description: string;
  category: string; status: string; priority: string; raised_by_name: string;
  assigned_vendor_name: string; created_at: string;
}
interface Comment { id: string; comment: string; created_by_name: string; created_at: string; }
interface Vendor { id: string; name: string; category: string; }

export default function ComplaintDetailPage() {
  const router = useRouter();
  const { id } = useParams<{ id: string }>();
  const [complaint, setComplaint] = useState<Complaint | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [newComment, setNewComment] = useState('');
  const [selectedVendor, setSelectedVendor] = useState('');
  const [showVendorSelect, setShowVendorSelect] = useState(false);
  const [loading, setLoading] = useState(true);

  const fetchComplaint = () => {
    complaintAPI.get(`/complaints/${id}`)
      .then((res) => setComplaint(res.data.data))
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  const fetchComments = () => {
    complaintAPI.get(`/complaints/${id}/comments`)
      .then((res) => setComments(res.data.data || []))
      .catch(() => {});
  };

  useEffect(() => {
    fetchComplaint();
    fetchComments();
    vendorAPI.get('/vendors').then((res) => setVendors(res.data.data || [])).catch(() => {});
  }, [id]);

  const updateStatus = (status: string) => {
    complaintAPI.put(`/complaints/${id}/status`, { status })
      .then(() => fetchComplaint())
      .catch(() => {});
  };

  const assignVendor = () => {
    if (!selectedVendor) return;
    complaintAPI.put(`/complaints/${id}/assign`, { vendor_id: selectedVendor })
      .then(() => { setShowVendorSelect(false); fetchComplaint(); })
      .catch(() => {});
  };

  const addComment = () => {
    if (!newComment.trim()) return;
    complaintAPI.post(`/complaints/${id}/comments`, { comment: newComment })
      .then(() => { setNewComment(''); fetchComments(); })
      .catch(() => {});
  };

  if (loading) return <p className="text-slate-500">Loading...</p>;
  if (!complaint) return <p className="text-slate-500">Complaint not found.</p>;

  const c = complaint;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" onClick={() => router.push('/complaints')}>Back</Button>
        <h1 className="text-2xl font-bold text-slate-900">Complaint {c.ticket_number}</h1>
      </div>

      <Card>
        <CardHeader><CardTitle>Details</CardTitle></CardHeader>
        <CardContent>
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4 text-sm">
            <div><dt className="text-slate-500">Title</dt><dd className="font-medium">{c.title}</dd></div>
            <div><dt className="text-slate-500">Category</dt><dd>{c.category}</dd></div>
            <div><dt className="text-slate-500">Status</dt><dd><Badge status={c.status} /></dd></div>
            <div><dt className="text-slate-500">Priority</dt><dd><Badge status={c.priority} /></dd></div>
            <div><dt className="text-slate-500">Raised By</dt><dd>{c.raised_by_name}</dd></div>
            <div><dt className="text-slate-500">Created</dt><dd>{formatDate(c.created_at)}</dd></div>
            <div><dt className="text-slate-500">Assigned Vendor</dt><dd>{c.assigned_vendor_name || 'None'}</dd></div>
            <div className="sm:col-span-2"><dt className="text-slate-500">Description</dt><dd className="mt-1 whitespace-pre-wrap">{c.description}</dd></div>
          </dl>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>Actions</CardTitle></CardHeader>
        <CardContent className="flex flex-wrap gap-3">
          {c.status === 'open' && (
            <>
              <Button size="sm" onClick={() => setShowVendorSelect(true)}>Assign Vendor</Button>
              <Button size="sm" variant="secondary" onClick={() => updateStatus('in_progress')}>Mark In Progress</Button>
            </>
          )}
          {c.status === 'in_progress' && (
            <Button size="sm" variant="success" onClick={() => updateStatus('resolved')}>Mark Resolved</Button>
          )}
          {c.status === 'resolved' && (
            <Button size="sm" variant="secondary" onClick={() => updateStatus('closed')}>Close</Button>
          )}
          {showVendorSelect && (
            <div className="flex items-center gap-2 w-full mt-2">
              <select
                className="rounded-lg border border-slate-300 px-3 py-2 text-sm flex-1"
                value={selectedVendor}
                onChange={(e) => setSelectedVendor(e.target.value)}
              >
                <option value="">Select vendor...</option>
                {vendors.map((v) => (
                  <option key={v.id} value={v.id}>{v.name} ({v.category})</option>
                ))}
              </select>
              <Button size="sm" onClick={assignVendor}>Assign</Button>
              <Button size="sm" variant="ghost" onClick={() => setShowVendorSelect(false)}>Cancel</Button>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>Comments</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          {comments.length === 0 && <p className="text-sm text-slate-500">No comments yet.</p>}
          {comments.map((cm) => (
            <div key={cm.id} className="border-b border-slate-100 pb-3 last:border-0">
              <div className="flex justify-between text-xs text-slate-500 mb-1">
                <span className="font-medium text-slate-700">{cm.created_by_name}</span>
                <span>{formatDate(cm.created_at)}</span>
              </div>
              <p className="text-sm">{cm.comment}</p>
            </div>
          ))}
          <div className="flex gap-2 pt-2">
            <Input
              placeholder="Add a comment..."
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              className="flex-1"
            />
            <Button size="sm" onClick={addComment} disabled={!newComment.trim()}>Post</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
