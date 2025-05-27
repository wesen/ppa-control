/**
 * Redux slice for audio control state
 */

import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface ControlState {
  volume: number;
  isMuted: boolean;
  currentPreset: number | null;
  isProcessingCommand: boolean;
  lastCommandStatus: 'success' | 'error' | 'pending' | null;
  errorMessage: string | null;
  feedbackMessages: string[];
}

const initialState: ControlState = {
  volume: 0.5, // 50% volume
  isMuted: false,
  currentPreset: null,
  isProcessingCommand: false,
  lastCommandStatus: null,
  errorMessage: null,
  feedbackMessages: [],
};

const controlSlice = createSlice({
  name: 'control',
  initialState,
  reducers: {
    // Volume control actions
    setVolume: (state, action: PayloadAction<number>) => {
      state.volume = Math.max(0, Math.min(1, action.payload));
      state.isMuted = false; // Unmute when volume is changed
    },
    toggleMute: (state) => {
      state.isMuted = !state.isMuted;
    },
    setMute: (state, action: PayloadAction<boolean>) => {
      state.isMuted = action.payload;
    },

    // Preset control actions
    setCurrentPreset: (state, action: PayloadAction<number | null>) => {
      state.currentPreset = action.payload;
    },

    // Command status actions
    setCommandProcessing: (state, action: PayloadAction<boolean>) => {
      state.isProcessingCommand = action.payload;
      if (action.payload) {
        state.lastCommandStatus = 'pending';
        state.errorMessage = null;
      }
    },
    setCommandStatus: (state, action: PayloadAction<'success' | 'error'>) => {
      state.lastCommandStatus = action.payload;
      state.isProcessingCommand = false;
    },
    setErrorMessage: (state, action: PayloadAction<string>) => {
      state.errorMessage = action.payload;
      state.lastCommandStatus = 'error';
      state.isProcessingCommand = false;
    },
    clearError: (state) => {
      state.errorMessage = null;
      state.lastCommandStatus = null;
    },

    // Feedback actions
    addFeedbackMessage: (state, action: PayloadAction<string>) => {
      state.feedbackMessages.push(action.payload);
      
      // Keep only last 50 messages
      if (state.feedbackMessages.length > 50) {
        state.feedbackMessages = state.feedbackMessages.slice(-50);
      }
    },
    clearFeedbackMessages: (state) => {
      state.feedbackMessages = [];
    },

    // Reset all control state
    resetControlState: (state) => {
      return { ...initialState };
    },
  },
});

export const {
  setVolume,
  toggleMute,
  setMute,
  setCurrentPreset,
  setCommandProcessing,
  setCommandStatus,
  setErrorMessage,
  clearError,
  addFeedbackMessage,
  clearFeedbackMessages,
  resetControlState,
} = controlSlice.actions;

export default controlSlice.reducer;