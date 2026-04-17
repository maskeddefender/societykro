import { useState, useRef } from 'react';
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet,
  KeyboardAvoidingView, Platform, Alert, ActivityIndicator,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useGuardStore } from '../store/guardStore';
import { colors, spacing, fontSize, borderRadius } from '../constants/theme';

export default function LoginScreen() {
  const router = useRouter();
  const { sendOTP, verifyOTP } = useGuardStore();
  const [phone, setPhone] = useState('');
  const [otp, setOtp] = useState('');
  const [step, setStep] = useState<'phone' | 'otp'>('phone');
  const [loading, setLoading] = useState(false);
  const otpRef = useRef<TextInput>(null);

  const handleSendOTP = async () => {
    const formatted = phone.startsWith('+91') ? phone : `+91${phone.replace(/\s/g, '')}`;
    if (formatted.length < 13) {
      Alert.alert('Error', 'Enter valid 10-digit phone number');
      return;
    }
    setLoading(true);
    const ok = await sendOTP(formatted);
    setLoading(false);
    if (ok) {
      setPhone(formatted);
      setStep('otp');
      setTimeout(() => otpRef.current?.focus(), 300);
    } else {
      Alert.alert('Error', 'Failed to send OTP');
    }
  };

  const handleVerifyOTP = async () => {
    if (otp.length !== 6) { Alert.alert('Error', 'Enter 6-digit OTP'); return; }
    setLoading(true);
    const ok = await verifyOTP(phone, otp);
    setLoading(false);
    if (ok) {
      router.replace('/dashboard');
    } else {
      Alert.alert('Access Denied', 'Only guard accounts can use this app.');
      setOtp('');
    }
  };

  return (
    <KeyboardAvoidingView style={styles.container} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
      <View style={styles.header}>
        <Text style={styles.shield}>GUARD</Text>
        <Text style={styles.logo}>SocietyKro</Text>
        <Text style={styles.subtitle}>Security Guard App</Text>
      </View>

      <View style={styles.card}>
        {step === 'phone' ? (
          <>
            <Text style={styles.label}>Phone Number</Text>
            <View style={styles.phoneRow}>
              <Text style={styles.prefix}>+91</Text>
              <TextInput
                style={styles.phoneInput}
                placeholder="98765 43210"
                placeholderTextColor={colors.textMuted}
                keyboardType="phone-pad"
                maxLength={10}
                value={phone.replace('+91', '')}
                onChangeText={(t) => setPhone(t.replace(/[^0-9]/g, ''))}
                autoFocus
              />
            </View>
            <TouchableOpacity style={[styles.btn, loading && styles.btnDisabled]} onPress={handleSendOTP} disabled={loading}>
              {loading ? <ActivityIndicator color="#fff" /> : <Text style={styles.btnText}>SEND OTP</Text>}
            </TouchableOpacity>
          </>
        ) : (
          <>
            <Text style={styles.label}>Enter OTP sent to {phone}</Text>
            <TextInput
              ref={otpRef}
              style={styles.otpInput}
              placeholder="000000"
              placeholderTextColor={colors.textMuted}
              keyboardType="number-pad"
              maxLength={6}
              value={otp}
              onChangeText={(t) => setOtp(t.replace(/[^0-9]/g, ''))}
              textAlign="center"
            />
            <TouchableOpacity style={[styles.btn, loading && styles.btnDisabled]} onPress={handleVerifyOTP} disabled={loading}>
              {loading ? <ActivityIndicator color="#fff" /> : <Text style={styles.btnText}>VERIFY</Text>}
            </TouchableOpacity>
            <TouchableOpacity style={styles.link} onPress={() => { setStep('phone'); setOtp(''); }}>
              <Text style={styles.linkText}>Change number</Text>
            </TouchableOpacity>
          </>
        )}
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background, justifyContent: 'center', padding: spacing.xl },
  header: { alignItems: 'center', marginBottom: spacing.xxl },
  shield: {
    fontSize: fontSize.sm, fontWeight: '800', color: colors.accent, letterSpacing: 4,
    backgroundColor: colors.surface, paddingHorizontal: spacing.lg, paddingVertical: spacing.xs,
    borderRadius: borderRadius.sm, marginBottom: spacing.md, overflow: 'hidden',
  },
  logo: { fontSize: fontSize.xxl, fontWeight: '800', color: colors.text },
  subtitle: { fontSize: fontSize.md, color: colors.textSecondary, marginTop: spacing.xs },
  card: { backgroundColor: colors.surface, borderRadius: borderRadius.lg, padding: spacing.xl },
  label: { fontSize: fontSize.sm, color: colors.textSecondary, marginBottom: spacing.md, fontWeight: '500' },
  phoneRow: { flexDirection: 'row', borderWidth: 1, borderColor: colors.border, borderRadius: borderRadius.md, marginBottom: spacing.lg },
  prefix: { paddingHorizontal: spacing.lg, paddingVertical: spacing.lg, fontSize: fontSize.lg, fontWeight: '700', color: colors.text, borderRightWidth: 1, borderRightColor: colors.border },
  phoneInput: { flex: 1, paddingHorizontal: spacing.lg, fontSize: fontSize.lg, color: colors.text, letterSpacing: 2 },
  otpInput: { borderWidth: 1, borderColor: colors.border, borderRadius: borderRadius.md, paddingVertical: spacing.lg, fontSize: fontSize.xl, fontWeight: '800', color: colors.text, letterSpacing: 12, marginBottom: spacing.lg },
  btn: { backgroundColor: colors.accent, borderRadius: borderRadius.md, paddingVertical: spacing.lg, alignItems: 'center' },
  btnDisabled: { opacity: 0.5 },
  btnText: { fontSize: fontSize.md, fontWeight: '800', color: '#fff', letterSpacing: 1 },
  link: { alignItems: 'center', marginTop: spacing.lg },
  linkText: { fontSize: fontSize.sm, color: colors.accent },
});
