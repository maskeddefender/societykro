import { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TextInput,
  TouchableOpacity,
  ScrollView,
  Alert,
  ActivityIndicator,
  Image,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import * as ImagePicker from 'expo-image-picker';
import { complaintAPI } from '../../services/api';
import { colors, spacing, fontSize, borderRadius } from '../../constants/theme';

const CATEGORIES = [
  'water',
  'electrical',
  'lift',
  'plumbing',
  'security',
  'parking',
  'garbage',
  'pest_control',
  'drain',
  'generator',
  'intercom',
  'other',
] as const;

const PRIORITIES = ['normal', 'high', 'emergency'] as const;

const CATEGORY_LABELS: Record<string, string> = {
  water: 'Water',
  electrical: 'Electrical',
  lift: 'Lift',
  plumbing: 'Plumbing',
  security: 'Security',
  parking: 'Parking',
  garbage: 'Garbage',
  pest_control: 'Pest Control',
  drain: 'Drain',
  generator: 'Generator',
  intercom: 'Intercom',
  other: 'Other',
};

const PRIORITY_COLORS: Record<string, string> = {
  normal: colors.textSecondary,
  high: colors.warning,
  emergency: colors.error,
};

export default function NewComplaintScreen() {
  const { t } = useTranslation();
  const router = useRouter();
  const queryClient = useQueryClient();

  const [category, setCategory] = useState('');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState<string>('normal');
  const [isEmergency, setIsEmergency] = useState(false);
  const [imageUri, setImageUri] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: (payload: any) => complaintAPI.post('/complaints', payload),
    onSuccess: (res) => {
      const ticket = res.data?.data?.ticket_number || 'N/A';
      queryClient.invalidateQueries({ queryKey: ['complaints'] });
      Alert.alert('Complaint Raised', `Ticket: ${ticket}`, [
        { text: 'OK', onPress: () => router.back() },
      ]);
    },
    onError: () => {
      Alert.alert('Error', 'Failed to submit complaint. Please try again.');
    },
  });

  const handlePickImage = async () => {
    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      quality: 0.8,
    });
    if (!result.canceled && result.assets[0]) {
      setImageUri(result.assets[0].uri);
    }
  };

  const handleSubmit = () => {
    if (!category) {
      Alert.alert('Required', 'Please select a category.');
      return;
    }
    if (!title.trim()) {
      Alert.alert('Required', 'Please enter a title.');
      return;
    }

    mutation.mutate({
      category,
      title: title.trim(),
      description: description.trim(),
      priority,
      is_emergency: isEmergency,
      image_urls: [],
    });
  };

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.content}
      keyboardShouldPersistTaps="handled"
    >
      {/* Category */}
      <Text style={styles.label}>Category *</Text>
      <View style={styles.chipRow}>
        {CATEGORIES.map((cat) => (
          <TouchableOpacity
            key={cat}
            style={[styles.chip, category === cat && styles.chipActive]}
            onPress={() => setCategory(cat)}
          >
            <Text
              style={[
                styles.chipText,
                category === cat && styles.chipTextActive,
              ]}
            >
              {CATEGORY_LABELS[cat]}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      {/* Title */}
      <Text style={styles.label}>Title *</Text>
      <TextInput
        style={styles.input}
        placeholder="Brief summary of the issue"
        placeholderTextColor={colors.textTertiary}
        value={title}
        onChangeText={setTitle}
        maxLength={120}
      />

      {/* Description */}
      <Text style={styles.label}>Description</Text>
      <TextInput
        style={[styles.input, styles.textArea]}
        placeholder="Describe the issue in detail..."
        placeholderTextColor={colors.textTertiary}
        value={description}
        onChangeText={setDescription}
        multiline
        numberOfLines={4}
        textAlignVertical="top"
      />

      {/* Priority */}
      <Text style={styles.label}>Priority</Text>
      <View style={styles.chipRow}>
        {PRIORITIES.map((p) => (
          <TouchableOpacity
            key={p}
            style={[
              styles.priorityChip,
              priority === p && {
                backgroundColor: PRIORITY_COLORS[p] + '20',
                borderColor: PRIORITY_COLORS[p],
              },
            ]}
            onPress={() => {
              setPriority(p);
              if (p === 'emergency') {
                setIsEmergency(true);
              } else {
                setIsEmergency(false);
              }
            }}
          >
            <Text
              style={[
                styles.priorityText,
                priority === p && { color: PRIORITY_COLORS[p] },
              ]}
            >
              {p.charAt(0).toUpperCase() + p.slice(1)}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      {/* Emergency toggle */}
      <TouchableOpacity
        style={styles.emergencyRow}
        onPress={() => setIsEmergency(!isEmergency)}
      >
        <View
          style={[
            styles.checkbox,
            isEmergency && styles.checkboxActive,
          ]}
        >
          {isEmergency && <Text style={styles.checkmark}>&#10003;</Text>}
        </View>
        <Text style={styles.emergencyLabel}>Mark as Emergency</Text>
      </TouchableOpacity>

      {/* Photo */}
      <TouchableOpacity style={styles.photoButton} onPress={handlePickImage}>
        <Text style={styles.photoButtonText}>
          {imageUri ? 'Change Photo' : 'Add Photo'}
        </Text>
      </TouchableOpacity>
      {imageUri && (
        <Image source={{ uri: imageUri }} style={styles.preview} />
      )}

      {/* Submit */}
      <TouchableOpacity
        style={[styles.submitButton, mutation.isPending && styles.submitDisabled]}
        onPress={handleSubmit}
        disabled={mutation.isPending}
        activeOpacity={0.8}
      >
        {mutation.isPending ? (
          <ActivityIndicator color={colors.textInverse} />
        ) : (
          <Text style={styles.submitText}>Submit Complaint</Text>
        )}
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  content: { padding: spacing.lg, paddingBottom: spacing.xxxl },
  label: {
    fontSize: fontSize.sm,
    fontWeight: '700',
    color: colors.text,
    marginBottom: spacing.sm,
    marginTop: spacing.lg,
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
  input: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    fontSize: fontSize.md,
    color: colors.text,
  },
  textArea: {
    minHeight: 100,
  },
  priorityChip: {
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    borderRadius: borderRadius.full,
    backgroundColor: colors.surfaceSecondary,
    borderWidth: 1,
    borderColor: colors.border,
  },
  priorityText: {
    fontSize: fontSize.sm,
    fontWeight: '600',
    color: colors.textSecondary,
  },
  emergencyRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: spacing.lg,
    gap: spacing.sm,
  },
  checkbox: {
    width: 22,
    height: 22,
    borderRadius: borderRadius.sm,
    borderWidth: 2,
    borderColor: colors.border,
    justifyContent: 'center',
    alignItems: 'center',
  },
  checkboxActive: {
    backgroundColor: colors.error,
    borderColor: colors.error,
  },
  checkmark: {
    color: colors.textInverse,
    fontSize: 14,
    fontWeight: '700',
  },
  emergencyLabel: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.error,
  },
  photoButton: {
    marginTop: spacing.lg,
    padding: spacing.md,
    borderRadius: borderRadius.md,
    borderWidth: 1,
    borderColor: colors.primaryLight,
    borderStyle: 'dashed',
    alignItems: 'center',
  },
  photoButtonText: {
    fontSize: fontSize.sm,
    fontWeight: '600',
    color: colors.primaryLight,
  },
  preview: {
    width: '100%',
    height: 200,
    borderRadius: borderRadius.md,
    marginTop: spacing.md,
  },
  submitButton: {
    backgroundColor: colors.primary,
    padding: spacing.lg,
    borderRadius: borderRadius.md,
    alignItems: 'center',
    marginTop: spacing.xxl,
  },
  submitDisabled: {
    opacity: 0.6,
  },
  submitText: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: colors.textInverse,
  },
});
