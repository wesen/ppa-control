/**
 * Settings Screen
 * App configuration and preferences
 */

import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Switch,
  TextInput,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { DebugPanel } from '../components/DebugPanel';
import { logger, LogLevel } from '../utils/logger';
import { useDispatch, useSelector } from 'react-redux';
import { RootState } from '../store';
import {
  setDefaultPort,
  setDiscoveryInterval,
  setDeviceTimeout,
  setBroadcastAddress,
  setTheme,
  setEnableHapticFeedback,
  setEnableNotifications,
  setLogLevel,
  setEnablePacketLogging,
  setAutoDiscovery,
  setKeepScreenOn,
  setConfirmCommands,
  resetSettings,
} from '../store/settingsSlice';
import { clearDevices } from '../store/deviceSlice';
import { clearFeedbackMessages, resetControlState } from '../store/controlSlice';

interface SettingsScreenProps {
  navigation: any;
}

export const SettingsScreen: React.FC<SettingsScreenProps> = ({ navigation }) => {
  const dispatch = useDispatch();
  const settings = useSelector((state: RootState) => state.settings);
  const [debugPanelVisible, setDebugPanelVisible] = useState(false);

  // Handle reset settings
  const handleResetSettings = () => {
    Alert.alert(
      'Reset Settings',
      'This will reset all settings to their default values. Are you sure?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Reset',
          style: 'destructive',
          onPress: () => {
            dispatch(resetSettings());
            Alert.alert('Success', 'Settings have been reset to defaults');
          },
        },
      ]
    );
  };

  // Handle clear data
  const handleClearData = () => {
    Alert.alert(
      'Clear App Data',
      'This will clear all discovered devices and feedback messages. Are you sure?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Clear',
          style: 'destructive',
          onPress: () => {
            dispatch(clearDevices());
            dispatch(clearFeedbackMessages());
            dispatch(resetControlState());
            Alert.alert('Success', 'App data has been cleared');
          },
        },
      ]
    );
  };

  // Render section
  const renderSection = (title: string, children: React.ReactNode) => (
    <View style={styles.section}>
      <Text style={styles.sectionTitle}>{title}</Text>
      {children}
    </View>
  );

  // Render setting item with switch
  const renderSwitchItem = (
    label: string,
    description: string,
    value: boolean,
    onValueChange: (value: boolean) => void,
  ) => (
    <View style={styles.settingItem}>
      <View style={styles.settingInfo}>
        <Text style={styles.settingLabel}>{label}</Text>
        <Text style={styles.settingDescription}>{description}</Text>
      </View>
      <Switch
        value={value}
        onValueChange={onValueChange}
        trackColor={{ false: '#767577', true: '#81b0ff' }}
        thumbColor={value ? '#007AFF' : '#f4f3f4'}
      />
    </View>
  );

  // Render setting item with text input
  const renderTextInputItem = (
    label: string,
    description: string,
    value: string,
    onChangeText: (text: string) => void,
    keyboardType: 'default' | 'numeric' = 'default',
    placeholder?: string,
  ) => (
    <View style={styles.settingItem}>
      <View style={styles.settingInfo}>
        <Text style={styles.settingLabel}>{label}</Text>
        <Text style={styles.settingDescription}>{description}</Text>
      </View>
      <TextInput
        style={styles.textInput}
        value={value}
        onChangeText={onChangeText}
        keyboardType={keyboardType}
        placeholder={placeholder}
      />
    </View>
  );

  // Render picker item
  const renderPickerItem = (
    label: string,
    description: string,
    value: string,
    options: { label: string; value: string }[],
    onValueChange: (value: string) => void,
  ) => (
    <View style={styles.settingItem}>
      <View style={styles.settingInfo}>
        <Text style={styles.settingLabel}>{label}</Text>
        <Text style={styles.settingDescription}>{description}</Text>
      </View>
      <View style={styles.pickerContainer}>
        {options.map((option) => (
          <TouchableOpacity
            key={option.value}
            style={[
              styles.pickerOption,
              value === option.value && styles.pickerOptionSelected,
            ]}
            onPress={() => onValueChange(option.value)}
          >
            <Text style={[
              styles.pickerOptionText,
              value === option.value && styles.pickerOptionTextSelected,
            ]}>
              {option.label}
            </Text>
          </TouchableOpacity>
        ))}
      </View>
    </View>
  );

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      {/* Network Settings */}
      {renderSection('Network Settings', (
        <>
          {renderTextInputItem(
            'Default Port',
            'Default UDP port for device communication',
            settings.defaultPort.toString(),
            (text) => {
              const port = parseInt(text, 10);
              if (!isNaN(port) && port >= 1 && port <= 65535) {
                dispatch(setDefaultPort(port));
              }
            },
            'numeric',
            '5001'
          )}
          
          {renderTextInputItem(
            'Discovery Interval',
            'Time between discovery broadcasts (ms)',
            settings.discoveryInterval.toString(),
            (text) => {
              const interval = parseInt(text, 10);
              if (!isNaN(interval) && interval >= 1000) {
                dispatch(setDiscoveryInterval(interval));
              }
            },
            'numeric',
            '5000'
          )}
          
          {renderTextInputItem(
            'Device Timeout',
            'Time before marking device as offline (ms)',
            settings.deviceTimeout.toString(),
            (text) => {
              const timeout = parseInt(text, 10);
              if (!isNaN(timeout) && timeout >= 5000) {
                dispatch(setDeviceTimeout(timeout));
              }
            },
            'numeric',
            '30000'
          )}
          
          {renderTextInputItem(
            'Broadcast Address',
            'Network broadcast address for discovery',
            settings.broadcastAddress,
            (text) => dispatch(setBroadcastAddress(text)),
            'default',
            '255.255.255.255'
          )}
        </>
      ))}

      {/* UI Settings */}
      {renderSection('User Interface', (
        <>
          {renderPickerItem(
            'Theme',
            'App color scheme preference',
            settings.theme,
            [
              { label: 'Auto', value: 'auto' },
              { label: 'Light', value: 'light' },
              { label: 'Dark', value: 'dark' },
            ],
            (value) => dispatch(setTheme(value as any))
          )}
          
          {renderSwitchItem(
            'Haptic Feedback',
            'Vibration feedback for interactions',
            settings.enableHapticFeedback,
            (value) => dispatch(setEnableHapticFeedback(value))
          )}
          
          {renderSwitchItem(
            'Notifications',
            'Enable push notifications for device events',
            settings.enableNotifications,
            (value) => dispatch(setEnableNotifications(value))
          )}
        </>
      ))}

      {/* App Behavior */}
      {renderSection('App Behavior', (
        <>
          {renderSwitchItem(
            'Auto Discovery',
            'Automatically start device discovery on app launch',
            settings.autoDiscovery,
            (value) => dispatch(setAutoDiscovery(value))
          )}
          
          {renderSwitchItem(
            'Keep Screen On',
            'Prevent screen from turning off during use',
            settings.keepScreenOn,
            (value) => dispatch(setKeepScreenOn(value))
          )}
          
          {renderSwitchItem(
            'Confirm Commands',
            'Show confirmation dialogs for control actions',
            settings.confirmCommands,
            (value) => dispatch(setConfirmCommands(value))
          )}
        </>
      ))}

      {/* Debug Settings */}
      {renderSection('Debug & Logging', (
        <>
          {renderPickerItem(
            'Log Level',
            'Amount of logging information to display',
            settings.logLevel,
            [
              { label: 'Error', value: 'error' },
              { label: 'Warning', value: 'warn' },
              { label: 'Info', value: 'info' },
              { label: 'Debug', value: 'debug' },
            ],
            (value) => {
              dispatch(setLogLevel(value as any));
              // Update logger level immediately
              const logLevelMap = {
                'error': LogLevel.ERROR,
                'warn': LogLevel.WARN,
                'info': LogLevel.INFO,
                'debug': LogLevel.DEBUG,
              };
              logger.setLevel(logLevelMap[value as keyof typeof logLevelMap]);
            }
          )}
          
          {renderSwitchItem(
            'Packet Logging',
            'Log detailed network packet information',
            settings.enablePacketLogging,
            (value) => dispatch(setEnablePacketLogging(value))
          )}
          
          <TouchableOpacity 
            style={styles.debugButton}
            onPress={() => setDebugPanelVisible(true)}
          >
            <Text style={styles.debugButtonText}>Open Debug Panel</Text>
          </TouchableOpacity>
        </>
      ))}

      {/* Actions */}
      {renderSection('Data Management', (
        <View style={styles.actionButtons}>
          <TouchableOpacity style={styles.clearButton} onPress={handleClearData}>
            <Text style={styles.clearButtonText}>Clear App Data</Text>
          </TouchableOpacity>
          
          <TouchableOpacity style={styles.resetButton} onPress={handleResetSettings}>
            <Text style={styles.resetButtonText}>Reset Settings</Text>
          </TouchableOpacity>
        </View>
      ))}

      {/* About */}
      {renderSection('About', (
        <>
          <View style={styles.aboutItem}>
            <Text style={styles.aboutLabel}>App Version</Text>
            <Text style={styles.aboutValue}>1.0.0</Text>
          </View>
          
          <View style={styles.aboutItem}>
            <Text style={styles.aboutLabel}>Protocol Version</Text>
            <Text style={styles.aboutValue}>PPA v1.0</Text>
          </View>
          
          <View style={styles.aboutItem}>
            <Text style={styles.aboutLabel}>Default Port</Text>
            <Text style={styles.aboutValue}>UDP 5001</Text>
          </View>
          
          <TouchableOpacity 
            style={styles.aboutButton}
            onPress={() => Alert.alert(
              'PPA Control',
              'Professional audio control application for PPA DSP systems.\n\nBuilt with React Native and Expo.',
              [{ text: 'OK' }]
            )}
          >
            <Text style={styles.aboutButtonText}>More Information</Text>
          </TouchableOpacity>
        </>
      ))}
      
      <DebugPanel 
        visible={debugPanelVisible}
        onClose={() => setDebugPanelVisible(false)}
      />
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  section: {
    backgroundColor: 'white',
    marginHorizontal: 16,
    marginVertical: 8,
    padding: 16,
    borderRadius: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
    marginBottom: 16,
  },
  settingItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  settingInfo: {
    flex: 1,
    marginRight: 16,
  },
  settingLabel: {
    fontSize: 16,
    fontWeight: '500',
    color: '#333',
    marginBottom: 2,
  },
  settingDescription: {
    fontSize: 13,
    color: '#666',
  },
  textInput: {
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 6,
    padding: 8,
    fontSize: 14,
    backgroundColor: '#fafafa',
    minWidth: 100,
    textAlign: 'right',
  },
  pickerContainer: {
    flexDirection: 'row',
    gap: 4,
  },
  pickerOption: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
    backgroundColor: '#f0f0f0',
    borderWidth: 1,
    borderColor: 'transparent',
  },
  pickerOptionSelected: {
    backgroundColor: '#007AFF',
    borderColor: '#0056CC',
  },
  pickerOptionText: {
    fontSize: 12,
    color: '#333',
    fontWeight: '500',
  },
  pickerOptionTextSelected: {
    color: 'white',
  },
  actionButtons: {
    gap: 12,
  },
  clearButton: {
    backgroundColor: '#FF5722',
    paddingVertical: 12,
    paddingHorizontal: 20,
    borderRadius: 8,
    alignItems: 'center',
  },
  clearButtonText: {
    color: 'white',
    fontSize: 16,
    fontWeight: '600',
  },
  resetButton: {
    backgroundColor: '#FF9800',
    paddingVertical: 12,
    paddingHorizontal: 20,
    borderRadius: 8,
    alignItems: 'center',
  },
  resetButtonText: {
    color: 'white',
    fontSize: 16,
    fontWeight: '600',
  },
  aboutItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  aboutLabel: {
    fontSize: 14,
    color: '#666',
  },
  aboutValue: {
    fontSize: 14,
    color: '#333',
    fontWeight: '500',
  },
  aboutButton: {
    marginTop: 12,
    paddingVertical: 10,
    paddingHorizontal: 16,
    backgroundColor: '#007AFF',
    borderRadius: 6,
    alignItems: 'center',
  },
  aboutButtonText: {
    color: 'white',
    fontSize: 14,
    fontWeight: '600',
  },
  debugButton: {
    marginTop: 16,
    paddingVertical: 12,
    paddingHorizontal: 16,
    backgroundColor: '#FF9500',
    borderRadius: 6,
    alignItems: 'center',
  },
  debugButtonText: {
    color: 'white',
    fontSize: 14,
    fontWeight: '600',
  },
});