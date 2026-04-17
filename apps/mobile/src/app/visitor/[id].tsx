import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  ActivityIndicator,
  Image,
} from 'react-native';
import { useLocalSearchParams } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { visitorAPI } from '../../services/api';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

const statusColors: Record<string, string> = {
  pending: colors.statusOpen,
  approved: colors.statusInProgress,
  checked_in: colors.success,
  checked_out: colors.textTertiary,
  denied: colors.error,
};

export default function VisitorDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { t } = useTranslation();

  const { data: visitor, isLoading } = useQuery({
    queryKey: ['visitor', id],
    queryFn: () =>
      visitorAPI.get(`/visitors/${id}`).then((r) => r.data.data),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <View style={styles.loader}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  if (!visitor) {
    return (
      <View style={styles.loader}>
        <Text style={styles.errorText}>Visitor not found</Text>
      </View>
    );
  }

  const sColor = statusColors[visitor.status] || colors.textTertiary;

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.content}
    >
      {/* Photo */}
      {visitor.photo && (
        <Image source={{ uri: visitor.photo }} style={styles.photo} />
      )}

      {/* Name & Status */}
      <View style={styles.header}>
        <Text style={styles.name}>{visitor.name}</Text>
        <View style={[styles.badge, { backgroundColor: sColor + '20' }]}>
          <Text style={[styles.badgeText, { color: sColor }]}>
            {visitor.status.replace('_', ' ').toUpperCase()}
          </Text>
        </View>
      </View>

      {/* Details */}
      <View style={styles.detailCard}>
        {visitor.phone && (
          <DetailRow label="Phone" value={visitor.phone} />
        )}
        <DetailRow label="Purpose" value={visitor.purpose} />
        {visitor.flat_number && (
          <DetailRow label="Flat" value={visitor.flat_number} />
        )}
        {visitor.vehicle_number && (
          <DetailRow label="Vehicle" value={visitor.vehicle_number} />
        )}
      </View>

      {/* Timestamps */}
      <View style={styles.detailCard}>
        {visitor.checked_in_at && (
          <DetailRow
            label="Checked In"
            value={new Date(visitor.checked_in_at).toLocaleString()}
          />
        )}
        {visitor.checked_out_at && (
          <DetailRow
            label="Checked Out"
            value={new Date(visitor.checked_out_at).toLocaleString()}
          />
        )}
      </View>
    </ScrollView>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.row}>
      <Text style={styles.rowLabel}>{label}</Text>
      <Text style={styles.rowValue}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  content: { padding: spacing.lg, paddingBottom: spacing.xxxl },
  loader: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  errorText: { fontSize: fontSize.md, color: colors.textTertiary },
  photo: {
    width: '100%',
    height: 200,
    borderRadius: borderRadius.lg,
    marginBottom: spacing.lg,
    backgroundColor: colors.surfaceSecondary,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.lg,
  },
  name: {
    fontSize: fontSize.xxl,
    fontWeight: '800',
    color: colors.text,
    flex: 1,
  },
  badge: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    borderRadius: borderRadius.sm,
    marginLeft: spacing.md,
  },
  badgeText: {
    fontSize: fontSize.xs,
    fontWeight: '700',
  },
  detailCard: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.lg,
    marginBottom: spacing.md,
    gap: spacing.md,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  rowLabel: {
    fontSize: fontSize.sm,
    color: colors.textTertiary,
  },
  rowValue: {
    fontSize: fontSize.sm,
    fontWeight: '600',
    color: colors.text,
  },
});
