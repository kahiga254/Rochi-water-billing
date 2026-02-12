'use client';

import { useState, useEffect } from 'react';
import { usePathname } from 'next/navigation';
import Link from 'next/link';
import { 
  Home, 
  FileText, 
  CreditCard, 
  User 
} from 'lucide-react';
import { Header } from './Header';
import { Sidebar } from './Sidebar';
import { Loading } from './Loading';
import { cn } from '@/lib/utils/cn';

interface MainLayoutProps {
  children: React.ReactNode;
}

export function MainLayout({ children }: MainLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const pathname = usePathname();

  // Close sidebar on route change (mobile)
  useEffect(() => {
    setSidebarOpen(false);
  }, [pathname]);

  // Simulate loading state
  useEffect(() => {
    const timer = setTimeout(() => setIsLoading(false), 500);
    return () => clearTimeout(timer);
  }, []);

  // Don't show layout on auth pages
  const isAuthPage = pathname?.startsWith('/login') || 
                    pathname?.startsWith('/register') || 
                    pathname === '/';

  if (isAuthPage) {
    return <>{children}</>;
  }

  if (isLoading) {
    return <Loading fullScreen />;
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950">
      <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      
      <div className={cn(
        'transition-all duration-300 lg:ml-64',
        'flex flex-col min-h-screen'
      )}>
        <Header onMenuClick={() => setSidebarOpen(true)} />
        
        <main className="flex-1 p-4 md:p-6">
          <div className="max-w-7xl mx-auto">
            {children}
          </div>
        </main>

        {/* Mobile bottom navigation */}
        <nav className="lg:hidden fixed bottom-0 left-0 right-0 bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-800">
          <div className="flex justify-around items-center h-16">
            <BottomNavItem href="/dashboard" icon={Home} label="Home" />
            <BottomNavItem href="/bills" icon={FileText} label="Bills" />
            <BottomNavItem href="/payments" icon={CreditCard} label="Pay" />
            <BottomNavItem href="/profile" icon={User} label="Profile" />
          </div>
        </nav>

        {/* Bottom padding for mobile navigation */}
        <div className="lg:hidden h-16" />
      </div>
    </div>
  );
}

// Bottom navigation item for mobile
function BottomNavItem({ href, icon: Icon, label }: { href: string; icon: any; label: string }) {
  const pathname = usePathname();
  const isActive = pathname === href;

  return (
    <Link
      href={href}
      className={cn(
        'flex flex-col items-center justify-center p-2 rounded-lg',
        isActive ? 'text-blue-600 dark:text-blue-400' : 'text-gray-600 dark:text-gray-400'
      )}
    >
      <Icon className="w-5 h-5" />
      <span className="text-xs mt-1">{label}</span>
    </Link>
  );
}