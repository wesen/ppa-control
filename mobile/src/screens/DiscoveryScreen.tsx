/**
 * Device Discovery Screen
 * Displays auto-discovered devices and allows manual IP entry
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TextInput,
  TouchableOpacity,
  FlatList,
  RefreshControl,
  Alert,
  Switch,
} from 'react-native';
import { useDispatch, useSelector } from 'react-redux';
import { RootState } from '../store';
import { useUDPService } from '../hooks/useUDPService';
import { selectDevice } from '../store/deviceSlice';
import { DeviceInfo } from '../protocol/types';

interface DiscoveryScreenProps {
  navigation: any;
}

export const DiscoveryScreen: React.FC<DiscoveryScreenProps> = ({ navigation }) => {
  const dispatch = useDispatch();
  const { devices, isDiscovering, discoveryError } = useSelector((state: RootState) => state.devices);
  const { autoDiscovery } = useSelector((state: RootState) => state.settings);
  
  const {
    startService,
    stopService,
    addDeviceManually,
    toggleDiscovery,
    sendPing,
    isServiceRunning,
  } = useUDPService();

  const [manualIP, setManualIP] = useState('');
  const [manualPort, setManualPort] = useState('5001');
  const [refreshing, setRefreshing] = useState(false);

  // Start UDP service on mount
  useEffect(() => {
    if (!isServiceRunning) {
      startService();
    }
  }, [startService, isServiceRunning]);

  // Handle device selection
  const handleDeviceSelect = (device: DeviceInfo) => {
    dispatch(selectDevice(device));
    navigation.navigate('Control');
  };

  // Handle manual device addition
  const handleAddDevice = () => {
    if (!manualIP.trim()) {
      Alert.alert('Error', 'Please enter an IP address');
      return;
    }

    const port = parseInt(manualPort, 10);
    if (isNaN(port) || port < 1 || port > 65535) {
      Alert.alert('Error', 'Please enter a valid port number (1-65535)');
      return;
    }

    // Check if device already exists
    const deviceAddress = `${manualIP.trim()}:${port}`;
    const existingDevice = devices.find(d => `${d.address}:${d.port}` === deviceAddress);
    
    if (existingDevice) {
      Alert.alert('Device Already Added', 'This device is already in the list');
      return;
    }

    const device = addDeviceManually(manualIP.trim(), port);
    if (device) {
      setManualIP('');
      Alert.alert('Success', `Device ${manualIP.trim()}:${port} added successfully`);
    }
  };

  // Handle refresh
  const handleRefresh = () => {
    setRefreshing(true);
    
    // Restart discovery
    if (isServiceRunning) {
      toggleDiscovery(false);
      setTimeout(() => {
        toggleDiscovery(true);
        setRefreshing(false);
      }, 1000);
    } else {
      setRefreshing(false);
    }
  };

  // Handle discovery toggle
  const handleDiscoveryToggle = (enabled: boolean) => {
    toggleDiscovery(enabled);
  };

  // Handle device ping
  const handlePingDevice = (device: DeviceInfo) => {
    sendPing(`${device.address}:${device.port}`);
  };

  // Render device item
  const renderDeviceItem = ({ item }: { item: DeviceInfo }) => {
    const deviceAddress = `${item.address}:${item.port}`;
    
    return (
      <View style={styles.deviceItem}>
        <TouchableOpacity
          style={[styles.deviceInfo, !item.isConnected && styles.deviceOffline]}
          onPress={() => handleDeviceSelect(item)}
        >
          <View style={styles.deviceHeader}>
            <Text style={styles.deviceName}>{item.name}</Text>
            <View style={[
              styles.statusIndicator,
              { backgroundColor: item.isConnected ? '#4CAF50' : '#FF5722' }
            ]} />
          </View>
          <Text style={styles.deviceAddress}>{deviceAddress}</Text>
          <Text style={styles.deviceStatus}>
            {item.isConnected ? 'Online' : 'Offline'} • Last seen: {item.lastSeen.toLocaleTimeString()}
          </Text>
        </TouchableOpacity>
        
        <TouchableOpacity
          style={styles.pingButton}
          onPress={() => handlePingDevice(item)}
        >
          <Text style={styles.pingButtonText}>Ping</Text>
        </TouchableOpacity>
      </View>
    );
  };

  return (
    <View style={styles.container}>
      {/* Discovery Controls */}
      <View style={styles.section}>
        <View style={styles.sectionHeader}>
          <Text style={styles.sectionTitle}>Device Discovery</Text>
          <Switch
            value={isDiscovering}
            onValueChange={handleDiscoveryToggle}
            trackColor={{ false: '#767577', true: '#81b0ff' }}
            thumbColor={isDiscovering ? '#007AFF' : '#f4f3f4'}
          />
        </View>
        
        {discoveryError && (
          <View style={styles.errorContainer}>
            <Text style={styles.errorText}>Error: {discoveryError}</Text>
          </View>
        )}
      </View>

      {/* Manual Entry */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Manual Entry</Text>
        <View style={styles.manualEntry}>
          <TextInput
            style={[styles.input, styles.ipInput]}
            placeholder="IP Address"
            value={manualIP}
            onChangeText={setManualIP}
            keyboardType="numeric"
            autoCapitalize="none"
            autoCorrect={false}
          />
          <TextInput
            style={[styles.input, styles.portInput]}
            placeholder="Port"
            value={manualPort}
            onChangeText={setManualPort}
            keyboardType="numeric"
          />
          <TouchableOpacity style={styles.addButton} onPress={handleAddDevice}>
            <Text style={styles.addButtonText}>Add</Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Device List */}
      <View style={[styles.section, styles.deviceListSection]}>
        <Text style={styles.sectionTitle}>
          Devices ({devices.length})
        </Text>
        
        <FlatList
          data={devices}
          keyExtractor={(item) => `${item.address}:${item.port}`}
          renderItem={renderDeviceItem}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={handleRefresh}
              tintColor="#007AFF"
            />
          }
          ListEmptyComponent={
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>
                {isDiscovering ? 'Searching for devices...' : 'No devices found'}
              </Text>
              <Text style={styles.emptySubtext}>
                Pull to refresh or add devices manually
              </Text>
            </View>
          }
          showsVerticalScrollIndicator={false}
        />
      </View>
    </View>
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
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
    marginBottom: 8,
  },
  errorContainer: {
    backgroundColor: '#FFEBEE',
    padding: 12,
    borderRadius: 8,
    marginTop: 8,
  },
  errorText: {
    color: '#D32F2F',
    fontSize: 14,
  },
  manualEntry: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  input: {
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 8,
    padding: 12,
    fontSize: 16,
    backgroundColor: '#fafafa',
  },
  ipInput: {
    flex: 2,
  },
  portInput: {
    flex: 1,
  },
  addButton: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 8,
  },
  addButtonText: {
    color: 'white',
    fontSize: 16,
    fontWeight: '600',
  },
  deviceListSection: {
    flex: 1,
  },
  deviceItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  deviceInfo: {
    flex: 1,
    padding: 12,
    backgroundColor: '#f8f9fa',
    borderRadius: 8,
    marginRight: 8,
  },
  deviceOffline: {
    opacity: 0.6,
  },
  deviceHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  deviceName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
  statusIndicator: {
    width: 10,
    height: 10,
    borderRadius: 5,
  },
  deviceAddress: {
    fontSize: 14,
    color: '#666',
    marginBottom: 2,
  },
  deviceStatus: {
    fontSize: 12,
    color: '#999',
  },
  pingButton: {
    backgroundColor: '#4CAF50',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 6,
  },
  pingButtonText: {
    color: 'white',
    fontSize: 14,
    fontWeight: '600',
  },
  emptyContainer: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 40,
  },
  emptyText: {
    fontSize: 16,
    color: '#666',
    marginBottom: 8,
  },
  emptySubtext: {
    fontSize: 14,
    color: '#999',
  },
});