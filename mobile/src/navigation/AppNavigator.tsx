/**
 * Main App Navigation
 * Tab-based navigation with stack navigators
 */

import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createStackNavigator } from '@react-navigation/stack';
import { Ionicons } from '@expo/vector-icons';

import { DiscoveryScreen } from '../screens/DiscoveryScreen';
import { ControlScreen } from '../screens/ControlScreen';
import { MultiDeviceScreen } from '../screens/MultiDeviceScreen';
import { SettingsScreen } from '../screens/SettingsScreen';

const Tab = createBottomTabNavigator();
const Stack = createStackNavigator();

// Discovery Stack
const DiscoveryStack = () => (
  <Stack.Navigator>
    <Stack.Screen 
      name="DiscoveryMain" 
      component={DiscoveryScreen}
      options={{ title: 'Device Discovery' }}
    />
  </Stack.Navigator>
);

// Control Stack
const ControlStack = () => (
  <Stack.Navigator>
    <Stack.Screen 
      name="ControlMain" 
      component={ControlScreen}
      options={{ title: 'Device Control' }}
    />
  </Stack.Navigator>
);

// Multi-Device Stack
const MultiDeviceStack = () => (
  <Stack.Navigator>
    <Stack.Screen 
      name="MultiDeviceMain" 
      component={MultiDeviceScreen}
      options={{ title: 'Multi-Device Control' }}
    />
  </Stack.Navigator>
);

// Settings Stack
const SettingsStack = () => (
  <Stack.Navigator>
    <Stack.Screen 
      name="SettingsMain" 
      component={SettingsScreen}
      options={{ title: 'Settings' }}
    />
  </Stack.Navigator>
);

// Main Tab Navigator
export const AppNavigator: React.FC = () => {
  return (
    <NavigationContainer>
      <Tab.Navigator
        screenOptions={({ route }) => ({
          tabBarIcon: ({ focused, color, size }) => {
            let iconName: keyof typeof Ionicons.glyphMap;

            switch (route.name) {
              case 'Discovery':
                iconName = focused ? 'search' : 'search-outline';
                break;
              case 'Control':
                iconName = focused ? 'volume-high' : 'volume-high-outline';
                break;
              case 'MultiDevice':
                iconName = focused ? 'grid' : 'grid-outline';
                break;
              case 'Settings':
                iconName = focused ? 'settings' : 'settings-outline';
                break;
              default:
                iconName = 'help-outline';
            }

            return <Ionicons name={iconName} size={size} color={color} />;
          },
          tabBarActiveTintColor: '#007AFF',
          tabBarInactiveTintColor: 'gray',
          tabBarStyle: {
            backgroundColor: 'white',
            borderTopWidth: 1,
            borderTopColor: '#E0E0E0',
            paddingTop: 5,
            paddingBottom: 5,
            height: 60,
          },
          tabBarLabelStyle: {
            fontSize: 12,
            fontWeight: '500',
          },
          headerShown: false, // Hide headers since we have stack navigators
        })}
      >
        <Tab.Screen 
          name="Discovery" 
          component={DiscoveryStack}
          options={{ title: 'Discovery' }}
        />
        <Tab.Screen 
          name="Control" 
          component={ControlStack}
          options={{ title: 'Control' }}
        />
        <Tab.Screen 
          name="MultiDevice" 
          component={MultiDeviceStack}
          options={{ title: 'Multi-Device' }}
        />
        <Tab.Screen 
          name="Settings" 
          component={SettingsStack}
          options={{ title: 'Settings' }}
        />
      </Tab.Navigator>
    </NavigationContainer>
  );
};