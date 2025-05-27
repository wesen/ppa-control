/**
 * Device Control Screen
 * Volume control, preset management, and real-time feedback
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
  Alert,
  ActivityIndicator,
} from 'react-native';
import Slider from '@react-native-community/slider';
import { useDispatch, useSelector } from 'react-redux';
import { RootState } from '../store';
import { useUDPService } from '../hooks/useUDPService';
import {
  setVolume,
  toggleMute,
  setCurrentPreset,
  clearError,
} from '../store/controlSlice';

interface ControlScreenProps {
  navigation: any;
}

export const ControlScreen: React.FC<ControlScreenProps> = ({ navigation }) => {
  const dispatch = useDispatch();
  const { selectedDevice } = useSelector((state: RootState) => state.devices);
  const {
    volume,
    isMuted,
    currentPreset,
    isProcessingCommand,
    lastCommandStatus,
    errorMessage,
    feedbackMessages,
  } = useSelector((state: RootState) => state.control);
  const { confirmCommands } = useSelector((state: RootState) => state.settings);

  const { sendVolumeCommand, sendPresetCommand } = useUDPService();

  const [localVolume, setLocalVolume] = useState(volume);
  const [showAdvanced, setShowAdvanced] = useState(false);

  // Update local volume when store volume changes
  useEffect(() => {
    setLocalVolume(volume);
  }, [volume]);

  // Navigate back if no device selected
  useEffect(() => {
    if (!selectedDevice) {
      navigation.goBack();
    }
  }, [selectedDevice, navigation]);

  if (!selectedDevice) {
    return (
      <View style={styles.container}>
        <Text style={styles.errorText}>No device selected</Text>
      </View>
    );
  }

  const deviceAddress = `${selectedDevice.address}:${selectedDevice.port}`;

  // Handle volume change
  const handleVolumeChange = (newVolume: number) => {
    setLocalVolume(newVolume);
  };

  // Handle volume change end (when user releases slider)
  const handleVolumeChangeEnd = async (newVolume: number) => {
    const confirmAction = async () => {
      dispatch(setVolume(newVolume));
      await sendVolumeCommand(deviceAddress, newVolume);
    };

    if (confirmCommands) {
      Alert.alert(
        'Confirm Volume Change',
        `Set volume to ${Math.round(newVolume * 100)}%?`,
        [
          { text: 'Cancel', style: 'cancel' },
          { text: 'Confirm', onPress: confirmAction },
        ]
      );
    } else {
      await confirmAction();
    }
  };

  // Handle mute toggle
  const handleMuteToggle = async () => {
    const confirmAction = async () => {
      dispatch(toggleMute());
      // Send volume 0 if muting, restore volume if unmuting
      const targetVolume = isMuted ? volume : 0;
      await sendVolumeCommand(deviceAddress, targetVolume);
    };

    if (confirmCommands) {
      Alert.alert(
        'Confirm Mute',
        isMuted ? 'Unmute device?' : 'Mute device?',
        [
          { text: 'Cancel', style: 'cancel' },
          { text: 'Confirm', onPress: confirmAction },
        ]
      );
    } else {
      await confirmAction();
    }
  };

  // Handle preset recall
  const handlePresetRecall = async (presetIndex: number) => {
    const confirmAction = async () => {
      dispatch(setCurrentPreset(presetIndex));
      await sendPresetCommand(deviceAddress, presetIndex);
    };

    if (confirmCommands) {
      Alert.alert(
        'Confirm Preset Recall',
        `Recall preset ${presetIndex}?`,
        [
          { text: 'Cancel', style: 'cancel' },
          { text: 'Confirm', onPress: confirmAction },
        ]
      );
    } else {
      await confirmAction();
    }
  };

  // Clear error message
  const handleClearError = () => {
    dispatch(clearError());
  };

  // Generate preset buttons (1-16)
  const renderPresetButtons = () => {
    const presets = Array.from({ length: 16 }, (_, i) => i + 1);
    
    return (
      <View style={styles.presetGrid}>
        {presets.map((preset) => (
          <TouchableOpacity
            key={preset}
            style={[
              styles.presetButton,
              currentPreset === preset && styles.presetButtonActive,
              isProcessingCommand && styles.presetButtonDisabled,
            ]}
            onPress={() => handlePresetRecall(preset)}
            disabled={isProcessingCommand}
          >
            <Text style={[
              styles.presetButtonText,
              currentPreset === preset && styles.presetButtonTextActive,
            ]}>
              {preset}
            </Text>
          </TouchableOpacity>
        ))}
      </View>
    );
  };

  // Calculate volume display value
  const displayVolume = isMuted ? 0 : Math.round(localVolume * 100);
  const volumeDb = isMuted ? '-∞' : `${Math.round((localVolume * 100 - 80))}`;

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      {/* Device Header */}
      <View style={styles.section}>
        <View style={styles.deviceHeader}>
          <View>
            <Text style={styles.deviceName}>{selectedDevice.name}</Text>
            <Text style={styles.deviceAddress}>{deviceAddress}</Text>
          </View>
          <View style={[
            styles.connectionStatus,
            { backgroundColor: selectedDevice.isConnected ? '#4CAF50' : '#FF5722' }
          ]}>
            <Text style={styles.connectionStatusText}>
              {selectedDevice.isConnected ? 'Connected' : 'Disconnected'}
            </Text>
          </View>
        </View>
      </View>

      {/* Error Display */}
      {errorMessage && (
        <View style={styles.section}>
          <View style={styles.errorContainer}>
            <Text style={styles.errorText}>{errorMessage}</Text>
            <TouchableOpacity onPress={handleClearError} style={styles.clearErrorButton}>
              <Text style={styles.clearErrorText}>Clear</Text>
            </TouchableOpacity>
          </View>
        </View>
      )}

      {/* Volume Control */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Volume Control</Text>
        
        <View style={styles.volumeContainer}>
          <View style={styles.volumeDisplay}>
            <Text style={styles.volumeValue}>{displayVolume}%</Text>
            <Text style={styles.volumeDb}>{volumeDb} dB</Text>
          </View>
          
          <Slider
            style={styles.volumeSlider}
            minimumValue={0}
            maximumValue={1}
            value={localVolume}
            onValueChange={handleVolumeChange}
            onSlidingComplete={handleVolumeChangeEnd}
            minimumTrackTintColor="#007AFF"
            maximumTrackTintColor="#E0E0E0"
            disabled={isProcessingCommand}
          />
          
          <TouchableOpacity
            style={[
              styles.muteButton,
              isMuted && styles.muteButtonActive,
              isProcessingCommand && styles.muteButtonDisabled,
            ]}
            onPress={handleMuteToggle}
            disabled={isProcessingCommand}
          >
            <Text style={[
              styles.muteButtonText,
              isMuted && styles.muteButtonTextActive,
            ]}>
              {isMuted ? 'Unmute' : 'Mute'}
            </Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Preset Management */}
      <View style={styles.section}>
        <View style={styles.sectionHeader}>
          <Text style={styles.sectionTitle}>Presets</Text>
          {currentPreset && (
            <Text style={styles.currentPresetText}>
              Current: {currentPreset}
            </Text>
          )}
        </View>
        
        {renderPresetButtons()}
      </View>

      {/* Advanced Controls */}
      <View style={styles.section}>
        <TouchableOpacity
          style={styles.advancedToggle}
          onPress={() => setShowAdvanced(!showAdvanced)}
        >
          <Text style={styles.advancedToggleText}>
            Advanced Controls {showAdvanced ? '▼' : '▶'}
          </Text>
        </TouchableOpacity>
        
        {showAdvanced && (
          <View style={styles.advancedContainer}>
            <Text style={styles.advancedNote}>
              Advanced controls (EQ, individual gains) will be available in future versions.
            </Text>
          </View>
        )}
      </View>

      {/* Status Feedback */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Status</Text>
        
        {isProcessingCommand && (
          <View style={styles.processingContainer}>
            <ActivityIndicator size="small" color="#007AFF" />
            <Text style={styles.processingText}>Processing command...</Text>
          </View>
        )}
        
        {lastCommandStatus && !isProcessingCommand && (
          <View style={[
            styles.statusContainer,
            lastCommandStatus === 'success' && styles.statusSuccess,
            lastCommandStatus === 'error' && styles.statusError,
          ]}>
            <Text style={styles.statusText}>
              Last command: {lastCommandStatus}
            </Text>
          </View>
        )}
        
        <View style={styles.feedbackContainer}>
          <Text style={styles.feedbackTitle}>Recent Activity:</Text>
          {feedbackMessages.slice(-5).map((message, index) => (
            <Text key={index} style={styles.feedbackMessage}>
              • {message}
            </Text>
          ))}
        </View>
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
    marginBottom: 12,
  },
  deviceHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  deviceName: {
    fontSize: 20,
    fontWeight: '700',
    color: '#333',
  },
  deviceAddress: {
    fontSize: 14,
    color: '#666',
    marginTop: 2,
  },
  connectionStatus: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
  },
  connectionStatusText: {
    color: 'white',
    fontSize: 12,
    fontWeight: '600',
  },
  errorContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: '#FFEBEE',
    padding: 12,
    borderRadius: 8,
  },
  errorText: {
    flex: 1,
    color: '#D32F2F',
    fontSize: 14,
  },
  clearErrorButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    backgroundColor: '#D32F2F',
    borderRadius: 6,
  },
  clearErrorText: {
    color: 'white',
    fontSize: 12,
    fontWeight: '600',
  },
  volumeContainer: {
    alignItems: 'center',
  },
  volumeDisplay: {
    alignItems: 'center',
    marginBottom: 20,
  },
  volumeValue: {
    fontSize: 36,
    fontWeight: '700',
    color: '#007AFF',
  },
  volumeDb: {
    fontSize: 16,
    color: '#666',
    marginTop: 4,
  },
  volumeSlider: {
    width: '100%',
    height: 40,
    marginBottom: 20,
  },

  muteButton: {
    backgroundColor: '#FF5722',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
    minWidth: 100,
    alignItems: 'center',
  },
  muteButtonActive: {
    backgroundColor: '#4CAF50',
  },
  muteButtonDisabled: {
    opacity: 0.5,
  },
  muteButtonText: {
    color: 'white',
    fontSize: 16,
    fontWeight: '600',
  },
  muteButtonTextActive: {
    color: 'white',
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
    opacity: 0.5,
  },
  presetButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
  presetButtonTextActive: {
    color: 'white',
  },
  advancedToggle: {
    paddingVertical: 8,
  },
  advancedToggleText: {
    fontSize: 16,
    color: '#007AFF',
    fontWeight: '600',
  },
  advancedContainer: {
    marginTop: 12,
    padding: 16,
    backgroundColor: '#f8f9fa',
    borderRadius: 8,
  },
  advancedNote: {
    fontSize: 14,
    color: '#666',
    fontStyle: 'italic',
  },
  processingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  processingText: {
    marginLeft: 8,
    fontSize: 14,
    color: '#666',
  },
  statusContainer: {
    padding: 8,
    borderRadius: 6,
    marginBottom: 12,
  },
  statusSuccess: {
    backgroundColor: '#E8F5E8',
  },
  statusError: {
    backgroundColor: '#FFEBEE',
  },
  statusText: {
    fontSize: 14,
    fontWeight: '600',
  },
  feedbackContainer: {
    marginTop: 8,
  },
  feedbackTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#333',
    marginBottom: 8,
  },
  feedbackMessage: {
    fontSize: 12,
    color: '#666',
    marginBottom: 4,
  },
});