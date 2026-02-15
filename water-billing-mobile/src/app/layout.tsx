import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import '../styles/globals.css';
import { MainLayout } from '@/components/layout/Mainlayout';
import { AuthInitializer } from '../components/auth/AuthInitializer';

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
      <body className={inter.className}>
        <AuthInitializer />
        <MainLayout>{children}</MainLayout>
      </body>
    </html>
  );
}