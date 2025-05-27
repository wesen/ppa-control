/**
 * Debug Panel Component
 * In-app logging and debugging interface
 */

import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Modal,
  Switch,
  Share,
  Alert,
} from 'react-native';
import { useSelector } from 'react-redux';
import { RootState } from '../store';
import { logger, LogLevel, LogEntry } from '../utils/logger';

interface DebugPanelProps {
  visible: boolean;
  onClose: () => void;
}

export const DebugPanel: React.FC<DebugPanelProps> = ({ visible, onClose }) => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [selectedLevel, setSelectedLevel] = useState<LogLevel>(LogLevel.DEBUG);
  const [selectedContext, setSelectedContext] = useState<string | null>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  
  const settings = useSelector((state: RootState) => state.settings);
  const devices = useSelector((state: RootState) => state.devices);
  const control = useSelector((state: RootState) => state.control);

  // Update logs when they change
  useEffect(() => {
    const updateLogs = () => {
      setLogs(logger.getLogs(selectedLevel, selectedContext || undefined));
    };

    updateLogs();
    logger.addListener(updateLogs);

    return () => {
      logger.removeListener(updateLogs);
    };
  }, [selectedLevel, selectedContext]);

  // Get unique contexts from logs
  const contexts = Array.from(new Set(logger.getLogs().map(entry => entry.context)));

  // Handle log export
  const handleExportLogs = async () => {
    try {
      const logText = logger.exportLogs();
      await Share.share({
        message: logText,
        title: 'PPA Control Debug Logs',
      });
    } catch (error) {
      Alert.alert('Error', 'Failed to export logs');
    }
  };

  // Handle clear logs
  const handleClearLogs = () => {
    Alert.alert(
      'Clear Logs',
      'Are you sure you want to clear all logs?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Clear',
          style: 'destructive',
          onPress: () => {
            logger.clear();
            setLogs([]);
          },
        },
      ]
    );
  };

  // Render log level selector
  const renderLevelSelector = () => (
    <View style={styles.selectorContainer}>
      <Text style={styles.selectorLabel}>Log Level:</Text>
      <View style={styles.levelButtons}>
        {Object.values(LogLevel).filter(v => typeof v === 'number').map((level) => (
          <TouchableOpacity
            key={level}
            style={[
              styles.levelButton,
              selectedLevel === level && styles.levelButtonActive,
            ]}
            onPress={() => setSelectedLevel(level as LogLevel)}
          >
            <Text style={[
              styles.levelButtonText,
              selectedLevel === level && styles.levelButtonTextActive,
            ]}>
              {LogLevel[level as LogLevel]}
            </Text>
          </TouchableOpacity>
        ))}
      </View>
    </View>
  );

  // Render context selector
  const renderContextSelector = () => (
    <View style={styles.selectorContainer}>
      <Text style={styles.selectorLabel}>Context:</Text>
      <ScrollView horizontal showsHorizontalScrollIndicator={false}>
        <TouchableOpacity
          style={[
            styles.contextButton,
            selectedContext === null && styles.contextButtonActive,
          ]}
          onPress={() => setSelectedContext(null)}
        >
          <Text style={[
            styles.contextButtonText,
            selectedContext === null && styles.contextButtonTextActive,
          ]}>
            All
          </Text>
        </TouchableOpacity>
        {contexts.map((context) => (
          <TouchableOpacity
            key={context}
            style={[
              styles.contextButton,
              selectedContext === context && styles.contextButtonActive,
            ]}
            onPress={() => setSelectedContext(context)}
          >
            <Text style={[
              styles.contextButtonText,
              selectedContext === context && styles.contextButtonTextActive,
            ]}>
              {context}
            </Text>
          </TouchableOpacity>
        ))}
      </ScrollView>
    </View>
  );

  // Render log entry
  const renderLogEntry = (entry: LogEntry, index: number) => {
    const levelColors = {
      [LogLevel.DEBUG]: '#666',
      [LogLevel.INFO]: '#007AFF',
      [LogLevel.WARN]: '#FF9500',
      [LogLevel.ERROR]: '#FF3B30',
    };

    const levelIcons = {
      [LogLevel.DEBUG]: '🐛',
      [LogLevel.INFO]: 'ℹ️',
      [LogLevel.WARN]: '⚠️',
      [LogLevel.ERROR]: '❌',
    };

    return (
      <View key={index} style={styles.logEntry}>
        <View style={styles.logHeader}>
          <Text style={styles.logTimestamp}>
            {entry.timestamp.toLocaleTimeString()}
          </Text>
          <Text style={[styles.logLevel, { color: levelColors[entry.level] }]}>
            {levelIcons[entry.level]} {LogLevel[entry.level]}
          </Text>
          <Text style={styles.logContext}>{entry.context}</Text>
        </View>
        
        <Text style={styles.logMessage}>{entry.message}</Text>
        
        {entry.data && (
          <View style={styles.logData}>
            <Text style={styles.logDataLabel}>Data:</Text>
            <Text style={styles.logDataText}>
              {JSON.stringify(entry.data, null, 2)}
            </Text>
          </View>
        )}
        
        {entry.error && (
          <View style={styles.logError}>
            <Text style={styles.logErrorLabel}>Error:</Text>
            <Text style={styles.logErrorText}>{entry.error.message}</Text>
            {entry.error.stack && (
              <Text style={styles.logStackText}>{entry.error.stack}</Text>
            )}
          </View>
        )}
      </View>
    );
  };

  // Render app state summary
  const renderAppState = () => (
    <View style={styles.stateContainer}>
      <Text style={styles.stateTitle}>App State Summary</Text>
      
      <View style={styles.stateSection}>
        <Text style={styles.stateSectionTitle}>Devices ({devices.devices.length})</Text>
        <Text style={styles.stateText}>
          Connected: {devices.devices.filter(d => d.isConnected).length}
        </Text>
        <Text style={styles.stateText}>
          Selected: {devices.selectedDevice?.name || 'None'}
        </Text>
        <Text style={styles.stateText}>
          Multi-selected: {devices.selectedDevices.length}
        </Text>
        <Text style={styles.stateText}>
          Discovery: {devices.isDiscovering ? 'Active' : 'Inactive'}
        </Text>
      </View>
      
      <View style={styles.stateSection}>
        <Text style={styles.stateSectionTitle}>Control</Text>
        <Text style={styles.stateText}>
          Volume: {Math.round(control.volume * 100)}%
        </Text>
        <Text style={styles.stateText}>
          Muted: {control.isMuted ? 'Yes' : 'No'}
        </Text>
        <Text style={styles.stateText}>
          Preset: {control.currentPreset || 'None'}
        </Text>
        <Text style={styles.stateText}>
          Processing: {control.isProcessingCommand ? 'Yes' : 'No'}
        </Text>
      </View>
      
      <View style={styles.stateSection}>
        <Text style={styles.stateSectionTitle}>Settings</Text>
        <Text style={styles.stateText}>
          Port: {settings.defaultPort}
        </Text>
        <Text style={styles.stateText}>
          Discovery Interval: {settings.discoveryInterval}ms
        </Text>
        <Text style={styles.stateText}>
          Auto Discovery: {settings.autoDiscovery ? 'On' : 'Off'}
        </Text>
        <Text style={styles.stateText}>
          Log Level: {settings.logLevel}
        </Text>
      </View>
    </View>
  );

  return (
    <Modal visible={visible} animationType="slide" presentationStyle="pageSheet">
      <View style={styles.container}>
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.title}>Debug Panel</Text>
          <TouchableOpacity onPress={onClose} style={styles.closeButton}>
            <Text style={styles.closeButtonText}>Done</Text>
          </TouchableOpacity>
        </View>

        {/* Controls */}
        <View style={styles.controls}>
          <View style={styles.controlRow}>
            <View style={styles.switchContainer}>
              <Text style={styles.switchLabel}>Auto Scroll</Text>
              <Switch
                value={autoScroll}
                onValueChange={setAutoScroll}
                trackColor={{ false: '#767577', true: '#81b0ff' }}
                thumbColor={autoScroll ? '#007AFF' : '#f4f3f4'}
              />
            </View>
            
            <TouchableOpacity onPress={handleExportLogs} style={styles.controlButton}>
              <Text style={styles.controlButtonText}>Export</Text>
            </TouchableOpacity>
            
            <TouchableOpacity onPress={handleClearLogs} style={styles.controlButton}>
              <Text style={styles.controlButtonText}>Clear</Text>
            </TouchableOpacity>
          </View>

          {renderLevelSelector()}
          {renderContextSelector()}
        </View>

        {/* Content */}
        <ScrollView 
          style={styles.content}
          showsVerticalScrollIndicator={true}
          ref={(ref) => {
            if (ref && autoScroll && logs.length > 0) {
              setTimeout(() => ref.scrollToEnd({ animated: false }), 100);
            }
          }}
        >
          {renderAppState()}
          
          <View style={styles.logsSection}>
            <Text style={styles.logsSectionTitle}>
              Logs ({logs.length}) - {LogLevel[selectedLevel]}+ {selectedContext || 'All Contexts'}
            </Text>
            
            {logs.length === 0 ? (
              <Text style={styles.noLogsText}>No logs to display</Text>
            ) : (
              logs.map((entry, index) => renderLogEntry(entry, index))
            )}
          </View>
        </ScrollView>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    backgroundColor: 'white',
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  title: {
    fontSize: 20,
    fontWeight: '600',
    color: '#333',
  },
  closeButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    backgroundColor: '#007AFF',
    borderRadius: 6,
  },
  closeButtonText: {
    color: 'white',
    fontSize: 16,
    fontWeight: '600',
  },
  controls: {
    backgroundColor: 'white',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#e0e0e0',
  },
  controlRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  switchContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  switchLabel: {
    fontSize: 14,
    color: '#333',
    marginRight: 8,
  },
  controlButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    backgroundColor: '#f0f0f0',
    borderRadius: 6,
  },
  controlButtonText: {
    fontSize: 14,
    color: '#333',
    fontWeight: '500',
  },
  selectorContainer: {
    marginBottom: 12,
  },
  selectorLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#333',
    marginBottom: 6,
  },
  levelButtons: {
    flexDirection: 'row',
    gap: 8,
  },
  levelButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    backgroundColor: '#f0f0f0',
    borderRadius: 6,
    borderWidth: 1,
    borderColor: 'transparent',
  },
  levelButtonActive: {
    backgroundColor: '#007AFF',
    borderColor: '#0056CC',
  },
  levelButtonText: {
    fontSize: 12,
    color: '#333',
    fontWeight: '500',
  },
  levelButtonTextActive: {
    color: 'white',
  },
  contextButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    backgroundColor: '#f0f0f0',
    borderRadius: 6,
    marginRight: 8,
    borderWidth: 1,
    borderColor: 'transparent',
  },
  contextButtonActive: {
    backgroundColor: '#007AFF',
    borderColor: '#0056CC',
  },
  contextButtonText: {
    fontSize: 12,
    color: '#333',
    fontWeight: '500',
  },
  contextButtonTextActive: {
    color: 'white',
  },
  content: {
    flex: 1,
  },
  stateContainer: {
    backgroundColor: 'white',
    margin: 16,
    padding: 16,
    borderRadius: 8,
  },
  stateTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
    marginBottom: 12,
  },
  stateSection: {
    marginBottom: 12,
  },
  stateSectionTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#007AFF',
    marginBottom: 4,
  },
  stateText: {
    fontSize: 12,
    color: '#666',
    marginLeft: 8,
  },
  logsSection: {
    backgroundColor: 'white',
    margin: 16,
    marginTop: 0,
    padding: 16,
    borderRadius: 8,
  },
  logsSectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginBottom: 12,
  },
  noLogsText: {
    fontSize: 14,
    color: '#999',
    textAlign: 'center',
    paddingVertical: 20,
  },
  logEntry: {
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
    paddingVertical: 8,
  },
  logHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
  },
  logTimestamp: {
    fontSize: 10,
    color: '#999',
    fontFamily: 'monospace',
  },
  logLevel: {
    fontSize: 10,
    fontWeight: '600',
  },
  logContext: {
    fontSize: 10,
    color: '#666',
    fontWeight: '500',
  },
  logMessage: {
    fontSize: 12,
    color: '#333',
    marginBottom: 4,
  },
  logData: {
    backgroundColor: '#f8f9fa',
    padding: 8,
    borderRadius: 4,
    marginBottom: 4,
  },
  logDataLabel: {
    fontSize: 10,
    fontWeight: '600',
    color: '#007AFF',
    marginBottom: 2,
  },
  logDataText: {
    fontSize: 10,
    color: '#333',
    fontFamily: 'monospace',
  },
  logError: {
    backgroundColor: '#fff5f5',
    padding: 8,
    borderRadius: 4,
    borderLeftWidth: 3,
    borderLeftColor: '#FF3B30',
  },
  logErrorLabel: {
    fontSize: 10,
    fontWeight: '600',
    color: '#FF3B30',
    marginBottom: 2,
  },
  logErrorText: {
    fontSize: 10,
    color: '#333',
    fontWeight: '500',
  },
  logStackText: {
    fontSize: 9,
    color: '#666',
    fontFamily: 'monospace',
    marginTop: 4,
  },
});