/**
 * UDP Communication Service for PPA Protocol
 * Handles device discovery, communication, and message processing
 */

import { logger } from '../utils/logger';

// Try to import react-native-udp with fallback handling
let dgram: any = null;
try {
  dgram = require('react-native-udp');
  logger.info('UDPService', 'react-native-udp loaded successfully');
} catch (error) {
  logger.error('UDPService', 'Failed to load react-native-udp - UDP functionality will be disabled', null, error as Error);
}
import {
  DeviceInfo,
  ReceivedMessage,
  MessageType,
  StatusType,
} from '../protocol/types';
import {
  createPingMessage,
  createVolumeMessage,
  createPresetRecallMessage,
  parseHeader,
} from '../protocol/encoding';

export interface UDPServiceConfig {
  discoveryPort: number;
  discoveryInterval: number;
  deviceTimeout: number;
  broadcastAddress: string;
}

export interface UDPServiceCallbacks {
  onDeviceDiscovered?: (device: DeviceInfo) => void;
  onDeviceTimeout?: (device: DeviceInfo) => void;
  onMessageReceived?: (message: ReceivedMessage) => void;
  onError?: (error: Error) => void;
}

export class UDPService {
  private socket: any = null;
  private sequenceNumber: number = 1;
  private discoveryTimer: NodeJS.Timeout | null = null;
  private devices: Map<string, DeviceInfo> = new Map();
  private callbacks: UDPServiceCallbacks = {};
  private config: UDPServiceConfig;
  private isRunning: boolean = false;

  constructor(
    config: Partial<UDPServiceConfig> = {},
    callbacks: UDPServiceCallbacks = {},
  ) {
    this.config = {
      discoveryPort: 5001,
      discoveryInterval: 5000, // 5 seconds
      deviceTimeout: 30000, // 30 seconds
      broadcastAddress: '255.255.255.255',
      ...config,
    };
    this.callbacks = callbacks;
  }

  /**
   * Start the UDP service
   */
  async start(): Promise<void> {
    logger.info('UDPService', 'Starting UDP service...', {
      config: this.config,
      isRunning: this.isRunning
    });

    if (this.isRunning) {
      logger.warn('UDPService', 'UDP service already running');
      return;
    }

    if (!dgram) {
      const error = new Error('react-native-udp not available - requires development build');
      logger.error('UDPService', 'Cannot start UDP service without native module', null, error);
      throw error;
    }

    try {
      logger.debug('UDPService', 'Creating UDP socket...');
      
      // Create UDP socket
      this.socket = dgram.createSocket({
        type: 'udp4',
        reusePort: true,
      });

      if (!this.socket) {
        throw new Error('Failed to create UDP socket - dgram.createSocket returned null');
      }

      logger.info('UDPService', 'UDP socket created successfully');

      // Set up socket event listeners
      this.setupSocketListeners();

      // Bind to random port for sending
      logger.debug('UDPService', 'Binding socket to random port...');
      
      await new Promise<void>((resolve, reject) => {
        this.socket.bind(0, (err: any) => {
          if (err) {
            logger.error('UDPService', 'Failed to bind socket', { error: err.message }, err);
            reject(new Error(`Failed to bind socket: ${err.message}`));
          } else {
            logger.info('UDPService', 'Socket bound successfully');
            resolve();
          }
        });
      });

      // Enable broadcast
      logger.debug('UDPService', 'Enabling broadcast mode...');
      this.socket.setBroadcast(true);
      logger.info('UDPService', 'Broadcast mode enabled');

      this.isRunning = true;
      logger.info('UDPService', 'UDP Service started successfully');

      // Start device discovery
      this.startDiscovery();
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      logger.error('UDPService', 'Failed to start UDP service', { error: errorMsg }, error as Error);
      this.handleError(new Error(`Failed to start UDP service: ${errorMsg}`));
      throw error;
    }
  }

  /**
   * Stop the UDP service
   */
  stop(): void {
    logger.info('UDPService', 'Stopping UDP service...', {
      isRunning: this.isRunning,
      deviceCount: this.devices.size
    });

    if (!this.isRunning) {
      logger.warn('UDPService', 'UDP service not running');
      return;
    }

    this.stopDiscovery();

    if (this.socket) {
      logger.debug('UDPService', 'Closing UDP socket...');
      this.socket.close();
      this.socket = null;
      logger.info('UDPService', 'UDP socket closed');
    }

    logger.debug('UDPService', `Clearing ${this.devices.size} devices from cache`);
    this.devices.clear();
    this.isRunning = false;
    logger.info('UDPService', 'UDP Service stopped');
  }

  /**
   * Start automatic device discovery
   */
  startDiscovery(): void {
    logger.info('UDPService', 'Starting device discovery...', {
      interval: this.config.discoveryInterval,
      broadcastAddress: this.config.broadcastAddress,
      port: this.config.discoveryPort
    });

    if (this.discoveryTimer) {
      logger.warn('UDPService', 'Discovery timer already running');
      return;
    }

    // Send initial discovery ping
    this.sendDiscoveryPing();

    // Set up periodic discovery
    this.discoveryTimer = setInterval(() => {
      this.sendDiscoveryPing();
      this.checkDeviceTimeouts();
    }, this.config.discoveryInterval);

    logger.info('UDPService', 'Device discovery started');
  }

  /**
   * Stop automatic device discovery
   */
  stopDiscovery(): void {
    if (this.discoveryTimer) {
      clearInterval(this.discoveryTimer);
      this.discoveryTimer = null;
      console.log('Device discovery stopped');
    }
  }

  /**
   * Send volume command to a specific device
   */
  sendVolumeCommand(deviceAddress: string, volume: number): void {
    logger.info('UDPService', 'Sending volume command', {
      deviceAddress,
      volume,
      volumePercent: Math.round(volume * 100)
    });

    if (!this.isRunning || !this.socket) {
      const error = new Error('UDP service not running');
      logger.error('UDPService', 'Cannot send volume command - service not running', {
        isRunning: this.isRunning,
        hasSocket: !!this.socket
      }, error);
      this.handleError(error);
      return;
    }

    const [host, portStr] = deviceAddress.split(':');
    const port = parseInt(portStr, 10) || this.config.discoveryPort;

    logger.debug('UDPService', 'Creating volume message', {
      host,
      port,
      sequenceNumber: this.sequenceNumber + 1
    });

    const message = createVolumeMessage(volume, this.getNextSequenceNumber());
    const buffer = Buffer.from(message);

    logger.udpPacket('sent', deviceAddress, 'VolumeCommand', {
      volume,
      bufferLength: buffer.length
    });

    this.socket.send(buffer, 0, buffer.length, port, host, (err: any) => {
      if (err) {
        logger.error('UDPService', 'Failed to send volume command', {
          deviceAddress,
          volume,
          error: err.message
        }, err);
        this.handleError(new Error(`Failed to send volume command: ${err.message}`));
      } else {
        logger.info('UDPService', 'Volume command sent successfully', {
          deviceAddress,
          volume,
          volumePercent: Math.round(volume * 100)
        });
      }
    });
  }

  /**
   * Send preset recall command to a specific device
   */
  sendPresetCommand(deviceAddress: string, presetIndex: number): void {
    if (!this.isRunning || !this.socket) {
      this.handleError(new Error('UDP service not running'));
      return;
    }

    const [host, portStr] = deviceAddress.split(':');
    const port = parseInt(portStr, 10) || this.config.discoveryPort;

    const message = createPresetRecallMessage(presetIndex, this.getNextSequenceNumber());
    const buffer = Buffer.from(message);

    this.socket.send(buffer, 0, buffer.length, port, host, (err: any) => {
      if (err) {
        this.handleError(new Error(`Failed to send preset command: ${err.message}`));
      } else {
        console.log(`Preset ${presetIndex} command sent to ${deviceAddress}`);
      }
    });
  }

  /**
   * Send ping to a specific device
   */
  sendPing(deviceAddress: string): void {
    if (!this.isRunning || !this.socket) {
      this.handleError(new Error('UDP service not running'));
      return;
    }

    const [host, portStr] = deviceAddress.split(':');
    const port = parseInt(portStr, 10) || this.config.discoveryPort;

    const message = createPingMessage(this.getNextSequenceNumber());
    const buffer = Buffer.from(message);

    this.socket.send(buffer, 0, buffer.length, port, host, (err: any) => {
      if (err) {
        this.handleError(new Error(`Failed to send ping: ${err.message}`));
      } else {
        console.log(`Ping sent to ${deviceAddress}`);
      }
    });
  }

  /**
   * Get list of discovered devices
   */
  getDevices(): DeviceInfo[] {
    return Array.from(this.devices.values());
  }

  /**
   * Get device by address
   */
  getDevice(address: string): DeviceInfo | undefined {
    return this.devices.get(address);
  }

  /**
   * Add device manually (for manual IP entry)
   */
  addDevice(address: string, port: number = 5001): DeviceInfo {
    const fullAddress = `${address}:${port}`;
    const device: DeviceInfo = {
      address,
      port,
      uniqueId: new Uint8Array([0, 0, 0, 0]), // Unknown until ping response
      name: `Device ${address}`,
      lastSeen: new Date(),
      isConnected: false,
    };

    this.devices.set(fullAddress, device);
    
    // Send ping to verify device
    this.sendPing(fullAddress);
    
    if (this.callbacks.onDeviceDiscovered) {
      this.callbacks.onDeviceDiscovered(device);
    }

    return device;
  }

  /**
   * Set up socket event listeners
   */
  private setupSocketListeners(): void {
    logger.debug('UDPService', 'Setting up socket event listeners...');
    
    this.socket.on('message', (buffer: Buffer, rinfo: any) => {
      logger.debug('UDPService', 'Received UDP message', {
        from: `${rinfo.address}:${rinfo.port}`,
        size: buffer.length
      });
      this.handleReceivedMessage(buffer, rinfo);
    });

    this.socket.on('error', (err: any) => {
      logger.error('UDPService', 'UDP socket error', {
        error: err.message,
        code: err.code,
        errno: err.errno
      }, err);
      this.handleError(new Error(`Socket error: ${err.message}`));
    });

    this.socket.on('close', () => {
      logger.info('UDPService', 'UDP socket closed');
    });

    this.socket.on('listening', () => {
      const address = this.socket.address();
      logger.info('UDPService', 'UDP socket listening', {
        address: address.address,
        port: address.port,
        family: address.family
      });
    });

    logger.info('UDPService', 'Socket event listeners configured');
  }

  /**
   * Handle received UDP messages
   */
  private handleReceivedMessage(buffer: Buffer, rinfo: any): void {
    try {
      const arrayBuffer: ArrayBuffer = buffer.buffer.slice(
        buffer.byteOffset,
        buffer.byteOffset + buffer.byteLength,
      ) as ArrayBuffer;

      // Parse header
      let header;
      try {
        header = parseHeader(arrayBuffer);
      } catch (error) {
        console.warn('Failed to parse message header:', error);
        return;
      }

      const message: ReceivedMessage = {
        buffer: arrayBuffer,
        remoteAddress: rinfo.address,
        remotePort: rinfo.port,
        header,
      };

      // Handle different message types
      this.processMessage(message);

      // Notify callback
      if (this.callbacks.onMessageReceived) {
        this.callbacks.onMessageReceived(message);
      }
    } catch (error) {
      this.handleError(new Error(`Failed to process received message: ${error}`));
    }
  }

  /**
   * Process received messages based on type
   */
  private processMessage(message: ReceivedMessage): void {
    if (!message.header) {
      return;
    }

    const deviceAddress = `${message.remoteAddress}:${message.remotePort}`;
    
    // Update or create device entry
    const existingDevice = this.devices.get(deviceAddress);
    if (existingDevice) {
      existingDevice.lastSeen = new Date();
      existingDevice.isConnected = true;
      if (message.header.deviceUniqueId) {
        existingDevice.uniqueId = message.header.deviceUniqueId;
      }
    } else if (message.header.messageType === MessageType.Ping) {
      // New device discovered via ping response
      const newDevice: DeviceInfo = {
        address: message.remoteAddress,
        port: message.remotePort,
        uniqueId: message.header.deviceUniqueId,
        name: `Speaker ${message.remoteAddress}`,
        lastSeen: new Date(),
        isConnected: true,
      };
      
      this.devices.set(deviceAddress, newDevice);
      
      if (this.callbacks.onDeviceDiscovered) {
        this.callbacks.onDeviceDiscovered(newDevice);
      }
    }

    console.log(
      `Received ${MessageType[message.header.messageType]} from ${deviceAddress}`,
    );
  }

  /**
   * Send discovery ping broadcast
   */
  private sendDiscoveryPing(): void {
    if (!this.socket) {
      return;
    }

    const message = createPingMessage(this.getNextSequenceNumber());
    const buffer = Buffer.from(message);

    this.socket.send(
      buffer,
      0,
      buffer.length,
      this.config.discoveryPort,
      this.config.broadcastAddress,
      (err: any) => {
        if (err) {
          this.handleError(new Error(`Failed to send discovery ping: ${err.message}`));
        } else {
          console.log('Discovery ping broadcast sent');
        }
      },
    );
  }

  /**
   * Check for device timeouts
   */
  private checkDeviceTimeouts(): void {
    const now = new Date();
    const timeoutThreshold = this.config.deviceTimeout;

    for (const [address, device] of this.devices.entries()) {
      const timeSinceLastSeen = now.getTime() - device.lastSeen.getTime();
      
      if (timeSinceLastSeen > timeoutThreshold && device.isConnected) {
        device.isConnected = false;
        console.log(`Device ${address} timed out`);
        
        if (this.callbacks.onDeviceTimeout) {
          this.callbacks.onDeviceTimeout(device);
        }
      }
    }
  }

  /**
   * Get next sequence number
   */
  private getNextSequenceNumber(): number {
    this.sequenceNumber = (this.sequenceNumber + 1) % 65536;
    return this.sequenceNumber;
  }

  /**
   * Handle errors
   */
  private handleError(error: Error): void {
    console.error('UDP Service Error:', error.message);
    
    if (this.callbacks.onError) {
      this.callbacks.onError(error);
    }
  }
}