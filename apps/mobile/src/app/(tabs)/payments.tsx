import { View, Text, StyleSheet, FlatList, TouchableOpacity, RefreshControl } from 'react-native';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { paymentAPI } from '../../services/api';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

export default function PaymentsScreen() {
  const { t } = useTranslation();

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['payments'],
    queryFn: () => paymentAPI.get('/payments?limit=20').then((r) => r.data.data),
  });

  const { data: pending } = useQuery({
    queryKey: ['payments', 'pending'],
    queryFn: () => paymentAPI.get('/payments/pending').then((r) => r.data.data),
  });

  const pendingAmount = pending?.total_due || 0;

  const statusColor = (s: string) => {
    const map: Record<string, string> = {
      pending: colors.warning, paid: colors.success, overdue: colors.error, partial: colors.info,
    };
    return map[s] || colors.textTertiary;
  };

  return (
    <View style={styles.container}>
      {/* Pending Banner */}
      {pendingAmount > 0 && (
        <View style={styles.pendingBanner}>
          <View>
            <Text style={styles.pendingLabel}>{t('payments.pending')}</Text>
            <Text style={styles.pendingAmount}>Rs {pendingAmount.toLocaleString()}</Text>
          </View>
          <TouchableOpacity style={styles.payButton}>
            <Text style={styles.payButtonText}>{t('payments.payNow')}</Text>
          </TouchableOpacity>
        </View>
      )}

      {/* Payment List */}
      <FlatList
        data={data || []}
        keyExtractor={(item) => item.id}
        refreshControl={<RefreshControl refreshing={isLoading} onRefresh={refetch} />}
        contentContainerStyle={{ padding: spacing.lg }}
        ListEmptyComponent={
          <Text style={styles.empty}>{t('common.noData')}</Text>
        }
        renderItem={({ item }) => (
          <View style={styles.card}>
            <View style={styles.cardHeader}>
              <Text style={styles.invoiceNo}>{item.invoice_number}</Text>
              <View style={[styles.badge, { backgroundColor: statusColor(item.status) + '20' }]}>
                <Text style={[styles.badgeText, { color: statusColor(item.status) }]}>
                  {item.status.toUpperCase()}
                </Text>
              </View>
            </View>
            <Text style={styles.month}>
              {new Date(item.bill_month).toLocaleDateString('en-IN', { month: 'long', year: 'numeric' })}
            </Text>
            <View style={styles.cardRow}>
              <Text style={styles.amount}>Rs {Number(item.total_due).toLocaleString()}</Text>
              <Text style={styles.dueDate}>Due: {new Date(item.due_date).toLocaleDateString('en-IN')}</Text>
            </View>
            {item.paid_at && (
              <Text style={styles.paidDate}>Paid: {new Date(item.paid_at).toLocaleDateString('en-IN')}</Text>
            )}
          </View>
        )}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  pendingBanner: {
    backgroundColor: colors.warning + '15', margin: spacing.lg, borderRadius: borderRadius.lg,
    padding: spacing.lg, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    borderWidth: 1, borderColor: colors.warning + '30',
  },
  pendingLabel: { fontSize: fontSize.sm, color: colors.textSecondary, fontWeight: '500' },
  pendingAmount: { fontSize: fontSize.xxl, fontWeight: '800', color: colors.warning, marginTop: spacing.xs },
  payButton: {
    backgroundColor: colors.success, paddingHorizontal: spacing.xl, paddingVertical: spacing.md,
    borderRadius: borderRadius.md,
  },
  payButtonText: { color: colors.textInverse, fontWeight: '700', fontSize: fontSize.md },
  card: {
    backgroundColor: colors.surface, borderRadius: borderRadius.md, padding: spacing.lg,
    marginBottom: spacing.sm,
  },
  cardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  invoiceNo: { fontSize: fontSize.sm, fontWeight: '700', color: colors.primary },
  badge: { paddingHorizontal: spacing.sm, paddingVertical: 2, borderRadius: borderRadius.sm },
  badgeText: { fontSize: 10, fontWeight: '700' },
  month: { fontSize: fontSize.md, fontWeight: '600', color: colors.text, marginTop: spacing.xs },
  cardRow: { flexDirection: 'row', justifyContent: 'space-between', marginTop: spacing.sm },
  amount: { fontSize: fontSize.lg, fontWeight: '800', color: colors.text },
  dueDate: { fontSize: fontSize.sm, color: colors.textSecondary },
  paidDate: { fontSize: fontSize.xs, color: colors.success, marginTop: spacing.xs },
  empty: { textAlign: 'center', color: colors.textTertiary, marginTop: spacing.xxxl, fontSize: fontSize.md },
});
