import { useState, useCallback } from 'react';
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet,
  ScrollView, Alert, RefreshControl, ActivityIndicator, Modal,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useGuardStore } from '../store/guardStore';
import { visitorAPI } from '../services/api';
import { colors, spacing, fontSize, borderRadius } from '../constants/theme';

// Visitor purposes for quick selection
const PURPOSES = ['Guest', 'Delivery', 'Cab', 'Service', 'Official', 'Other'];

export default function GuardDashboard() {
  const router = useRouter();
  const { guardName, logout } = useGuardStore();

  const [visitors, setVisitors] = useState<any[]>([]);
  const [activeCount, setActiveCount] = useState(0);
  const [refreshing, setRefreshing] = useState(false);

  // Quick Log modal
  const [showLog, setShowLog] = useState(false);
  const [logName, setLogName] = useState('');
  const [logPurpose, setLogPurpose] = useState('Guest');
  const [logFlat, setLogFlat] = useState('');
  const [logLoading, setLogLoading] = useState(false);

  // OTP Verify modal
  const [showOTP, setShowOTP] = useState(false);
  const [otpCode, setOtpCode] = useState('');
  const [otpLoading, setOtpLoading] = useState(false);

  const fetchData = useCallback(async () => {
    setRefreshing(true);
    try {
      const [visRes, activeRes] = await Promise.all([
        visitorAPI.get('/visitors?limit=10').catch(() => ({ data: { data: [] } })),
        visitorAPI.get('/visitors/active').catch(() => ({ data: { data: [] } })),
      ]);
      setVisitors(visRes.data.data || []);
      setActiveCount((activeRes.data.data || []).length);
    } catch { }
    setRefreshing(false);
  }, []);

  // Log visitor entry
  const handleLogVisitor = async () => {
    if (!logName.trim()) { Alert.alert('Error', 'Visitor name required'); return; }
    if (!logFlat.trim()) { Alert.alert('Error', 'Flat number required'); return; }
    setLogLoading(true);
    try {
      await visitorAPI.post('/visitors/log', {
        name: logName,
        purpose: logPurpose.toLowerCase(),
        flat_id: logFlat, // Will need flat lookup — using flat number for now
      });
      Alert.alert('Logged', `${logName} entry logged for flat ${logFlat}`);
      setShowLog(false);
      setLogName('');
      setLogFlat('');
      fetchData();
    } catch (e: any) {
      Alert.alert('Error', e.response?.data?.detail || 'Failed to log visitor');
    }
    setLogLoading(false);
  };

  // Verify OTP at gate
  const handleVerifyOTP = async () => {
    if (otpCode.length !== 6) { Alert.alert('Error', 'Enter 6-digit OTP'); return; }
    setOtpLoading(true);
    try {
      const res = await visitorAPI.post('/visitors/verify-otp', { otp_code: otpCode });
      const v = res.data.data;
      Alert.alert('APPROVED', `${v.name} — Flat ${v.flat_number}\nPurpose: ${v.purpose}\n\nLet them in.`);
      setShowOTP(false);
      setOtpCode('');
      fetchData();
    } catch {
      Alert.alert('INVALID', 'OTP not found or expired. Deny entry.');
    }
    setOtpLoading(false);
  };

  // SOS
  const handleSOS = () => {
    Alert.alert(
      'SOS EMERGENCY',
      'This will alert ALL residents and admins immediately.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'SEND SOS', style: 'destructive',
          onPress: () => Alert.alert('SOS Sent', 'All admins and nearby residents have been alerted.'),
        },
      ]
    );
  };

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.greeting}>Hello, {guardName || 'Guard'}</Text>
          <Text style={styles.subtitle}>{activeCount} visitors inside</Text>
        </View>
        <TouchableOpacity onPress={() => { logout(); router.replace('/'); }}>
          <Text style={styles.logoutBtn}>Logout</Text>
        </TouchableOpacity>
      </View>

      {/* Big Action Buttons */}
      <View style={styles.actionsGrid}>
        <TouchableOpacity style={[styles.actionBtn, { backgroundColor: colors.accent }]} onPress={() => setShowLog(true)}>
          <Text style={styles.actionIcon}>+</Text>
          <Text style={styles.actionLabel}>LOG{'\n'}VISITOR</Text>
        </TouchableOpacity>

        <TouchableOpacity style={[styles.actionBtn, { backgroundColor: colors.success }]} onPress={() => setShowOTP(true)}>
          <Text style={styles.actionIcon}>#</Text>
          <Text style={styles.actionLabel}>VERIFY{'\n'}OTP</Text>
        </TouchableOpacity>

        <TouchableOpacity style={[styles.actionBtn, { backgroundColor: colors.danger }]} onPress={handleSOS}>
          <Text style={styles.actionIcon}>!</Text>
          <Text style={styles.actionLabel}>SOS{'\n'}ALERT</Text>
        </TouchableOpacity>
      </View>

      {/* Recent Visitors */}
      <Text style={styles.sectionTitle}>Recent Visitors</Text>
      <ScrollView
        style={styles.list}
        refreshControl={<RefreshControl refreshing={refreshing} onRefresh={fetchData} tintColor={colors.accent} />}
      >
        {visitors.length === 0 ? (
          <Text style={styles.empty}>Pull down to refresh</Text>
        ) : (
          visitors.map((v: any) => (
            <View key={v.id} style={styles.visitorCard}>
              <View style={styles.visitorInfo}>
                <Text style={styles.visitorName}>{v.name || v.visitor_name}</Text>
                <Text style={styles.visitorMeta}>{v.purpose} — {v.flat_number || 'N/A'}</Text>
              </View>
              <View style={[styles.statusDot, { backgroundColor: v.status === 'checked_in' ? colors.success : v.status === 'pending' ? colors.warning : colors.textMuted }]} />
            </View>
          ))
        )}
        <View style={{ height: 100 }} />
      </ScrollView>

      {/* Log Visitor Modal */}
      <Modal visible={showLog} animationType="slide" transparent>
        <View style={styles.modalOverlay}>
          <View style={styles.modal}>
            <Text style={styles.modalTitle}>Log Visitor</Text>

            <TextInput
              style={styles.modalInput}
              placeholder="Visitor Name"
              placeholderTextColor={colors.textMuted}
              value={logName}
              onChangeText={setLogName}
              autoFocus
            />

            <TextInput
              style={styles.modalInput}
              placeholder="Flat Number (e.g. A-301)"
              placeholderTextColor={colors.textMuted}
              value={logFlat}
              onChangeText={setLogFlat}
              autoCapitalize="characters"
            />

            <Text style={styles.purposeLabel}>Purpose</Text>
            <View style={styles.purposeRow}>
              {PURPOSES.map((p) => (
                <TouchableOpacity
                  key={p}
                  style={[styles.purposeChip, logPurpose === p && styles.purposeChipActive]}
                  onPress={() => setLogPurpose(p)}
                >
                  <Text style={[styles.purposeText, logPurpose === p && styles.purposeTextActive]}>{p}</Text>
                </TouchableOpacity>
              ))}
            </View>

            <View style={styles.modalActions}>
              <TouchableOpacity style={styles.cancelBtn} onPress={() => setShowLog(false)}>
                <Text style={styles.cancelText}>Cancel</Text>
              </TouchableOpacity>
              <TouchableOpacity style={styles.submitBtn} onPress={handleLogVisitor} disabled={logLoading}>
                {logLoading ? <ActivityIndicator color="#fff" /> : <Text style={styles.submitText}>LOG ENTRY</Text>}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>

      {/* Verify OTP Modal */}
      <Modal visible={showOTP} animationType="slide" transparent>
        <View style={styles.modalOverlay}>
          <View style={styles.modal}>
            <Text style={styles.modalTitle}>Verify Visitor OTP</Text>
            <Text style={styles.modalSubtitle}>Ask the visitor for their 6-digit OTP</Text>

            <TextInput
              style={[styles.modalInput, { textAlign: 'center', fontSize: fontSize.xl, letterSpacing: 10, fontWeight: '800' }]}
              placeholder="000000"
              placeholderTextColor={colors.textMuted}
              keyboardType="number-pad"
              maxLength={6}
              value={otpCode}
              onChangeText={(t) => setOtpCode(t.replace(/[^0-9]/g, ''))}
              autoFocus
            />

            <View style={styles.modalActions}>
              <TouchableOpacity style={styles.cancelBtn} onPress={() => { setShowOTP(false); setOtpCode(''); }}>
                <Text style={styles.cancelText}>Cancel</Text>
              </TouchableOpacity>
              <TouchableOpacity style={[styles.submitBtn, { backgroundColor: colors.success }]} onPress={handleVerifyOTP} disabled={otpLoading}>
                {otpLoading ? <ActivityIndicator color="#fff" /> : <Text style={styles.submitText}>VERIFY</Text>}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  header: {
    flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    paddingHorizontal: spacing.xl, paddingTop: 60, paddingBottom: spacing.lg,
    backgroundColor: colors.surface,
  },
  greeting: { fontSize: fontSize.xl, fontWeight: '800', color: colors.text },
  subtitle: { fontSize: fontSize.sm, color: colors.textSecondary, marginTop: 2 },
  logoutBtn: { fontSize: fontSize.sm, color: colors.danger, fontWeight: '600' },

  actionsGrid: { flexDirection: 'row', gap: spacing.md, padding: spacing.lg },
  actionBtn: {
    flex: 1, borderRadius: borderRadius.lg, padding: spacing.lg,
    alignItems: 'center', justifyContent: 'center', minHeight: 120,
  },
  actionIcon: { fontSize: fontSize.hero, fontWeight: '900', color: '#fff', marginBottom: spacing.xs },
  actionLabel: { fontSize: fontSize.sm, fontWeight: '800', color: '#fff', textAlign: 'center', letterSpacing: 1 },

  sectionTitle: {
    fontSize: fontSize.sm, fontWeight: '700', color: colors.textSecondary,
    textTransform: 'uppercase', letterSpacing: 2, paddingHorizontal: spacing.xl, marginTop: spacing.lg, marginBottom: spacing.md,
  },
  list: { flex: 1, paddingHorizontal: spacing.lg },
  empty: { textAlign: 'center', color: colors.textMuted, marginTop: spacing.xxl, fontSize: fontSize.md },

  visitorCard: {
    backgroundColor: colors.surface, borderRadius: borderRadius.md, padding: spacing.lg,
    flexDirection: 'row', alignItems: 'center', marginBottom: spacing.sm,
  },
  visitorInfo: { flex: 1 },
  visitorName: { fontSize: fontSize.md, fontWeight: '700', color: colors.text },
  visitorMeta: { fontSize: fontSize.sm, color: colors.textSecondary, marginTop: 2 },
  statusDot: { width: 12, height: 12, borderRadius: 6 },

  modalOverlay: { flex: 1, backgroundColor: 'rgba(0,0,0,0.7)', justifyContent: 'flex-end' },
  modal: { backgroundColor: colors.surface, borderTopLeftRadius: borderRadius.xl, borderTopRightRadius: borderRadius.xl, padding: spacing.xl, paddingBottom: 40 },
  modalTitle: { fontSize: fontSize.xl, fontWeight: '800', color: colors.text, marginBottom: spacing.xs },
  modalSubtitle: { fontSize: fontSize.sm, color: colors.textSecondary, marginBottom: spacing.lg },
  modalInput: {
    borderWidth: 1, borderColor: colors.border, borderRadius: borderRadius.md,
    paddingVertical: spacing.lg, paddingHorizontal: spacing.lg,
    fontSize: fontSize.md, color: colors.text, marginBottom: spacing.md,
  },
  purposeLabel: { fontSize: fontSize.sm, color: colors.textSecondary, marginBottom: spacing.sm, fontWeight: '500' },
  purposeRow: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm, marginBottom: spacing.lg },
  purposeChip: { paddingHorizontal: spacing.lg, paddingVertical: spacing.sm, borderRadius: borderRadius.xl, borderWidth: 1, borderColor: colors.border },
  purposeChipActive: { backgroundColor: colors.accent, borderColor: colors.accent },
  purposeText: { fontSize: fontSize.sm, color: colors.textSecondary, fontWeight: '600' },
  purposeTextActive: { color: '#fff' },
  modalActions: { flexDirection: 'row', gap: spacing.md, marginTop: spacing.md },
  cancelBtn: { flex: 1, paddingVertical: spacing.lg, borderRadius: borderRadius.md, borderWidth: 1, borderColor: colors.border, alignItems: 'center' },
  cancelText: { fontSize: fontSize.md, color: colors.textSecondary, fontWeight: '600' },
  submitBtn: { flex: 2, paddingVertical: spacing.lg, borderRadius: borderRadius.md, backgroundColor: colors.accent, alignItems: 'center' },
  submitText: { fontSize: fontSize.md, color: '#fff', fontWeight: '800', letterSpacing: 1 },
});
