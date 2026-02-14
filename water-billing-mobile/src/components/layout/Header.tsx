"use client";

import { usePathname } from 'next/navigation';
import { Bell, User } from 'lucide-react';

export function Header() {
  const pathname = usePathname();
  const pageName = pathname.split('/').filter(Boolean).pop() || 'Dashboard';
  const title = pageName.charAt(0).toUpperCase() + pageName.slice(1);

  return (
    <header className="h-16 bg-white border-b flex items-center justify-between px-6 fixed top-0 left-64 right-0">
      <h1 className="text-xl font-semibold">{title}</h1>
      <div className="flex items-center gap-4">
        <button className="p-2 hover:bg-gray-100 rounded-lg">
          <Bell className="w-5 h-5" />
        </button>
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
            <User className="w-4 h-4 text-blue-600" />
          </div>
          <span className="text-sm font-medium">Admin</span>
        </div>
      </div>
    </header>
  );
}