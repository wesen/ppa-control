/**
 * Redux slice for app settings
 */

import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface SettingsState {
  // Network settings
  defaultPort: number;
  discoveryInterval: number;
  deviceTimeout: number;
  broadcastAddress: string;
  
  // UI settings
  theme: 'light' | 'dark' | 'auto';
  enableHapticFeedback: boolean;
  enableNotifications: boolean;
  
  // Debug settings
  logLevel: 'debug' | 'info' | 'warn' | 'error';
  enablePacketLogging: boolean;
  
  // App preferences
  autoDiscovery: boolean;
  keepScreenOn: boolean;
  confirmCommands: boolean;
}

const initialState: SettingsState = {
  // Network settings
  defaultPort: 5001,
  discoveryInterval: 5000,
  deviceTimeout: 30000,
  broadcastAddress: '255.255.255.255',
  
  // UI settings
  theme: 'auto',
  enableHapticFeedback: true,
  enableNotifications: true,
  
  // Debug settings
  logLevel: 'info',
  enablePacketLogging: false,
  
  // App preferences
  autoDiscovery: true,
  keepScreenOn: false,
  confirmCommands: false,
};

const settingsSlice = createSlice({
  name: 'settings',
  initialState,
  reducers: {
    // Network settings
    setDefaultPort: (state, action: PayloadAction<number>) => {
      state.defaultPort = action.payload;
    },
    setDiscoveryInterval: (state, action: PayloadAction<number>) => {
      state.discoveryInterval = action.payload;
    },
    setDeviceTimeout: (state, action: PayloadAction<number>) => {
      state.deviceTimeout = action.payload;
    },
    setBroadcastAddress: (state, action: PayloadAction<string>) => {
      state.broadcastAddress = action.payload;
    },

    // UI settings
    setTheme: (state, action: PayloadAction<'light' | 'dark' | 'auto'>) => {
      state.theme = action.payload;
    },
    setEnableHapticFeedback: (state, action: PayloadAction<boolean>) => {
      state.enableHapticFeedback = action.payload;
    },
    setEnableNotifications: (state, action: PayloadAction<boolean>) => {
      state.enableNotifications = action.payload;
    },

    // Debug settings
    setLogLevel: (state, action: PayloadAction<'debug' | 'info' | 'warn' | 'error'>) => {
      state.logLevel = action.payload;
    },
    setEnablePacketLogging: (state, action: PayloadAction<boolean>) => {
      state.enablePacketLogging = action.payload;
    },

    // App preferences
    setAutoDiscovery: (state, action: PayloadAction<boolean>) => {
      state.autoDiscovery = action.payload;
    },
    setKeepScreenOn: (state, action: PayloadAction<boolean>) => {
      state.keepScreenOn = action.payload;
    },
    setConfirmCommands: (state, action: PayloadAction<boolean>) => {
      state.confirmCommands = action.payload;
    },

    // Reset settings
    resetSettings: (state) => {
      return { ...initialState };
    },
  },
});

export const {
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
} = settingsSlice.actions;

export default settingsSlice.reducer;