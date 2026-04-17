'use client';

import { useEffect, useState } from 'react';
import { authAPI } from '@/services/api';
import { useAuthStore } from '@/store/authStore';
import { Badge } from '@/components/ui/badge';

interface Flat {
  id: string; flat_number: string; block: string; floor: number;
  type: string; is_occupied: boolean; occupancy_type: string;
}

export default function MembersPage() {
  const { getSocietyId } = useAuthStore();
  const [flats, setFlats] = useState<Flat[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const societyId = getSocietyId();
    if (!societyId) { setLoading(false); return; }
    authAPI.get(`/societies/${societyId}/flats`)
      .then((res) => setFlats(res.data.data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [getSocietyId]);

  const occupied = flats.filter((f) => f.is_occupied).length;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Members</h1>
        <p className="text-slate-500 mt-1">View flats and occupancy details</p>
      </div>

      {!loading && flats.length > 0 && (
        <div className="flex gap-4 text-sm">
          <span className="bg-slate-100 px-3 py-1 rounded-lg">Total: {flats.length}</span>
          <span className="bg-green-50 text-green-700 px-3 py-1 rounded-lg">Occupied: {occupied}</span>
          <span className="bg-amber-50 text-amber-700 px-3 py-1 rounded-lg">Vacant: {flats.length - occupied}</span>
        </div>
      )}

      {loading ? (
        <p className="text-slate-500">Loading flats...</p>
      ) : flats.length === 0 ? (
        <p className="text-slate-500">No flats found.</p>
      ) : (
        <div className="overflow-x-auto bg-white rounded-xl border border-slate-200 shadow-sm">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-left text-slate-500">
                <th className="px-4 py-3 font-medium">Flat#</th>
                <th className="px-4 py-3 font-medium">Block</th>
                <th className="px-4 py-3 font-medium">Floor</th>
                <th className="px-4 py-3 font-medium">Type</th>
                <th className="px-4 py-3 font-medium">Occupied</th>
                <th className="px-4 py-3 font-medium">Occupancy Type</th>
              </tr>
            </thead>
            <tbody>
              {flats.map((f) => (
                <tr key={f.id} className="border-b border-slate-50">
                  <td className="px-4 py-3 font-medium text-slate-900">{f.flat_number}</td>
                  <td className="px-4 py-3">{f.block || '-'}</td>
                  <td className="px-4 py-3">{f.floor}</td>
                  <td className="px-4 py-3">{f.type}</td>
                  <td className="px-4 py-3">
                    <Badge status={f.is_occupied ? 'approved' : 'pending'} />
                  </td>
                  <td className="px-4 py-3 text-slate-600">{f.occupancy_type || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
