'use client';

import { MainLayout } from '@/components/common/MainLayout';

export default function CustomerLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <MainLayout>{children}</MainLayout>;
}