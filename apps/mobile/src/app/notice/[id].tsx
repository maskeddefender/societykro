import { View, Text, StyleSheet, ScrollView, ActivityIndicator } from 'react-native';
import { useLocalSearchParams } from 'expo-router';
import { useQuery } from '@tanstack/react-query';
import { noticeAPI } from '../../services/api';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

export default function NoticeDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();

  const { data: notice, isLoading } = useQuery({
    queryKey: ['notice', id],
    queryFn: () => noticeAPI.get(`/notices/${id}`).then((r) => r.data.data),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  if (!notice) {
    return (
      <View style={styles.center}>
        <Text style={styles.errorText}>Notice not found</Text>
      </View>
    );
  }

  return (
    <ScrollView style={styles.container}>
      <View style={styles.card}>
        {notice.is_pinned && <Text style={styles.pinnedBadge}>PINNED</Text>}
        <View style={styles.categoryBadge}>
          <Text style={styles.categoryText}>{notice.category.toUpperCase()}</Text>
        </View>
        <Text style={styles.title}>{notice.title}</Text>
        <Text style={styles.meta}>
          By {notice.created_by_name} - {new Date(notice.created_at).toLocaleDateString('en-IN', { day: 'numeric', month: 'long', year: 'numeric' })}
        </Text>

        <View style={styles.divider} />

        <Text style={styles.body}>{notice.body}</Text>

        {notice.read_count !== undefined && (
          <View style={styles.readInfo}>
            <Text style={styles.readText}>
              Read by {notice.read_count}{notice.total_members ? ` of ${notice.total_members}` : ''} members
            </Text>
          </View>
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: colors.background },
  errorText: { fontSize: fontSize.md, color: colors.textTertiary },
  card: {
    backgroundColor: colors.surface, margin: spacing.lg, borderRadius: borderRadius.lg,
    padding: spacing.xxl,
  },
  pinnedBadge: {
    fontSize: 9, fontWeight: '800', color: colors.primary, backgroundColor: colors.primary + '15',
    paddingHorizontal: spacing.sm, paddingVertical: 2, borderRadius: borderRadius.sm,
    alignSelf: 'flex-start', marginBottom: spacing.sm,
  },
  categoryBadge: {
    backgroundColor: colors.primaryLight + '15', paddingHorizontal: spacing.sm, paddingVertical: 2,
    borderRadius: borderRadius.sm, alignSelf: 'flex-start', marginBottom: spacing.md,
  },
  categoryText: { fontSize: 10, fontWeight: '700', color: colors.primaryLight },
  title: { fontSize: fontSize.xl, fontWeight: '800', color: colors.text },
  meta: { fontSize: fontSize.sm, color: colors.textTertiary, marginTop: spacing.sm },
  divider: { height: 1, backgroundColor: colors.border, marginVertical: spacing.lg },
  body: { fontSize: fontSize.md, color: colors.text, lineHeight: 24 },
  readInfo: {
    marginTop: spacing.xxl, paddingTop: spacing.lg, borderTopWidth: 1, borderTopColor: colors.border,
  },
  readText: { fontSize: fontSize.sm, color: colors.textSecondary },
});
