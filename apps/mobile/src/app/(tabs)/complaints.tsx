import { useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { complaintAPI } from '../../services/api';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

const STATUS_FILTERS = ['all', 'open', 'in_progress', 'resolved'] as const;
type StatusFilter = (typeof STATUS_FILTERS)[number];

const STATUS_LABELS: Record<StatusFilter, string> = {
  all: 'All',
  open: 'Open',
  in_progress: 'In Progress',
  resolved: 'Resolved',
};

const statusColors: Record<string, string> = {
  open: colors.statusOpen,
  in_progress: colors.statusInProgress,
  resolved: colors.statusResolved,
  closed: colors.statusClosed,
};

export default function ComplaintsScreen() {
  const { t } = useTranslation();
  const router = useRouter();
  const [activeFilter, setActiveFilter] = useState<StatusFilter>('all');

  const queryParam = activeFilter === 'all' ? '' : `&status=${activeFilter}`;

  const { data, isLoading, refetch, isRefetching } = useQuery({
    queryKey: ['complaints', activeFilter],
    queryFn: () =>
      complaintAPI
        .get(`/complaints?limit=20${queryParam}`)
        .then((r) => r.data),
  });

  const complaints = data?.data || [];

  const renderItem = useCallback(
    ({ item }: { item: any }) => (
      <TouchableOpacity
        style={styles.card}
        onPress={() => router.push(`/complaint/${item.id}`)}
        activeOpacity={0.7}
      >
        <View style={styles.cardHeader}>
          <Text style={styles.ticketNumber}>{item.ticket_number}</Text>
          <StatusBadge status={item.status} />
        </View>
        <Text style={styles.cardTitle} numberOfLines={1}>
          {item.title}
        </Text>
        <View style={styles.cardFooter}>
          <Text style={styles.category}>{item.category}</Text>
          <Text style={styles.meta}>
            {item.raised_by_name} &middot;{' '}
            {new Date(item.created_at).toLocaleDateString()}
          </Text>
        </View>
      </TouchableOpacity>
    ),
    [router],
  );

  return (
    <View style={styles.container}>
      {/* Status filter tabs */}
      <View style={styles.filterRow}>
        {STATUS_FILTERS.map((filter) => (
          <TouchableOpacity
            key={filter}
            style={[
              styles.filterTab,
              activeFilter === filter && styles.filterTabActive,
            ]}
            onPress={() => setActiveFilter(filter)}
          >
            <Text
              style={[
                styles.filterText,
                activeFilter === filter && styles.filterTextActive,
              ]}
            >
              {STATUS_LABELS[filter]}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      {isLoading ? (
        <ActivityIndicator
          size="large"
          color={colors.primary}
          style={styles.loader}
        />
      ) : (
        <FlatList
          data={complaints}
          keyExtractor={(item) => item.id}
          renderItem={renderItem}
          contentContainerStyle={styles.list}
          refreshControl={
            <RefreshControl refreshing={isRefetching} onRefresh={refetch} />
          }
          ListEmptyComponent={
            <Text style={styles.emptyText}>No complaints found</Text>
          }
        />
      )}

      {/* FAB */}
      <TouchableOpacity
        style={styles.fab}
        onPress={() => router.push('/complaint/new')}
        activeOpacity={0.8}
      >
        <Text style={styles.fabIcon}>+</Text>
      </TouchableOpacity>
    </View>
  );
}

function StatusBadge({ status }: { status: string }) {
  const color = statusColors[status] || colors.textTertiary;
  return (
    <View style={[styles.badge, { backgroundColor: color + '20' }]}>
      <Text style={[styles.badgeText, { color }]}>
        {status.replace('_', ' ').toUpperCase()}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  filterRow: {
    flexDirection: 'row',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.md,
    gap: spacing.sm,
    backgroundColor: colors.surface,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  filterTab: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: borderRadius.full,
    backgroundColor: colors.surfaceSecondary,
  },
  filterTabActive: {
    backgroundColor: colors.primary,
  },
  filterText: {
    fontSize: fontSize.sm,
    fontWeight: '600',
    color: colors.textSecondary,
  },
  filterTextActive: {
    color: colors.textInverse,
  },
  list: {
    padding: spacing.lg,
    paddingBottom: 100,
  },
  card: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.lg,
    marginBottom: spacing.sm,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  ticketNumber: {
    fontSize: fontSize.sm,
    fontWeight: '700',
    color: colors.primary,
  },
  cardTitle: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.text,
    marginTop: spacing.xs,
  },
  cardFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: spacing.sm,
  },
  category: {
    fontSize: fontSize.xs,
    fontWeight: '600',
    color: colors.primaryLight,
    backgroundColor: colors.primaryLight + '15',
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
    borderRadius: borderRadius.sm,
    overflow: 'hidden',
  },
  meta: {
    fontSize: fontSize.xs,
    color: colors.textTertiary,
  },
  badge: {
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
    borderRadius: borderRadius.sm,
  },
  badgeText: {
    fontSize: 10,
    fontWeight: '700',
  },
  emptyText: {
    textAlign: 'center',
    color: colors.textTertiary,
    fontSize: fontSize.md,
    marginTop: spacing.xxxl,
  },
  loader: {
    marginTop: spacing.xxxl,
  },
  fab: {
    position: 'absolute',
    bottom: spacing.xxl,
    right: spacing.xxl,
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: colors.primary,
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 6,
  },
  fabIcon: {
    fontSize: 28,
    color: colors.textInverse,
    lineHeight: 30,
  },
});
