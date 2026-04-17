import { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useLocalSearchParams } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { complaintAPI } from '../../services/api';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

const statusColors: Record<string, string> = {
  open: colors.statusOpen,
  in_progress: colors.statusInProgress,
  resolved: colors.statusResolved,
  closed: colors.statusClosed,
};

const priorityColors: Record<string, string> = {
  normal: colors.textSecondary,
  high: colors.warning,
  emergency: colors.error,
};

export default function ComplaintDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [comment, setComment] = useState('');

  const { data: complaint, isLoading } = useQuery({
    queryKey: ['complaint', id],
    queryFn: () =>
      complaintAPI.get(`/complaints/${id}`).then((r) => r.data.data),
    enabled: !!id,
  });

  const { data: comments = [], refetch: refetchComments } = useQuery({
    queryKey: ['complaint-comments', id],
    queryFn: () =>
      complaintAPI
        .get(`/complaints/${id}/comments`)
        .then((r) => r.data.data || []),
    enabled: !!id,
  });

  const addComment = useMutation({
    mutationFn: (text: string) =>
      complaintAPI.post(`/complaints/${id}/comments`, { comment: text }),
    onSuccess: () => {
      setComment('');
      refetchComments();
      queryClient.invalidateQueries({ queryKey: ['complaint-comments', id] });
    },
  });

  if (isLoading) {
    return (
      <View style={styles.loader}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  if (!complaint) {
    return (
      <View style={styles.loader}>
        <Text style={styles.errorText}>Complaint not found</Text>
      </View>
    );
  }

  const sColor = statusColors[complaint.status] || colors.textTertiary;
  const pColor = priorityColors[complaint.priority] || colors.textSecondary;

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      keyboardVerticalOffset={90}
    >
      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.scrollContent}
      >
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.ticketNumber}>{complaint.ticket_number}</Text>
          <View style={[styles.badge, { backgroundColor: sColor + '20' }]}>
            <Text style={[styles.badgeText, { color: sColor }]}>
              {complaint.status.replace('_', ' ').toUpperCase()}
            </Text>
          </View>
        </View>

        {/* Priority */}
        <View style={[styles.priorityBadge, { backgroundColor: pColor + '15' }]}>
          <Text style={[styles.priorityText, { color: pColor }]}>
            {complaint.priority.toUpperCase()} PRIORITY
          </Text>
        </View>

        {/* Category */}
        <Text style={styles.category}>{complaint.category}</Text>

        {/* Title & Description */}
        <Text style={styles.title}>{complaint.title}</Text>
        {complaint.description ? (
          <Text style={styles.description}>{complaint.description}</Text>
        ) : null}

        {/* Meta */}
        <View style={styles.metaSection}>
          <MetaRow label="Raised by" value={complaint.raised_by_name} />
          <MetaRow
            label="Date"
            value={new Date(complaint.created_at).toLocaleDateString()}
          />
          {complaint.assigned_vendor_name && (
            <MetaRow label="Vendor" value={complaint.assigned_vendor_name} />
          )}
        </View>

        {/* Comments */}
        <Text style={styles.sectionTitle}>Comments</Text>
        {comments.length === 0 ? (
          <Text style={styles.noComments}>No comments yet</Text>
        ) : (
          comments.map((c: any) => (
            <View key={c.id} style={styles.commentCard}>
              <View style={styles.commentHeader}>
                <Text style={styles.commentAuthor}>{c.author_name}</Text>
                <Text style={styles.commentDate}>
                  {new Date(c.created_at).toLocaleDateString()}
                </Text>
              </View>
              <Text style={styles.commentText}>{c.comment}</Text>
            </View>
          ))
        )}
      </ScrollView>

      {/* Add Comment */}
      <View style={styles.commentBar}>
        <TextInput
          style={styles.commentInput}
          placeholder="Add a comment..."
          placeholderTextColor={colors.textTertiary}
          value={comment}
          onChangeText={setComment}
          multiline
        />
        <TouchableOpacity
          style={[
            styles.sendButton,
            (!comment.trim() || addComment.isPending) && styles.sendDisabled,
          ]}
          onPress={() => comment.trim() && addComment.mutate(comment.trim())}
          disabled={!comment.trim() || addComment.isPending}
        >
          {addComment.isPending ? (
            <ActivityIndicator size="small" color={colors.textInverse} />
          ) : (
            <Text style={styles.sendText}>Send</Text>
          )}
        </TouchableOpacity>
      </View>
    </KeyboardAvoidingView>
  );
}

function MetaRow({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.metaRow}>
      <Text style={styles.metaLabel}>{label}</Text>
      <Text style={styles.metaValue}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  scroll: { flex: 1 },
  scrollContent: { padding: spacing.lg, paddingBottom: spacing.xxxl },
  loader: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  errorText: { fontSize: fontSize.md, color: colors.textTertiary },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  ticketNumber: {
    fontSize: fontSize.lg,
    fontWeight: '800',
    color: colors.primary,
  },
  badge: {
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    borderRadius: borderRadius.sm,
  },
  badgeText: {
    fontSize: fontSize.xs,
    fontWeight: '700',
  },
  priorityBadge: {
    alignSelf: 'flex-start',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    borderRadius: borderRadius.sm,
    marginTop: spacing.md,
  },
  priorityText: {
    fontSize: 10,
    fontWeight: '700',
  },
  category: {
    fontSize: fontSize.sm,
    fontWeight: '600',
    color: colors.primaryLight,
    marginTop: spacing.md,
  },
  title: {
    fontSize: fontSize.xl,
    fontWeight: '700',
    color: colors.text,
    marginTop: spacing.sm,
  },
  description: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
    marginTop: spacing.md,
    lineHeight: 22,
  },
  metaSection: {
    marginTop: spacing.xxl,
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.lg,
    gap: spacing.md,
  },
  metaRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  metaLabel: {
    fontSize: fontSize.sm,
    color: colors.textTertiary,
  },
  metaValue: {
    fontSize: fontSize.sm,
    fontWeight: '600',
    color: colors.text,
  },
  sectionTitle: {
    fontSize: fontSize.lg,
    fontWeight: '700',
    color: colors.text,
    marginTop: spacing.xxl,
    marginBottom: spacing.md,
  },
  noComments: {
    fontSize: fontSize.sm,
    color: colors.textTertiary,
    fontStyle: 'italic',
  },
  commentCard: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    marginBottom: spacing.sm,
  },
  commentHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: spacing.xs,
  },
  commentAuthor: {
    fontSize: fontSize.sm,
    fontWeight: '700',
    color: colors.text,
  },
  commentDate: {
    fontSize: fontSize.xs,
    color: colors.textTertiary,
  },
  commentText: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
    lineHeight: 20,
  },
  commentBar: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    padding: spacing.md,
    backgroundColor: colors.surface,
    borderTopWidth: 1,
    borderTopColor: colors.border,
    gap: spacing.sm,
  },
  commentInput: {
    flex: 1,
    backgroundColor: colors.surfaceSecondary,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    fontSize: fontSize.sm,
    color: colors.text,
    maxHeight: 80,
  },
  sendButton: {
    backgroundColor: colors.primary,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderRadius: borderRadius.md,
  },
  sendDisabled: {
    opacity: 0.5,
  },
  sendText: {
    fontSize: fontSize.sm,
    fontWeight: '700',
    color: colors.textInverse,
  },
});
