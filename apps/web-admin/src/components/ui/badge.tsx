import { cn } from '@/lib/utils';

const statusColors: Record<string, string> = {
  open: 'bg-amber-100 text-amber-800',
  in_progress: 'bg-blue-100 text-blue-800',
  resolved: 'bg-green-100 text-green-800',
  closed: 'bg-slate-100 text-slate-600',
  reopened: 'bg-red-100 text-red-800',
  pending: 'bg-amber-100 text-amber-800',
  paid: 'bg-green-100 text-green-800',
  overdue: 'bg-red-100 text-red-800',
  approved: 'bg-green-100 text-green-800',
  denied: 'bg-red-100 text-red-800',
  checked_in: 'bg-blue-100 text-blue-800',
  checked_out: 'bg-slate-100 text-slate-600',
  emergency: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  normal: 'bg-slate-100 text-slate-600',
  low: 'bg-slate-50 text-slate-500',
};

export function Badge({ status, className }: { status: string; className?: string }) {
  return (
    <span className={cn(
      'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold',
      statusColors[status] || 'bg-slate-100 text-slate-600',
      className,
    )}>
      {status.replace(/_/g, ' ').toUpperCase()}
    </span>
  );
}
