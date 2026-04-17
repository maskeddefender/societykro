import { View, Text, StyleSheet, ScrollView, TouchableOpacity, Alert } from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '../../store/authStore';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

export default function MoreScreen() {
  const { t, i18n } = useTranslation();
  const router = useRouter();
  const { user, memberships, logout } = useAuthStore();
  const membership = memberships[0];

  const handleLogout = () => {
    Alert.alert(t('auth.logout'), 'Are you sure?', [
      { text: t('common.cancel'), style: 'cancel' },
      { text: t('auth.logout'), style: 'destructive', onPress: () => { logout(); router.replace('/'); } },
    ]);
  };

  const toggleLanguage = () => {
    const newLang = i18n.language === 'en' ? 'hi' : 'en';
    i18n.changeLanguage(newLang);
  };

  return (
    <ScrollView style={styles.container}>
      {/* Profile Card */}
      <View style={styles.profileCard}>
        <View style={styles.avatar}>
          <Text style={styles.avatarText}>{user?.name?.charAt(0) || '?'}</Text>
        </View>
        <View style={styles.profileInfo}>
          <Text style={styles.profileName}>{user?.name}</Text>
          <Text style={styles.profilePhone}>{user?.phone}</Text>
          <Text style={styles.profileRole}>{membership?.role?.toUpperCase() || 'MEMBER'}</Text>
        </View>
      </View>

      {/* Menu Items */}
      <Text style={styles.sectionTitle}>{t('profile.title')}</Text>

      <MenuItem label={t('profile.language')} value={i18n.language === 'en' ? 'English' : 'Hindi'} onPress={toggleLanguage} />
      <MenuItem label={t('profile.society')} value={membership ? 'Member' : 'Not joined'} />
      <MenuItem label={t('profile.role')} value={membership?.role || 'N/A'} />

      <Text style={styles.sectionTitle}>Emergency</Text>
      <MenuItem label="SOS Emergency" value="Tap for help" onPress={() => Alert.alert('SOS', 'Emergency alert would be sent to all guards and admins.')} danger />
      <MenuItem label="Emergency Contacts" value="Hospital, Fire, Police" />

      <Text style={styles.sectionTitle}>About</Text>
      <MenuItem label="App Version" value="0.1.0" />
      <MenuItem label="Help & Support" value="" />

      {/* Logout */}
      <TouchableOpacity style={styles.logoutButton} onPress={handleLogout}>
        <Text style={styles.logoutText}>{t('auth.logout')}</Text>
      </TouchableOpacity>

      <View style={{ height: spacing.xxxl * 2 }} />
    </ScrollView>
  );
}

function MenuItem({ label, value, onPress, danger }: { label: string; value: string; onPress?: () => void; danger?: boolean }) {
  return (
    <TouchableOpacity style={styles.menuItem} onPress={onPress} disabled={!onPress}>
      <Text style={[styles.menuLabel, danger && { color: colors.error }]}>{label}</Text>
      <Text style={styles.menuValue}>{value}</Text>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  profileCard: {
    backgroundColor: colors.primary, padding: spacing.xxl, flexDirection: 'row', alignItems: 'center',
  },
  avatar: {
    width: 56, height: 56, borderRadius: 28, backgroundColor: colors.textInverse + '20',
    justifyContent: 'center', alignItems: 'center',
  },
  avatarText: { fontSize: fontSize.xxl, fontWeight: '800', color: colors.textInverse },
  profileInfo: { marginLeft: spacing.lg, flex: 1 },
  profileName: { fontSize: fontSize.xl, fontWeight: '700', color: colors.textInverse },
  profilePhone: { fontSize: fontSize.sm, color: colors.textInverse, opacity: 0.8, marginTop: 2 },
  profileRole: {
    fontSize: 10, fontWeight: '800', color: colors.primary,
    backgroundColor: colors.textInverse, paddingHorizontal: spacing.sm, paddingVertical: 2,
    borderRadius: borderRadius.sm, alignSelf: 'flex-start', marginTop: spacing.xs,
  },
  sectionTitle: {
    fontSize: fontSize.sm, fontWeight: '700', color: colors.textSecondary, textTransform: 'uppercase',
    letterSpacing: 1, paddingHorizontal: spacing.lg, marginTop: spacing.xxl, marginBottom: spacing.sm,
  },
  menuItem: {
    backgroundColor: colors.surface, paddingHorizontal: spacing.lg, paddingVertical: spacing.lg,
    flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    borderBottomWidth: 1, borderBottomColor: colors.borderLight,
  },
  menuLabel: { fontSize: fontSize.md, color: colors.text, fontWeight: '500' },
  menuValue: { fontSize: fontSize.sm, color: colors.textTertiary },
  logoutButton: {
    marginHorizontal: spacing.lg, marginTop: spacing.xxxl, paddingVertical: spacing.lg,
    borderRadius: borderRadius.md, borderWidth: 1.5, borderColor: colors.error, alignItems: 'center',
  },
  logoutText: { fontSize: fontSize.md, fontWeight: '700', color: colors.error },
});
