"use client";

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  Home, 
  Users, 
  FileText, 
  CreditCard, 
  Droplets, 
  Settings, 
  LayoutDashboard,
  Camera
} from 'lucide-react';
import { useAuth } from '@/lib/hooks/useAuth';

export function Sidebar() {
  const pathname = usePathname();
  const { user } = useAuth();

  // Admin navigation items
  const adminNavItems = [
    { href: '/admin', label: 'Dashboard', icon: Home },
    { href: '/customers', label: 'Customers', icon: Users },
    { href: '/users', label: 'Users', icon: Users },
    { href: '/bills', label: 'Bills', icon: FileText },
    { href: '/payments', label: 'Payments', icon: CreditCard },
    { href: '/settings', label: 'Settings', icon: Settings },
  ];

  // Reader navigation items
const readerNavItems = [
  { href: '/reader-direct/dashboard', label: 'Meter Reading', icon: Camera },
  { href: '/reader-direct/history', label: 'My Readings', icon: FileText },
  { href: '/reader-direct/settings', label: 'Settings', icon: Settings },
];
  

  // Customer navigation items
  const customerNavItems = [
    { href: '/my-account', label: 'Dashboard', icon: LayoutDashboard },
    { href: '/my-account/bills', label: 'My Bills', icon: FileText },
    { href: '/my-account/payments', label: 'Payments', icon: CreditCard },
  ];

  // Select nav items based on user role
  const getNavItems = () => {
    if (!user) return [];
    switch (user.role) {
      case 'admin':
        return adminNavItems;
      case 'reader':
        return readerNavItems;
      case 'customer_service':
        return customerNavItems;
      default:
        return [];
    }
  };

  const navItems = getNavItems();

  return (
    <aside className="w-64 h-screen bg-white border-r fixed left-0 top-0">
      <div className="p-4 border-b">
        <h1 className="text-xl font-bold text-blue-600">Rochi Pure</h1>
      </div>
      <nav className="p-4">
        {navItems.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-3 py-2 rounded-lg mb-1 ${
                isActive 
                  ? 'bg-blue-50 text-blue-600' 
                  : 'text-gray-700 hover:bg-gray-100'
              }`}
            >
              <item.icon className="w-5 h-5" />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}