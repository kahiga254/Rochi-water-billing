'use client';

import { useRouter } from 'next/navigation';
import { Shield, ArrowLeft, Home } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/Card';
import { useAuth } from '@/lib/hooks/useAuth';

export default function UnauthorizedPage() {
  const router = useRouter();
  const { user } = useAuth();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 p-4">
      <Card className="w-full max-w-md">
        <CardContent className="p-8 text-center">
          <div className="w-20 h-20 bg-red-100 dark:bg-red-900/50 rounded-full flex items-center justify-center mx-auto mb-6">
            <Shield className="w-10 h-10 text-red-600 dark:text-red-400" />
          </div>
          
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            Access Denied
          </h1>
          
          <p className="text-gray-600 dark:text-gray-400 mb-6">
            You don't have permission to access this page.
            {user && (
              <span className="block mt-2 text-sm">
                Your role: <span className="font-mono font-medium">{user.role}</span>
              </span>
            )}
          </p>

          <div className="flex flex-col sm:flex-row gap-3">
            <button
              onClick={() => router.back()}
              className="flex-1 inline-flex items-center justify-center gap-2 px-4 py-2 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg font-medium transition-colors"
            >
              <ArrowLeft className="w-4 h-4" />
              Go Back
            </button>
            
            <button
              onClick={() => {
                if (user?.role === 'customer') router.push('/customer/dashboard');
                else if (user?.role === 'admin') router.push('/admin/dashboard');
                else if (user?.role === 'reader') router.push('/reader/dashboard');
                else router.push('/dashboard');
              }}
              className="flex-1 inline-flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
            >
              <Home className="w-4 h-4" />
              Go to Dashboard
            </button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}