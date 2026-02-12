import type { Metadata, Viewport } from 'next';
import { Inter } from 'next/font/google';
import '@/styles/globals.css';
import { ThemeProvider } from '@/components/common/ThemeProvider';
import { MainLayout } from '@/components/common/MainLayout';
import { Toaster } from 'sonner';

const inter = Inter({ subsets: ['latin'] });

// ✅ SEPARATE VIEWPORT EXPORT - Move all viewport-related settings here
export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#3b82f6' },
    { media: '(prefers-color-scheme: dark)', color: '#2563eb' },
  ],
  colorScheme: 'light dark',
};

// ✅ METADATA - Only non-viewport metadata here
export const metadata: Metadata = {
  title: 'Water Billing System',
  description: 'Mobile application for water billing management',
  manifest: '/manifest.json',
  appleWebApp: {
    capable: true,
    statusBarStyle: 'default',
    title: 'Water Billing',
  },
  formatDetection: {
    telephone: true,
    email: true,
    address: true,
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <ThemeProvider>
          <MainLayout>{children}</MainLayout>
          <Toaster 
            position="top-center"
            toastOptions={{
              className: 'text-sm',
              duration: 4000,
            }}
          />
        </ThemeProvider>
      </body>
    </html>
  );
}