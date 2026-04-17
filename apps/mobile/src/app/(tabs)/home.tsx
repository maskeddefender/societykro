import { View, Text, StyleSheet, ScrollView, TouchableOpacity, RefreshControl } from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '../../store/authStore';
import { authAPI, complaintAPI, noticeAPI } from '../../services/api';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

export default function HomeScreen() {
  const { t } = useTranslation();
  const router = useRouter();
  const { user, getSocietyId } = useAuthStore();
  const societyId = getSocietyId();

  const { data: society, refetch: refetchSociety } = useQuery({
    queryKey: ['society', societyId],
    queryFn: () => authAPI.get(`/societies/${societyId}`).then((r) => r.data.data),
    enabled: !!societyId,
  });

  const { data: complaints } = useQuery({
    queryKey: ['complaints', 'open'],
    queryFn: () => complaintAPI.get('/complaints?status=open&limit=5').then((r) => r.data.data),
    enabled: !!societyId,
  });

  const { data: notices } = useQuery({
    queryKey: ['notices'],
    queryFn: () => noticeAPI.get('/notices?limit=5').then((r) => r.data.data),
    enabled: !!societyId,
  });

  const openCount = complaints?.length || 0;

  return (
    <ScrollView
      style={styles.container}
      refreshControl={<RefreshControl refreshing={false} onRefresh={refetchSociety} />}
    >
      {/* Greeting */}
      <View style={styles.greetingCard}>
        <Text style={styles.greeting}>{t('home.greeting', { name: user?.name || 'User' })}</Text>
        {society && (
          <Text style={styles.societyName}>{society.name}</Text>
        )}
      </View>

      {/* Quick Actions */}
      <Text style={styles.sectionTitle}>{t('home.quickActions')}</Text>
      <View style={styles.actionsRow}>
        <QuickAction
          title={t('home.raiseComplaint')}
          icon="\u26A0"
          color={colors.error}
          onPress={() => router.push('/complaint/new')}
        />
        <QuickAction
          title={t('home.approveVisitor')}
          icon="\u263A"
          color={colors.info}
          onPress={() => router.push('/(tabs)/visitors')}
        />
        <QuickAction
          title={t('home.payDues')}
          icon="\u2B50"
          color={colors.success}
          onPress={() => router.push('/(tabs)/payments')}
        />
      </View>

      {/* Stats */}
      <View style={styles.statsRow}>
        <StatCard label={t('home.openComplaints')} value={openCount} color={colors.warning} />
        <StatCard label={t('home.notices')} value={notices?.length || 0} color={colors.info} />
      </View>

      {/* Recent Notices */}
      {notices && notices.length > 0 && (
        <>
          <Text style={styles.sectionTitle}>{t('home.notices')}</Text>
          {notices.map((n: any) => (
            <TouchableOpacity
              key={n.id}
              style={styles.noticeCard}
              onPress={() => router.push(`/notice/${n.id}`)}
            >
              {n.is_pinned && <Text style={styles.pinnedBadge}>PINNED</Text>}
              <Text style={styles.noticeTitle}>{n.title}</Text>
              <Text style={styles.noticeBody} numberOfLines={2}>{n.body}</Text>
              <Text style={styles.noticeMeta}>{n.created_by_name} - {new Date(n.created_at).toLocaleDateString()}</Text>
            </TouchableOpacity>
          ))}
        </>
      )}

      {/* Recent Complaints */}
      {complaints && complaints.length > 0 && (
        <>
          <Text style={styles.sectionTitle}>{t('home.openComplaints')}</Text>
          {complaints.map((c: any) => (
            <TouchableOpacity
              key={c.id}
              style={styles.complaintCard}
              onPress={() => router.push(`/complaint/${c.id}`)}
            >
              <View style={styles.complaintHeader}>
                <Text style={styles.ticketNumber}>{c.ticket_number}</Text>
                <StatusBadge status={c.status} />
              </View>
              <Text style={styles.complaintTitle}>{c.title}</Text>
              <Text style={styles.complaintMeta}>{c.category} - {c.raised_by_name}</Text>
            </TouchableOpacity>
          ))}
        </>
      )}

      <View style={{ height: spacing.xxxl }} />
    </ScrollView>
  );
}

function QuickAction({ title, icon, color, onPress }: { title: string; icon: string; color: string; onPress: () => void }) {
  return (
    <TouchableOpacity style={styles.actionCard} onPress={onPress}>
      <View style={[styles.actionIcon, { backgroundColor: color + '15' }]}>
        <Text style={{ fontSize: 24, color }}>{icon}</Text>
      </View>
      <Text style={styles.actionLabel}>{title}</Text>
    </TouchableOpacity>
  );
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <View style={styles.statCard}>
      <Text style={[styles.statValue, { color }]}>{value}</Text>
      <Text style={styles.statLabel}>{label}</Text>
    </View>
  );
}

function StatusBadge({ status }: { status: string }) {
  const statusColors: Record<string, string> = {
    open: colors.statusOpen,
    in_progress: colors.statusInProgress,
    resolved: colors.statusResolved,
    closed: colors.statusClosed,
  };
  return (
    <View style={[styles.badge, { backgroundColor: (statusColors[status] || colors.textTertiary) + '20' }]}>
      <Text style={[styles.badgeText, { color: statusColors[status] || colors.textTertiary }]}>
        {status.replace('_', ' ').toUpperCase()}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  greetingCard: {
    backgroundColor: colors.primary,
    padding: spacing.xxl,
    paddingTop: spacing.lg,
    paddingBottom: spacing.xxxl,
    borderBottomLeftRadius: borderRadius.xl,
    borderBottomRightRadius: borderRadius.xl,
  },
  greeting: { fontSize: fontSize.xl, fontWeight: '700', color: colors.textInverse },
  societyName: { fontSize: fontSize.sm, color: colors.textInverse, opacity: 0.8, marginTop: spacing.xs },
  sectionTitle: {
    fontSize: fontSize.lg, fontWeight: '700', color: colors.text,
    paddingHorizontal: spacing.lg, marginTop: spacing.xxl, marginBottom: spacing.md,
  },
  actionsRow: {
    flexDirection: 'row', paddingHorizontal: spacing.md, gap: spacing.sm,
  },
  actionCard: {
    flex: 1, backgroundColor: colors.surface, borderRadius: borderRadius.lg,
    padding: spacing.lg, alignItems: 'center',
    shadowColor: '#000', shadowOffset: { width: 0, height: 1 }, shadowOpacity: 0.05, shadowRadius: 4, elevation: 2,
  },
  actionIcon: {
    width: 48, height: 48, borderRadius: borderRadius.lg,
    justifyContent: 'center', alignItems: 'center', marginBottom: spacing.sm,
  },
  actionLabel: { fontSize: fontSize.xs, fontWeight: '600', color: colors.text, textAlign: 'center' },
  statsRow: {
    flexDirection: 'row', paddingHorizontal: spacing.md, gap: spacing.sm, marginTop: spacing.lg,
  },
  statCard: {
    flex: 1, backgroundColor: colors.surface, borderRadius: borderRadius.lg,
    padding: spacing.lg, alignItems: 'center',
    shadowColor: '#000', shadowOffset: { width: 0, height: 1 }, shadowOpacity: 0.05, shadowRadius: 4, elevation: 2,
  },
  statValue: { fontSize: fontSize.xxl, fontWeight: '800' },
  statLabel: { fontSize: fontSize.xs, color: colors.textSecondary, marginTop: spacing.xs, textAlign: 'center' },
  noticeCard: {
    backgroundColor: colors.surface, marginHorizontal: spacing.lg, marginBottom: spacing.sm,
    borderRadius: borderRadius.md, padding: spacing.lg,
    borderLeftWidth: 3, borderLeftColor: colors.primaryLight,
  },
  pinnedBadge: {
    fontSize: 9, fontWeight: '800', color: colors.primary, backgroundColor: colors.primary + '15',
    paddingHorizontal: spacing.sm, paddingVertical: 2, borderRadius: borderRadius.sm, alignSelf: 'flex-start', marginBottom: spacing.xs,
  },
  noticeTitle: { fontSize: fontSize.md, fontWeight: '700', color: colors.text },
  noticeBody: { fontSize: fontSize.sm, color: colors.textSecondary, marginTop: spacing.xs },
  noticeMeta: { fontSize: fontSize.xs, color: colors.textTertiary, marginTop: spacing.sm },
  complaintCard: {
    backgroundColor: colors.surface, marginHorizontal: spacing.lg, marginBottom: spacing.sm,
    borderRadius: borderRadius.md, padding: spacing.lg,
  },
  complaintHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  ticketNumber: { fontSize: fontSize.sm, fontWeight: '700', color: colors.primary },
  complaintTitle: { fontSize: fontSize.md, fontWeight: '600', color: colors.text, marginTop: spacing.xs },
  complaintMeta: { fontSize: fontSize.xs, color: colors.textTertiary, marginTop: spacing.xs },
  badge: { paddingHorizontal: spacing.sm, paddingVertical: 2, borderRadius: borderRadius.sm },
  badgeText: { fontSize: 10, fontWeight: '700' },
});
