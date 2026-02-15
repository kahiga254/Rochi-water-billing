"use client";

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/hooks/useAuth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  allowedRoles?: string[];
}

export function ProtectedRoute({ children, allowedRoles }: ProtectedRouteProps) {
  const router = useRouter();
  const { user, token, isLoading, checkAuth } = useAuth();
  const [isAuthorized, setIsAuthorized] = useState(false);

  useEffect(() => {
    const verifyAuth = async () => {
      // Check if token exists in localStorage directly
      const storedToken = localStorage.getItem('token');
      
      if (!storedToken) {
        console.log('No token found, redirecting to login');
        router.replace('/login');
        return;
      }

      // If we have a token but no user, try to check auth
      if (storedToken && !user) {
        await checkAuth();
      }

      // Check role authorization
      if (user) {
        if (allowedRoles && !allowedRoles.includes(user.role)) {
          console.log(`Role ${user.role} not authorized`);
          router.replace('/unauthorized');
          return;
        }
        setIsAuthorized(true);
      }
    };

    if (!isLoading) {
      verifyAuth();
    }
  }, [isLoading, user, token, router, allowedRoles, checkAuth]);

  // Show loading while checking
  if (isLoading || !isAuthorized) {
    return (
      <div className="h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}