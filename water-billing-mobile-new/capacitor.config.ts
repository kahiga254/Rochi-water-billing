import { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.yourcompany.waterbilling',
  appName: 'Water Billing',
  webDir: '.next',
  android: {
    allowMixedContent: true
  }
};

export default config;
