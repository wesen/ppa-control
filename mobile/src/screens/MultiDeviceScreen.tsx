/**
 * Multi-Device Control Screen
 * Control multiple devices simultaneously
 */

import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
  Alert,
  Switch,
} from 'react-native';
import Slider from '@react-native-community/slider';
import { useDispatch, useSelector } from 'react-redux';
import { RootState } from '../store';
import { useUDPService } from '../hooks/useUDPService';
import {
  toggleDeviceSelection,
  selectAllDevices,
  clearDeviceSelection,
} from '../store/deviceSlice';
import { setVolume } from '../store/controlSlice';
import { DeviceInfo } from '../protocol/types';

interface MultiDeviceScreenProps {
  navigation: any;
}

export const MultiDeviceScreen: React.FC<MultiDeviceScreenProps> = ({ navigation }) => {
  const dispatch = useDispatch();
  const { devices, selectedDevices } = useSelector((state: RootState) => state.devices);
  const { volume, isProcessingCommand } = useSelector((state: RootState) => state.control);
  const { confirmCommands } = useSelector((state: RootState) => state.settings);

  const { sendVolumeCommand, sendPresetCommand } = useUDPService();

  const [masterVolume, setMasterVolume] = useState(volume);
  const [syncedPreset, setSyncedPreset] = useState<number | null>(null);

  // Get connected devices
  const connectedDevices = devices.filter(d => d.isConnected);
  const selectedDeviceList = devices.filter(d => 
    selectedDevices.includes(`${d.address}:${d.port}`)
  );

  // Handle device selection toggle
  const handleDeviceToggle = (device: DeviceInfo) => {
    const deviceAddress = `${device.address}:${device.port}`;
    dispatch(toggleDeviceSelection(deviceAddress));
  };

  // Handle select all devices
  const handleSelectAll = () => {
    if (selectedDevices.length === connectedDevices.length) {
      dispatch(clearDeviceSelection());
    } else {
      dispatch(selectAllDevices());
    }
  };

  // Handle master volume change
  const handleMasterVolumeChange = (newVolume: number) => {
    setMasterVolume(newVolume);
  };

  // Handle master volume change end
  const handleMasterVolumeChangeEnd = async (newVolume: number) => {
    if (selectedDevices.length === 0) {
      Alert.alert('No Devices Selected', 'Please select devices to control');
      return;
    }

    const confirmAction = async () => {
      dispatch(setVolume(newVolume));
      
      // Send volume command to all selected devices
      const promises = selectedDevices.map(deviceAddress => 
        sendVolumeCommand(deviceAddress, newVolume)
      );
      
      try {
        await Promise.all(promises);
        Alert.alert(
          'Success', 
          `Volume set to ${Math.round(newVolume * 100)}% on ${selectedDevices.length} devices`
        );
      } catch (error) {
        Alert.alert('Error', 'Failed to set volume on some devices');
      }
    };

    if (confirmCommands) {
      Alert.alert(
        'Confirm Volume Change',
        `Set volume to ${Math.round(newVolume * 100)}% on ${selectedDevices.length} selected devices?`,
        [
          { text: 'Cancel', style: 'cancel' },
          { text: 'Confirm', onPress: confirmAction },
        ]
      );
    } else {
      await confirmAction();
    }
  };

  // Handle preset sync
  const handlePresetSync = async (presetIndex: number) => {
    if (selectedDevices.length === 0) {
      Alert.alert('No Devices Selected', 'Please select devices to control');
      return;
    }

    const confirmAction = async () => {
      setSyncedPreset(presetIndex);
      
      // Send preset command to all selected devices
      const promises = selectedDevices.map(deviceAddress => 
        sendPresetCommand(deviceAddress, presetIndex)
      );
      
      try {
        await Promise.all(promises);
        Alert.alert(
          'Success', 
          `Preset ${presetIndex} recalled on ${selectedDevices.length} devices`
        );
      } catch (error) {
        Alert.alert('Error', 'Failed to recall preset on some devices');
      }
    };

    if (confirmCommands) {
      Alert.alert(
        'Confirm Preset Sync',
        `Recall preset ${presetIndex} on ${selectedDevices.length} selected devices?`,
        [
          { text: 'Cancel', style: 'cancel' },
          { text: 'Confirm', onPress: confirmAction },
        ]
      );
    } else {
      await confirmAction();
    }
  };

  // Render device selection item
  const renderDeviceItem = (device: DeviceInfo) => {
    const deviceAddress = `${device.address}:${device.port}`;
    const isSelected = selectedDevices.includes(deviceAddress);

    return (
      <View key={deviceAddress} style={styles.deviceItem}>
        <View style={styles.deviceInfo}>
          <Text style={styles.deviceName}>{device.name}</Text>
          <Text style={styles.deviceAddress}>{deviceAddress}</Text>
          <View style={styles.deviceStatus}>
            <View style={[
              styles.statusIndicator,
              { backgroundColor: device.isConnected ? '#4CAF50' : '#FF5722' }
            ]} />
            <Text style={styles.statusText}>
              {device.isConnected ? 'Online' : 'Offline'}
            </Text>
          </View>
        </View>
        
        <Switch
          value={isSelected}
          onValueChange={() => handleDeviceToggle(device)}
          trackColor={{ false: '#767577', true: '#81b0ff' }}
          thumbColor={isSelected ? '#007AFF' : '#f4f3f4'}
          disabled={!device.isConnected}
        />
      </View>
    );
  };

  // Generate preset buttons
  const renderPresetButtons = () => {
    const presets = Array.from({ length: 16 }, (_, i) => i + 1);
    
    return (
      <View style={styles.presetGrid}>
        {presets.map((preset) => (
          <TouchableOpacity
            key={preset}
            style={[
              styles.presetButton,
              syncedPreset === preset && styles.presetButtonActive,
              (isProcessingCommand || selectedDevices.length === 0) && styles.presetButtonDisabled,
            ]}
            onPress={() => handlePresetSync(preset)}
            disabled={isProcessingCommand || selectedDevices.length === 0}
          >
            <Text style={[
              styles.presetButtonText,
              syncedPreset === preset && styles.presetButtonTextActive,
            ]}>
              {preset}
            </Text>
          </TouchableOpacity>
        ))}
      </View>
    );
  };

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      {/* Selection Summary */}
      <View style={styles.section}>
        <View style={styles.summaryHeader}>
          <Text style={styles.sectionTitle}>Device Selection</Text>
          <TouchableOpacity onPress={handleSelectAll} style={styles.selectAllButton}>
            <Text style={styles.selectAllText}>
              {selectedDevices.length === connectedDevices.length ? 'Deselect All' : 'Select All'}
            </Text>
          </TouchableOpacity>
        </View>
        
        <Text style={styles.summaryText}>
          {selectedDevices.length} of {connectedDevices.length} devices selected
        </Text>
        
        {selectedDevices.length === 0 && (
          <View style={styles.warningContainer}>
            <Text style={styles.warningText}>
              Select devices to enable multi-device control
            </Text>
          </View>
        )}
      </View>

      {/* Device List */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Available Devices</Text>
        
        {connectedDevices.length === 0 ? (
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyText}>No connected devices found</Text>
            <TouchableOpacity 
              style={styles.discoveryButton}
              onPress={() => navigation.navigate('Discovery')}
            >
              <Text style={styles.discoveryButtonText}>Go to Discovery</Text>
            </TouchableOpacity>
          </View>
        ) : (
          connectedDevices.map(renderDeviceItem)
        )}
      </View>

      {/* Master Volume Control */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Master Volume Control</Text>
        
        <View style={styles.volumeContainer}>
          <View style={styles.volumeDisplay}>
            <Text style={styles.volumeValue}>
              {Math.round(masterVolume * 100)}%
            </Text>
            <Text style={styles.selectedDevicesText}>
              for {selectedDevices.length} devices
            </Text>
          </View>
          
          <Slider
            style={styles.volumeSlider}
            minimumValue={0}
            maximumValue={1}
            value={masterVolume}
            onValueChange={handleMasterVolumeChange}
            onSlidingComplete={handleMasterVolumeChangeEnd}
            minimumTrackTintColor="#007AFF"
            maximumTrackTintColor="#E0E0E0"
            disabled={isProcessingCommand || selectedDevices.length === 0}
          />
        </View>
      </View>

      {/* Preset Synchronization */}
      <View style={styles.section}>
        <View style={styles.sectionHeader}>
          <Text style={styles.sectionTitle}>Preset Synchronization</Text>
          {syncedPreset && (
            <Text style={styles.currentPresetText}>
              Last Synced: {syncedPreset}
            </Text>
          )}
        </View>
        
        <Text style={styles.presetDescription}>
          Apply the same preset to all selected devices
        </Text>
        
        {renderPresetButtons()}
      </View>

      {/* Status Overview */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Status Overview</Text>
        
        <View style={styles.statusGrid}>
          {selectedDeviceList.map((device) => {
            const deviceAddress = `${device.address}:${device.port}`;
            return (
              <View key={deviceAddress} style={styles.statusItem}>
                <Text style={styles.statusDeviceName} numberOfLines={1}>
                  {device.name}
                </Text>
                <View style={[
                  styles.statusIndicator,
                  { backgroundColor: device.isConnected ? '#4CAF50' : '#FF5722' }
                ]} />
              </View>
            );
          })}
        </View>
        
        {selectedDeviceList.length === 0 && (
          <Text style={styles.noStatusText}>
            No devices selected for monitoring
          </Text>
        )}
      </View>
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
    marginBottom: 12,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  summaryHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  selectAllButton: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
  },
  selectAllText: {
    color: 'white',
    fontSize: 14,
    fontWeight: '600',
  },
  summaryText: {
    fontSize: 16,
    color: '#666',
    marginBottom: 8,
  },
  warningContainer: {
    backgroundColor: '#FFF3CD',
    padding: 12,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#FFE69C',
  },
  warningText: {
    color: '#856404',
    fontSize: 14,
  },
  deviceItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  deviceInfo: {
    flex: 1,
  },
  deviceName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 2,
  },
  deviceAddress: {
    fontSize: 14,
    color: '#666',
    marginBottom: 4,
  },
  deviceStatus: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statusIndicator: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 6,
  },
  statusText: {
    fontSize: 12,
    color: '#999',
  },
  emptyContainer: {
    alignItems: 'center',
    paddingVertical: 24,
  },
  emptyText: {
    fontSize: 16,
    color: '#666',
    marginBottom: 16,
  },
  discoveryButton: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 8,
  },
  discoveryButtonText: {
    color: 'white',
    fontSize: 16,
    fontWeight: '600',
  },
  volumeContainer: {
    alignItems: 'center',
  },
  volumeDisplay: {
    alignItems: 'center',
    marginBottom: 16,
  },
  volumeValue: {
    fontSize: 32,
    fontWeight: '700',
    color: '#007AFF',
  },
  selectedDevicesText: {
    fontSize: 14,
    color: '#666',
    marginTop: 4,
  },
  volumeSlider: {
    width: '100%',
    height: 40,
  },

  presetDescription: {
    fontSize: 14,
    color: '#666',
    marginBottom: 16,
  },
  currentPresetText: {
    fontSize: 14,
    color: '#007AFF',
    fontWeight: '600',
  },
  presetGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  presetButton: {
    width: 60,
    height: 60,
    borderRadius: 8,
    backgroundColor: '#f0f0f0',
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: 'transparent',
  },
  presetButtonActive: {
    backgroundColor: '#007AFF',
    borderColor: '#0056CC',
  },
  presetButtonDisabled: {
    opacity: 0.3,
  },
  presetButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
  presetButtonTextActive: {
    color: 'white',
  },
  statusGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  statusItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#f8f9fa',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 8,
    minWidth: 120,
    maxWidth: 150,
  },
  statusDeviceName: {
    flex: 1,
    fontSize: 12,
    color: '#333',
    marginRight: 8,
  },
  noStatusText: {
    fontSize: 14,
    color: '#999',
    textAlign: 'center',
    paddingVertical: 20,
  },
});