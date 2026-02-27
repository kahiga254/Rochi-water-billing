"use client";

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  Home, 
  Users, 
  FileText, 
  CreditCard, 
  Settings, 
  LayoutDashboard,
  Camera,
  X
} from 'lucide-react';
import { useAuth } from '@/lib/hooks/useAuth';

interface SidebarProps {
  isOpen?: boolean;
  onClose?: () => void;
}

export function Sidebar({ isOpen = true, onClose }: SidebarProps) {
  const pathname = usePathname();
  const { user } = useAuth();

  const adminNavItems = [
    { href: '/admin', label: 'Dashboard', icon: Home },
    { href: '/customers', label: 'Customers', icon: Users },
    { href: '/users', label: 'Users', icon: Users },
    { href: '/bills', label: 'Bills', icon: FileText },
    { href: '/payments', label: 'Payments', icon: CreditCard },
    { href: '/settings', label: 'Settings', icon: Settings },
  ];

  const readerNavItems = [
    { href: '/reader-direct/dashboard', label: 'Meter Reading', icon: Camera },
    { href: '/reader-direct/history', label: 'My Readings', icon: FileText },
    { href: '/reader-direct/settings', label: 'Settings', icon: Settings },
  ];

  const customerNavItems = [
    { href: '/my-account', label: 'Dashboard', icon: LayoutDashboard },
    { href: '/my-account/bills', label: 'My Bills', icon: FileText },
    { href: '/my-account/payments', label: 'Payments', icon: CreditCard },
  ];

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

  const NavContent = () => (
    <nav className="p-4">
      {navItems.map((item) => {
        const isActive = pathname === item.href;
        return (
          <Link
            key={item.href}
            href={item.href}
            onClick={onClose}
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
  );

  return (
    <>
      {/* Desktop Sidebar - fixed on the left */}
      <aside className="hidden lg:block fixed left-0 top-0 w-64 h-screen bg-white border-r z-30">
        <div className="p-4 border-b h-16 flex items-center">
          <h1 className="text-xl font-bold text-blue-600">Rochi Pure</h1>
        </div>
        <div className="overflow-y-auto h-[calc(100vh-4rem)]">
          <NavContent />
        </div>
      </aside>

      {/* Mobile Sidebar - overlay */}
      <aside className={`
        lg:hidden fixed left-0 top-0 w-64 h-screen bg-white border-r z-50
        transform transition-transform duration-300
        ${isOpen ? 'translate-x-0' : '-translate-x-full'}
      `}>
        <div className="p-4 border-b h-16 flex items-center justify-between">
          <h1 className="text-xl font-bold text-blue-600">Rochi Pure</h1>
          <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-lg">
            <X size={20} />
          </button>
        </div>
        <div className="overflow-y-auto h-[calc(100vh-4rem)]">
          <NavContent />
        </div>
      </aside>
    </>
  );
}