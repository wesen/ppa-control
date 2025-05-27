/**
 * Redux slice for device management
 */

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { DeviceInfo } from '../protocol/types';

export interface DeviceState {
  devices: DeviceInfo[];
  selectedDevice: DeviceInfo | null;
  selectedDevices: string[]; // For multi-device control
  isDiscovering: boolean;
  discoveryError: string | null;
  connectionStatus: 'disconnected' | 'connecting' | 'connected';
}

const initialState: DeviceState = {
  devices: [],
  selectedDevice: null,
  selectedDevices: [],
  isDiscovering: false,
  discoveryError: null,
  connectionStatus: 'disconnected',
};

const deviceSlice = createSlice({
  name: 'devices',
  initialState,
  reducers: {
    // Device discovery actions
    startDiscovery: (state) => {
      state.isDiscovering = true;
      state.discoveryError = null;
    },
    stopDiscovery: (state) => {
      state.isDiscovering = false;
    },
    setDiscoveryError: (state, action: PayloadAction<string>) => {
      state.discoveryError = action.payload;
      state.isDiscovering = false;
    },

    // Device management actions
    addDevice: (state, action: PayloadAction<DeviceInfo>) => {
      const deviceIndex = state.devices.findIndex(
        d => `${d.address}:${d.port}` === `${action.payload.address}:${action.payload.port}`
      );
      
      if (deviceIndex >= 0) {
        // Update existing device
        state.devices[deviceIndex] = action.payload;
      } else {
        // Add new device
        state.devices.push(action.payload);
      }
    },
    removeDevice: (state, action: PayloadAction<string>) => {
      const deviceAddress = action.payload;
      state.devices = state.devices.filter(
        d => `${d.address}:${d.port}` !== deviceAddress
      );
      
      // Clear selection if removed device was selected
      if (state.selectedDevice && `${state.selectedDevice.address}:${state.selectedDevice.port}` === deviceAddress) {
        state.selectedDevice = null;
      }
      
      // Remove from multi-selection
      state.selectedDevices = state.selectedDevices.filter(addr => addr !== deviceAddress);
    },
    updateDevice: (state, action: PayloadAction<DeviceInfo>) => {
      const deviceIndex = state.devices.findIndex(
        d => `${d.address}:${d.port}` === `${action.payload.address}:${action.payload.port}`
      );
      
      if (deviceIndex >= 0) {
        state.devices[deviceIndex] = action.payload;
        
        // Update selected device if it's the same
        if (state.selectedDevice && 
            `${state.selectedDevice.address}:${state.selectedDevice.port}` === 
            `${action.payload.address}:${action.payload.port}`) {
          state.selectedDevice = action.payload;
        }
      }
    },
    clearDevices: (state) => {
      state.devices = [];
      state.selectedDevice = null;
      state.selectedDevices = [];
    },

    // Device selection actions
    selectDevice: (state, action: PayloadAction<DeviceInfo>) => {
      state.selectedDevice = action.payload;
      state.connectionStatus = 'connecting';
    },
    deselectDevice: (state) => {
      state.selectedDevice = null;
      state.connectionStatus = 'disconnected';
    },
    setConnectionStatus: (state, action: PayloadAction<'disconnected' | 'connecting' | 'connected'>) => {
      state.connectionStatus = action.payload;
    },

    // Multi-device selection actions
    toggleDeviceSelection: (state, action: PayloadAction<string>) => {
      const deviceAddress = action.payload;
      const index = state.selectedDevices.indexOf(deviceAddress);
      
      if (index >= 0) {
        state.selectedDevices.splice(index, 1);
      } else {
        state.selectedDevices.push(deviceAddress);
      }
    },
    selectAllDevices: (state) => {
      state.selectedDevices = state.devices
        .filter(d => d.isConnected)
        .map(d => `${d.address}:${d.port}`);
    },
    clearDeviceSelection: (state) => {
      state.selectedDevices = [];
    },
  },
});

export const {
  startDiscovery,
  stopDiscovery,
  setDiscoveryError,
  addDevice,
  removeDevice,
  updateDevice,
  clearDevices,
  selectDevice,
  deselectDevice,
  setConnectionStatus,
  toggleDeviceSelection,
  selectAllDevices,
  clearDeviceSelection,
} = deviceSlice.actions;

export default deviceSlice.reducer;