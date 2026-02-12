'use client';

import { MainLayout } from '@/components/common/MainLayout';

export default function ReaderLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <MainLayout>{children}</MainLayout>;
}