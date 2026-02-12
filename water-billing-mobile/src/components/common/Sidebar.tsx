'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  Home, 
  Droplets, 
  FileText, 
  CreditCard, 
  Users, 
  Gauge, 
  MapPin,
  Settings,
  BarChart3,
  Camera,
  Receipt,
  Zap,
  X,
  User
} from 'lucide-react';
import { cn } from '@/lib/utils/cn';

// Temporarily use mock data until useAuth is ready
// Replace this with the actual import when ready
// import { useAuth } from '@/lib/hooks/useAuth';

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

interface NavItem {
  title: string;
  href: string;
  icon: React.ElementType;
  roles?: string[];
  children?: NavItem[];
}

// Mock user for development
const mockUser = {
  first_name: 'John',
  last_name: 'Doe',
  role: 'admin', // Change this to test different roles: 'admin', 'reader', 'customer', etc.
  email: 'john.doe@example.com',
};

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const pathname = usePathname();
  
  // Use mock user for now, replace with actual hook when ready
  // const { user } = useAuth();
  const user = mockUser;

  // Navigation items based on user role
  const navItems: NavItem[] = [
    {
      title: 'Dashboard',
      href: '/dashboard',
      icon: Home,
      roles: ['admin', 'manager', 'reader', 'cashier', 'customer_service', 'customer'],
    },
    {
      title: 'Water Consumption',
      href: '/consumption',
      icon: Droplets,
      roles: ['customer'],
    },
    {
      title: 'My Bills',
      href: '/bills',
      icon: FileText,
      roles: ['customer'],
    },
    {
      title: 'Make Payment',
      href: '/payments',
      icon: CreditCard,
      roles: ['customer'],
    },
    {
      title: 'Meter Reading',
      href: '/readings',
      icon: Camera,
      roles: ['reader'],
    },
    {
      title: 'Customers',
      href: '/customers',
      icon: Users,
      roles: ['admin', 'manager', 'reader', 'customer_service'],
    },
    {
      title: 'Zones',
      href: '/zones',
      icon: MapPin,
      roles: ['admin', 'manager'],
    },
    {
      title: 'Reports',
      href: '/reports',
      icon: BarChart3,
      roles: ['admin', 'manager'],
    },
    {
      title: 'Billing',
      href: '/billing',
      icon: Receipt,
      roles: ['admin', 'manager', 'cashier'],
    },
    {
      title: 'Performance',
      href: '/performance',
      icon: Gauge,
      roles: ['admin', 'manager'],
    },
    {
      title: 'Settings',
      href: '/settings',
      icon: Settings,
      roles: ['admin', 'manager', 'reader', 'cashier', 'customer_service', 'customer'],
    },
  ];

  // Filter navigation items based on user role
  const filteredNavItems = navItems.filter(item => {
    if (!user) return false;
    if (user.role === 'admin') return true; // Admin sees everything
    return item.roles?.includes(user.role);
  });

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={onClose}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          'fixed top-0 left-0 z-50 h-full w-64 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 transition-transform duration-300 ease-in-out',
          isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        )}
      >
        {/* Logo area */}
        <div className="flex items-center justify-between h-16 px-4 border-b border-gray-200 dark:border-gray-800">
          <Link href="/dashboard" className="flex items-center gap-2">
            <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
              <Zap className="w-5 h-5 text-white" />
            </div>
            <span className="text-lg font-bold text-gray-900 dark:text-white">
              Water Billing
            </span>
          </Link>
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 lg:hidden"
          >
            <X className="w-5 h-5 text-gray-700 dark:text-gray-300" />
          </button>
        </div>

        {/* Navigation */}
        <nav className="p-4 space-y-1 overflow-y-auto h-[calc(100%-4rem)]">
          {filteredNavItems.map((item) => {
            const isActive = pathname === item.href || pathname.startsWith(`${item.href}/`);
            
            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={onClose}
                className={cn(
                  'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-blue-50 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
                )}
              >
                <item.icon className={cn(
                  'w-5 h-5',
                  isActive ? 'text-blue-600 dark:text-blue-400' : 'text-gray-500 dark:text-gray-400'
                )} />
                {item.title}
              </Link>
            );
          })}
        </nav>

        {/* User info - mobile only */}
        {user && (
          <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center">
                <User className="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                  {user.first_name} {user.last_name}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400 truncate capitalize">
                  {user.role}
                </p>
              </div>
            </div>
          </div>
        )}
      </aside>
    </>
  );
}