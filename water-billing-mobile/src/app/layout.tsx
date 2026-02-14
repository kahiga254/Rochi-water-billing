import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import '../styles/globals.css'; // Use relative path, not @/
import { MainLayout } from '@/components/layout/Mainlayout';

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
        <MainLayout>{children}</MainLayout>
      </body>
    </html>
  );
}