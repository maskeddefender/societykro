import { Tabs } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { colors } from '../../constants/theme';

export default function TabLayout() {
  const { t } = useTranslation();

  return (
    <Tabs
      screenOptions={{
        headerStyle: { backgroundColor: colors.primary },
        headerTintColor: colors.textInverse,
        headerTitleStyle: { fontWeight: '700' },
        tabBarActiveTintColor: colors.primary,
        tabBarInactiveTintColor: colors.textTertiary,
        tabBarStyle: {
          backgroundColor: colors.surface,
          borderTopColor: colors.border,
          paddingBottom: 4,
          height: 56,
        },
        tabBarLabelStyle: { fontSize: 11, fontWeight: '600' },
      }}
    >
      <Tabs.Screen
        name="home"
        options={{
          title: t('tabs.home'),
          headerTitle: 'SocietyKro',
          tabBarIcon: ({ color }) => <TabIcon name="home" color={color} />,
        }}
      />
      <Tabs.Screen
        name="complaints"
        options={{
          title: t('tabs.complaints'),
          tabBarIcon: ({ color }) => <TabIcon name="alert" color={color} />,
        }}
      />
      <Tabs.Screen
        name="visitors"
        options={{
          title: t('tabs.visitors'),
          tabBarIcon: ({ color }) => <TabIcon name="people" color={color} />,
        }}
      />
      <Tabs.Screen
        name="payments"
        options={{
          title: t('tabs.payments'),
          tabBarIcon: ({ color }) => <TabIcon name="card" color={color} />,
        }}
      />
      <Tabs.Screen
        name="more"
        options={{
          title: t('tabs.more'),
          tabBarIcon: ({ color }) => <TabIcon name="menu" color={color} />,
        }}
      />
    </Tabs>
  );
}

// Simple text-based icons (replace with expo-vector-icons if installed)
function TabIcon({ name, color }: { name: string; color: string }) {
  const icons: Record<string, string> = {
    home: '\u2302',
    alert: '\u26A0',
    people: '\u263A',
    card: '\u2B50',
    menu: '\u2630',
  };
  const { Text } = require('react-native');
  return <Text style={{ fontSize: 20, color }}>{icons[name] || '?'}</Text>;
}
