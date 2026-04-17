import { Redirect } from 'expo-router';
import { ActivityIndicator, View } from 'react-native';
import { useGuardStore } from '../store/guardStore';
import { colors } from '../constants/theme';

export default function Index() {
  const { isAuthenticated, isLoading } = useGuardStore();

  if (isLoading) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: colors.background }}>
        <ActivityIndicator size="large" color={colors.accent} />
      </View>
    );
  }

  return <Redirect href={isAuthenticated ? '/dashboard' : '/login'} />;
}
