import { useState, useRef } from 'react';
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet,
  KeyboardAvoidingView, Platform, Alert, ActivityIndicator,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '../../store/authStore';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

export default function LoginScreen() {
  const { t } = useTranslation();
  const router = useRouter();
  const { sendOTP, verifyOTP } = useAuthStore();

  const [phone, setPhone] = useState('');
  const [otp, setOtp] = useState('');
  const [step, setStep] = useState<'phone' | 'otp'>('phone');
  const [loading, setLoading] = useState(false);
  const otpRef = useRef<TextInput>(null);

  const handleSendOTP = async () => {
    const formatted = phone.startsWith('+91') ? phone : `+91${phone.replace(/\s/g, '')}`;
    if (formatted.length < 13) {
      Alert.alert('Error', 'Enter a valid 10-digit phone number');
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
      Alert.alert('Error', t('auth.tooManyAttempts'));
    }
  };

  const handleVerifyOTP = async () => {
    if (otp.length !== 6) {
      Alert.alert('Error', 'Enter 6-digit OTP');
      return;
    }

    setLoading(true);
    const ok = await verifyOTP(phone, otp);
    setLoading(false);

    if (ok) {
      router.replace('/(tabs)/home');
    } else {
      Alert.alert('Error', t('auth.invalidOTP'));
      setOtp('');
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <View style={styles.header}>
        <Text style={styles.logo}>SocietyKro</Text>
        <Text style={styles.tagline}>{t('auth.tagline')}</Text>
      </View>

      <View style={styles.card}>
        {step === 'phone' ? (
          <>
            <Text style={styles.label}>{t('auth.phoneLabel')}</Text>
            <View style={styles.phoneRow}>
              <Text style={styles.countryCode}>+91</Text>
              <TextInput
                style={styles.phoneInput}
                placeholder="98765 43210"
                placeholderTextColor={colors.textTertiary}
                keyboardType="phone-pad"
                maxLength={10}
                value={phone.replace('+91', '')}
                onChangeText={(t) => setPhone(t.replace(/[^0-9]/g, ''))}
                autoFocus
              />
            </View>
            <TouchableOpacity
              style={[styles.button, loading && styles.buttonDisabled]}
              onPress={handleSendOTP}
              disabled={loading}
            >
              {loading ? (
                <ActivityIndicator color={colors.textInverse} />
              ) : (
                <Text style={styles.buttonText}>{t('auth.sendOTP')}</Text>
              )}
            </TouchableOpacity>
          </>
        ) : (
          <>
            <Text style={styles.label}>{t('auth.otpSent', { phone })}</Text>
            <TextInput
              ref={otpRef}
              style={styles.otpInput}
              placeholder="000000"
              placeholderTextColor={colors.textTertiary}
              keyboardType="number-pad"
              maxLength={6}
              value={otp}
              onChangeText={(t) => setOtp(t.replace(/[^0-9]/g, ''))}
              textAlign="center"
            />
            <TouchableOpacity
              style={[styles.button, loading && styles.buttonDisabled]}
              onPress={handleVerifyOTP}
              disabled={loading}
            >
              {loading ? (
                <ActivityIndicator color={colors.textInverse} />
              ) : (
                <Text style={styles.buttonText}>{t('auth.verifyOTP')}</Text>
              )}
            </TouchableOpacity>
            <TouchableOpacity
              style={styles.linkButton}
              onPress={() => { setStep('phone'); setOtp(''); }}
            >
              <Text style={styles.linkText}>Change phone number</Text>
            </TouchableOpacity>
          </>
        )}
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.primary,
    justifyContent: 'center',
    padding: spacing.xl,
  },
  header: {
    alignItems: 'center',
    marginBottom: spacing.xxxl,
  },
  logo: {
    fontSize: fontSize.xxxl,
    fontWeight: '800',
    color: colors.textInverse,
    letterSpacing: 1,
  },
  tagline: {
    fontSize: fontSize.md,
    color: colors.textInverse,
    opacity: 0.8,
    marginTop: spacing.sm,
  },
  card: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.xl,
    padding: spacing.xxl,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.15,
    shadowRadius: 12,
    elevation: 8,
  },
  label: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.text,
    marginBottom: spacing.md,
  },
  phoneRow: {
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1.5,
    borderColor: colors.border,
    borderRadius: borderRadius.md,
    marginBottom: spacing.lg,
  },
  countryCode: {
    paddingHorizontal: spacing.lg,
    fontSize: fontSize.lg,
    fontWeight: '600',
    color: colors.text,
    borderRightWidth: 1,
    borderRightColor: colors.border,
    paddingVertical: spacing.lg,
  },
  phoneInput: {
    flex: 1,
    paddingHorizontal: spacing.lg,
    fontSize: fontSize.lg,
    color: colors.text,
    paddingVertical: spacing.lg,
    letterSpacing: 2,
  },
  otpInput: {
    borderWidth: 1.5,
    borderColor: colors.border,
    borderRadius: borderRadius.md,
    paddingVertical: spacing.lg,
    fontSize: fontSize.xxl,
    fontWeight: '700',
    color: colors.text,
    letterSpacing: 12,
    marginBottom: spacing.lg,
  },
  button: {
    backgroundColor: colors.primary,
    borderRadius: borderRadius.md,
    paddingVertical: spacing.lg,
    alignItems: 'center',
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  buttonText: {
    fontSize: fontSize.lg,
    fontWeight: '700',
    color: colors.textInverse,
  },
  linkButton: {
    alignItems: 'center',
    marginTop: spacing.lg,
  },
  linkText: {
    fontSize: fontSize.sm,
    color: colors.primaryLight,
    fontWeight: '500',
  },
});
