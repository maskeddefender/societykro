import { cn } from '@/lib/utils';

interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ReactNode;
  trend?: { value: number; positive: boolean };
  className?: string;
}

export function StatCard({ title, value, subtitle, icon, trend, className }: StatCardProps) {
  return (
    <div className={cn('bg-white rounded-xl border border-slate-200 p-6 shadow-sm', className)}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-slate-500">{title}</p>
          <p className="mt-1 text-3xl font-bold text-slate-900">{value}</p>
          {subtitle && <p className="mt-1 text-sm text-slate-400">{subtitle}</p>}
          {trend && (
            <p className={cn('mt-1 text-sm font-medium', trend.positive ? 'text-green-600' : 'text-red-600')}>
              {trend.positive ? '+' : ''}{trend.value}% from last month
            </p>
          )}
        </div>
        <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-sky-50 text-sky-600">
          {icon}
        </div>
      </div>
    </div>
  );
}
