import { useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  RefreshControl,
  ActivityIndicator,
  Modal,
  TextInput,
  Alert,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { visitorAPI } from '../../services/api';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

const PURPOSE_OPTIONS = [
  'delivery',
  'guest',
  'cab',
  'maintenance',
  'other',
] as const;

const PURPOSE_LABELS: Record<string, string> = {
  delivery: 'Delivery',
  guest: 'Guest',
  cab: 'Cab',
  maintenance: 'Maintenance',
  other: 'Other',
};

export default function VisitorsScreen() {
  const { t } = useTranslation();
  const router = useRouter();
  const queryClient = useQueryClient();
  const [showPreApprove, setShowPreApprove] = useState(false);
  const [preApproveName, setPreApproveName] = useState('');
  const [preApprovePurpose, setPreApprovePurpose] = useState('');

  const {
    data: pendingData,
    isLoading: pendingLoading,
    refetch: refetchPending,
    isRefetching: pendingRefetching,
  } = useQuery({
    queryKey: ['visitors', 'pending'],
    queryFn: () =>
      visitorAPI.get('/visitors?status=pending').then((r) => r.data.data || []),
  });

  const {
    data: todayData,
    isLoading: todayLoading,
    refetch: refetchToday,
    isRefetching: todayRefetching,
  } = useQuery({
    queryKey: ['visitors', 'today'],
    queryFn: () =>
      visitorAPI.get('/visitors?limit=20').then((r) => r.data.data || []),
  });

  const approveMutation = useMutation({
    mutationFn: (id: string) => visitorAPI.put(`/visitors/${id}/approve`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['visitors'] });
    },
  });

  const denyMutation = useMutation({
    mutationFn: (id: string) =>
      visitorAPI.put(`/visitors/${id}/deny`, { reason: 'Denied by resident' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['visitors'] });
    },
  });

  const preApproveMutation = useMutation({
    mutationFn: (payload: { name: string; purpose: string }) =>
      visitorAPI.post('/visitors/pre-approve', payload),
    onSuccess: (res) => {
      const otp = res.data?.data?.otp || res.data?.data?.code || 'N/A';
      setShowPreApprove(false);
      setPreApproveName('');
      setPreApprovePurpose('');
      queryClient.invalidateQueries({ queryKey: ['visitors'] });
      Alert.alert('Pre-Approved', `Share this OTP with your visitor: ${otp}`);
    },
    onError: () => {
      Alert.alert('Error', 'Failed to pre-approve visitor.');
    },
  });

  const handleRefresh = useCallback(() => {
    refetchPending();
    refetchToday();
  }, [refetchPending, refetchToday]);

  const isRefreshing = pendingRefetching || todayRefetching;
  const isLoading = pendingLoading || todayLoading;
  const pendingVisitors = pendingData || [];
  const todayVisitors = todayData || [];

  const handlePreApproveSubmit = () => {
    if (!preApproveName.trim()) {
      Alert.alert('Required', 'Please enter visitor name.');
      return;
    }
    if (!preApprovePurpose) {
      Alert.alert('Required', 'Please select a purpose.');
      return;
    }
    preApproveMutation.mutate({
      name: preApproveName.trim(),
      purpose: preApprovePurpose,
    });
  };

  if (isLoading) {
    return (
      <View style={styles.loader}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <FlatList
        data={todayVisitors}
        keyExtractor={(item) => item.id}
        refreshControl={
          <RefreshControl refreshing={isRefreshing} onRefresh={handleRefresh} />
        }
        contentContainerStyle={styles.list}
        ListHeaderComponent={
          <>
            {/* Pre-approve button */}
            <TouchableOpacity
              style={styles.preApproveButton}
              onPress={() => setShowPreApprove(true)}
              activeOpacity={0.8}
            >
              <Text style={styles.preApproveText}>+ Pre-approve Visitor</Text>
            </TouchableOpacity>

            {/* Pending Approval */}
            {pendingVisitors.length > 0 && (
              <>
                <Text style={styles.sectionTitle}>
                  Pending Approval ({pendingVisitors.length})
                </Text>
                {pendingVisitors.map((v: any) => (
                  <View key={v.id} style={styles.pendingCard}>
                    <TouchableOpacity
                      style={styles.pendingInfo}
                      onPress={() => router.push(`/visitor/${v.id}`)}
                    >
                      <Text style={styles.visitorName}>{v.name}</Text>
                      <Text style={styles.visitorMeta}>
                        {v.purpose} &middot; {v.flat_number}
                      </Text>
                      <Text style={styles.visitorTime}>
                        {new Date(v.created_at || v.checked_in_at).toLocaleTimeString([], {
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </Text>
                    </TouchableOpacity>
                    <View style={styles.actionButtons}>
                      <TouchableOpacity
                        style={styles.approveBtn}
                        onPress={() => approveMutation.mutate(v.id)}
                        disabled={approveMutation.isPending}
                      >
                        <Text style={styles.approveBtnText}>Approve</Text>
                      </TouchableOpacity>
                      <TouchableOpacity
                        style={styles.denyBtn}
                        onPress={() => denyMutation.mutate(v.id)}
                        disabled={denyMutation.isPending}
                      >
                        <Text style={styles.denyBtnText}>Deny</Text>
                      </TouchableOpacity>
                    </View>
                  </View>
                ))}
              </>
            )}

            {/* Today section header */}
            <Text style={styles.sectionTitle}>Today</Text>
          </>
        }
        renderItem={({ item }) => (
          <TouchableOpacity
            style={styles.visitorCard}
            onPress={() => router.push(`/visitor/${item.id}`)}
            activeOpacity={0.7}
          >
            <View style={styles.visitorRow}>
              <View style={styles.visitorDetails}>
                <Text style={styles.visitorName}>{item.name}</Text>
                <Text style={styles.visitorMeta}>
                  {item.purpose}
                  {item.flat_number ? ` \u00B7 ${item.flat_number}` : ''}
                </Text>
              </View>
              <VisitorStatusBadge status={item.status} />
            </View>
          </TouchableOpacity>
        )}
        ListEmptyComponent={
          <Text style={styles.emptyText}>No visitors today</Text>
        }
      />

      {/* Pre-approve Modal */}
      <Modal
        visible={showPreApprove}
        animationType="slide"
        transparent
        onRequestClose={() => setShowPreApprove(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>Pre-approve Visitor</Text>

            <Text style={styles.label}>Visitor Name *</Text>
            <TextInput
              style={styles.input}
              placeholder="Enter visitor name"
              placeholderTextColor={colors.textTertiary}
              value={preApproveName}
              onChangeText={setPreApproveName}
            />

            <Text style={styles.label}>Purpose *</Text>
            <View style={styles.chipRow}>
              {PURPOSE_OPTIONS.map((p) => (
                <TouchableOpacity
                  key={p}
                  style={[
                    styles.chip,
                    preApprovePurpose === p && styles.chipActive,
                  ]}
                  onPress={() => setPreApprovePurpose(p)}
                >
                  <Text
                    style={[
                      styles.chipText,
                      preApprovePurpose === p && styles.chipTextActive,
                    ]}
                  >
                    {PURPOSE_LABELS[p]}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            <View style={styles.modalActions}>
              <TouchableOpacity
                style={styles.cancelBtn}
                onPress={() => {
                  setShowPreApprove(false);
                  setPreApproveName('');
                  setPreApprovePurpose('');
                }}
              >
                <Text style={styles.cancelBtnText}>Cancel</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[
                  styles.confirmBtn,
                  preApproveMutation.isPending && { opacity: 0.6 },
                ]}
                onPress={handlePreApproveSubmit}
                disabled={preApproveMutation.isPending}
              >
                {preApproveMutation.isPending ? (
                  <ActivityIndicator size="small" color={colors.textInverse} />
                ) : (
                  <Text style={styles.confirmBtnText}>Pre-approve</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </View>
  );
}

function VisitorStatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    pending: colors.statusOpen,
    approved: colors.statusInProgress,
    checked_in: colors.success,
    checked_out: colors.textTertiary,
    denied: colors.error,
  };
  const color = colorMap[status] || colors.textTertiary;
  return (
    <View style={[vstyles.badge, { backgroundColor: color + '20' }]}>
      <Text style={[vstyles.badgeText, { color }]}>
        {status.replace('_', ' ').toUpperCase()}
      </Text>
    </View>
  );
}

const vstyles = StyleSheet.create({
  badge: {
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
    borderRadius: borderRadius.sm,
  },
  badgeText: {
    fontSize: 10,
    fontWeight: '700',
  },
});

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  loader: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  list: { padding: spacing.lg, paddingBottom: spacing.xxxl },
  preApproveButton: {
    backgroundColor: colors.primary,
    padding: spacing.md,
    borderRadius: borderRadius.md,
    alignItems: 'center',
    marginBottom: spacing.lg,
  },
  preApproveText: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: colors.textInverse,
  },
  sectionTitle: {
    fontSize: fontSize.lg,
    fontWeight: '700',
    color: colors.text,
    marginTop: spacing.lg,
    marginBottom: spacing.md,
  },
  pendingCard: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.lg,
    marginBottom: spacing.sm,
    borderLeftWidth: 3,
    borderLeftColor: colors.statusOpen,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  pendingInfo: {
    marginBottom: spacing.md,
  },
  visitorName: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: colors.text,
  },
  visitorMeta: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
    marginTop: spacing.xs,
  },
  visitorTime: {
    fontSize: fontSize.xs,
    color: colors.textTertiary,
    marginTop: spacing.xs,
  },
  actionButtons: {
    flexDirection: 'row',
    gap: spacing.sm,
  },
  approveBtn: {
    flex: 1,
    backgroundColor: colors.success + '15',
    padding: spacing.sm,
    borderRadius: borderRadius.md,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: colors.success,
  },
  approveBtnText: {
    fontSize: fontSize.sm,
    fontWeight: '700',
    color: colors.success,
  },
  denyBtn: {
    flex: 1,
    backgroundColor: colors.error + '15',
    padding: spacing.sm,
    borderRadius: borderRadius.md,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: colors.error,
  },
  denyBtnText: {
    fontSize: fontSize.sm,
    fontWeight: '700',
    color: colors.error,
  },
  visitorCard: {
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
  visitorRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  visitorDetails: { flex: 1 },
  emptyText: {
    textAlign: 'center',
    color: colors.textTertiary,
    fontSize: fontSize.md,
    marginTop: spacing.xxl,
  },

  // Modal
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: borderRadius.xl,
    borderTopRightRadius: borderRadius.xl,
    padding: spacing.xxl,
    paddingBottom: spacing.xxxl,
  },
  modalTitle: {
    fontSize: fontSize.xl,
    fontWeight: '700',
    color: colors.text,
    marginBottom: spacing.lg,
  },
  label: {
    fontSize: fontSize.sm,
    fontWeight: '700',
    color: colors.text,
    marginBottom: spacing.sm,
    marginTop: spacing.md,
  },
  input: {
    backgroundColor: colors.surfaceSecondary,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    fontSize: fontSize.md,
    color: colors.text,
  },
  chipRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
  },
  chip: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: borderRadius.full,
    backgroundColor: colors.surfaceSecondary,
    borderWidth: 1,
    borderColor: colors.border,
  },
  chipActive: {
    backgroundColor: colors.primary + '15',
    borderColor: colors.primary,
  },
  chipText: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
  },
  chipTextActive: {
    color: colors.primary,
    fontWeight: '700',
  },
  modalActions: {
    flexDirection: 'row',
    gap: spacing.md,
    marginTop: spacing.xxl,
  },
  cancelBtn: {
    flex: 1,
    padding: spacing.md,
    borderRadius: borderRadius.md,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: colors.border,
  },
  cancelBtnText: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.textSecondary,
  },
  confirmBtn: {
    flex: 1,
    backgroundColor: colors.primary,
    padding: spacing.md,
    borderRadius: borderRadius.md,
    alignItems: 'center',
  },
  confirmBtnText: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: colors.textInverse,
  },
});
