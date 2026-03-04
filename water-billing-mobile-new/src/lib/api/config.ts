// Check if running in Capacitor native environment
const isNative = typeof window !== 'undefined' && 
  !!(window as any).Capacitor?.isNativePlatform?.();

// Get the appropriate API URL based on environment
export const getApiUrl = () => {
  // Default for browser development
  if (!isNative) {
    return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  }

  // For Android emulator
  const platform = (window as any).Capacitor?.getPlatform?.();
  if (platform === 'android') {
    // Check if it's an emulator (you can add more sophisticated detection if needed)
    return 'http://10.0.2.2:8080'; // Android emulator host machine
  }

  // For physical device - you'll need to set this manually
  return 'http://192.168.1.100:8080'; // Replace with your computer's IP
};

export const API_BASE_URL = getApiUrl();