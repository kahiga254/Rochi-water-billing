import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import '../styles/globals.css';
import { MainLayout } from '@/components/layout/Mainlayout';
import { AuthInitializer } from '../components/auth/AuthInitializer';
import Script from 'next/script';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Water Billing System',
  description: 'Water billing management system',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=yes" />
        <meta name="theme-color" content="#000000" />
        <meta name="apple-mobile-web-app-capable" content="yes" />
        <meta name="apple-mobile-web-app-status-bar-style" content="default" />
      </head>
      <body className={inter.className}>
        {/* Capacitor initialization script */}
        <Script id="capacitor-init" strategy="afterInteractive">
          {`
            // Initialize Capacitor plugins when running natively
            if (typeof window !== 'undefined' && window.Capacitor) {
              console.log('Running on native platform:', window.Capacitor.getPlatform());
              
              // Handle back button on Android
              document.addEventListener('backbutton', function(e) {
                e.preventDefault();
                if (window.history.length > 1) {
                  window.history.back();
                } else {
                  window.Capacitor.Plugins.App.exitApp();
                }
              }, false);
            }
          `}
        </Script>
        
        <AuthInitializer />
        <MainLayout>{children}</MainLayout>
      </body>
    </html>
  );
}