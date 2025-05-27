/**
 * Redux store configuration
 */

import { configureStore } from '@reduxjs/toolkit';
import deviceReducer from './deviceSlice';
import controlReducer from './controlSlice';
import settingsReducer from './settingsSlice';

export const store = configureStore({
  reducer: {
    devices: deviceReducer,
    control: controlReducer,
    settings: settingsReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types
        ignoredActions: ['devices/addDevice', 'devices/updateDevice'],
        // Ignore these field paths in all actions
        ignoredActionsPaths: ['payload.uniqueId', 'payload.lastSeen'],
        // Ignore these paths in the state
        ignoredPaths: ['devices.devices.uniqueId', 'devices.devices.lastSeen'],
      },
    }),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;