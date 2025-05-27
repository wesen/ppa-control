/**
 * Custom hook for managing UDP service
 */

import { useEffect, useRef, useCallback } from 'react';
import { logger } from '../utils/logger';
import { useDispatch, useSelector } from 'react-redux';
import { UDPService, UDPServiceCallbacks } from '../services/UDPService';
import { RootState } from '../store';
import {
  addDevice,
  updateDevice,
  startDiscovery,
  stopDiscovery,
  setDiscoveryError,
  setConnectionStatus,
} from '../store/deviceSlice';
import {
  addFeedbackMessage,
  setCommandProcessing,
  setCommandStatus,
  setErrorMessage,
} from '../store/controlSlice';

export const useUDPService = () => {
  const dispatch = useDispatch();
  const settings = useSelector((state: RootState) => state.settings);
  const { autoDiscovery } = settings;
  
  const udpServiceRef = useRef<UDPService | null>(null);

  // Initialize UDP service
  const initializeService = useCallback(() => {
    logger.debug('useUDPService', 'Initializing UDP service...');
    
    if (udpServiceRef.current) {
      logger.debug('useUDPService', 'UDP service already initialized');
      return udpServiceRef.current;
    }

    logger.info('useUDPService', 'Creating new UDP service instance');

    const callbacks: UDPServiceCallbacks = {
      onDeviceDiscovered: (device) => {
        logger.info('useUDPService', 'Device discovered callback', {
          device: device.name,
          address: `${device.address}:${device.port}`,
          isConnected: device.isConnected
        });
        dispatch(addDevice(device));
        dispatch(addFeedbackMessage(`Device discovered: ${device.address}`));
      },
      onDeviceTimeout: (device) => {
        const updatedDevice = { ...device, isConnected: false };
        dispatch(updateDevice(updatedDevice));
        dispatch(addFeedbackMessage(`Device timeout: ${device.address}`));
      },
      onMessageReceived: (message) => {
        if (settings.enablePacketLogging) {
          dispatch(addFeedbackMessage(
            `Received ${message.header?.messageType} from ${message.remoteAddress}`
          ));
        }
      },
      onError: (error) => {
        dispatch(setDiscoveryError(error.message));
        dispatch(setErrorMessage(error.message));
        dispatch(addFeedbackMessage(`Error: ${error.message}`));
      },
    };

    const config = {
      discoveryPort: settings.defaultPort,
      discoveryInterval: settings.discoveryInterval,
      deviceTimeout: settings.deviceTimeout,
      broadcastAddress: settings.broadcastAddress,
    };

    udpServiceRef.current = new UDPService(config, callbacks);
    return udpServiceRef.current;
  }, [dispatch, settings]);

  // Start service
  const startService = useCallback(async () => {
    logger.info('useUDPService', 'Starting UDP service...', { autoDiscovery });
    
    try {
      const service = initializeService();
      logger.debug('useUDPService', 'UDP service initialized, starting...');
      
      await service.start();
      logger.info('useUDPService', 'UDP service started successfully');
      
      if (autoDiscovery) {
        logger.info('useUDPService', 'Auto-discovery enabled, starting discovery...');
        dispatch(startDiscovery());
        service.startDiscovery();
      }
      
      dispatch(setConnectionStatus('connected'));
      dispatch(addFeedbackMessage('UDP service started'));
      logger.info('useUDPService', 'UDP service startup complete');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('useUDPService', 'Failed to start UDP service', { error: errorMessage }, error as Error);
      
      dispatch(setDiscoveryError(errorMessage));
      dispatch(setErrorMessage('Failed to start UDP service'));
    }
  }, [initializeService, autoDiscovery, dispatch]);

  // Stop service
  const stopService = useCallback(() => {
    if (udpServiceRef.current) {
      udpServiceRef.current.stop();
      udpServiceRef.current = null;
    }
    dispatch(stopDiscovery());
    dispatch(setConnectionStatus('disconnected'));
    dispatch(addFeedbackMessage('UDP service stopped'));
  }, [dispatch]);

  // Send volume command
  const sendVolumeCommand = useCallback(async (deviceAddress: string, volume: number) => {
    const service = udpServiceRef.current;
    if (!service) {
      dispatch(setErrorMessage('UDP service not available'));
      return;
    }

    try {
      dispatch(setCommandProcessing(true));
      service.sendVolumeCommand(deviceAddress, volume);
      dispatch(setCommandStatus('success'));
      dispatch(addFeedbackMessage(`Volume set to ${Math.round(volume * 100)}% on ${deviceAddress}`));
    } catch (error) {
      dispatch(setCommandStatus('error'));
      dispatch(setErrorMessage(error instanceof Error ? error.message : 'Unknown error'));
    }
  }, [dispatch]);

  // Send preset command
  const sendPresetCommand = useCallback(async (deviceAddress: string, presetIndex: number) => {
    const service = udpServiceRef.current;
    if (!service) {
      dispatch(setErrorMessage('UDP service not available'));
      return;
    }

    try {
      dispatch(setCommandProcessing(true));
      service.sendPresetCommand(deviceAddress, presetIndex);
      dispatch(setCommandStatus('success'));
      dispatch(addFeedbackMessage(`Preset ${presetIndex} recalled on ${deviceAddress}`));
    } catch (error) {
      dispatch(setCommandStatus('error'));
      dispatch(setErrorMessage(error instanceof Error ? error.message : 'Unknown error'));
    }
  }, [dispatch]);

  // Send ping
  const sendPing = useCallback((deviceAddress: string) => {
    const service = udpServiceRef.current;
    if (service) {
      service.sendPing(deviceAddress);
      dispatch(addFeedbackMessage(`Ping sent to ${deviceAddress}`));
    }
  }, [dispatch]);

  // Add device manually
  const addDeviceManually = useCallback((address: string, port: number = 5001) => {
    const service = udpServiceRef.current;
    if (service) {
      const device = service.addDevice(address, port);
      dispatch(addDevice(device));
      dispatch(addFeedbackMessage(`Device added manually: ${address}:${port}`));
      return device;
    }
    return null;
  }, [dispatch]);

  // Start/stop discovery
  const toggleDiscovery = useCallback((enable: boolean) => {
    const service = udpServiceRef.current;
    if (!service) {
      return;
    }

    if (enable) {
      dispatch(startDiscovery());
      service.startDiscovery();
      dispatch(addFeedbackMessage('Device discovery started'));
    } else {
      dispatch(stopDiscovery());
      service.stopDiscovery();
      dispatch(addFeedbackMessage('Device discovery stopped'));
    }
  }, [dispatch]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      stopService();
    };
  }, [stopService]);

  return {
    startService,
    stopService,
    sendVolumeCommand,
    sendPresetCommand,
    sendPing,
    addDeviceManually,
    toggleDiscovery,
    isServiceRunning: !!udpServiceRef.current,
  };
};