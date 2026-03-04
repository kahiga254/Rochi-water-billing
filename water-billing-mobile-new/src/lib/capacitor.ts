import { Capacitor } from '@capacitor/core';

// Initialize Capacitor plugins
export const initializeCapacitor = () => {
  if (Capacitor.isNativePlatform()) {
    // Register any custom plugins here
    console.log('Running on native platform');
  }
};