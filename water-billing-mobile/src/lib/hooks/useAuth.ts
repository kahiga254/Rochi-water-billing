
"use client";
import { useRouter } from 'next/navigation';
import { useCallback } from 'react';
import { useAuthStore } from '@/lib/store/authStore';
import { toast } from 'sonner';

interface ApiResult {
  success: boolean;
  error?: string;
}

export const useAuth = () => {
  const router = useRouter();
  const { 
    user, 
    token, 
    isAuthenticated, 
    isLoading, 
    error,
    setUser, 
    setToken, 
    setLoading, 
    setError,
    logout: storeLogout 
  } = useAuthStore();

  // ✅ SIMPLIFIED LOGIN - Matches your working test exactly
  const login = async (username: string, password: string): Promise<ApiResult> => {
    try {
      setLoading(true);
      setError(null);
      
      // Use direct fetch - this is what works!
      const response = await fetch('http://localhost:8080/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username,
          password
        })
      });
      
      const result = await response.json();
      console.log('📥 Login response:', result);
      
      if (result.success && result.data) {
        const { token, user } = result.data;
        
        setToken(token);
        setUser(user);
        
        toast.success(`Welcome back, ${user.first_name || user.username}!`);
        
        // Redirect based on role
        if (user.role === 'admin') {
          router.push('/dashboard');
        } else if (user.role === 'customer') {
          router.push('/customer');
        } else if (user.role === 'reader') {
          router.push('/reader/dashboard');
        } else {
          router.push('/dashboard');
        }
        
        return { success: true };
      }
      
      throw new Error(result.message || 'Login failed');
      
    } catch (error: any) {
      console.error('❌ Login error:', error);
      const errorMessage = error.message || 'Invalid username or password';
      setError(errorMessage);
      toast.error(errorMessage);
      return { success: false, error: errorMessage };
    } finally {
      setLoading(false);
    }
  };

  const logout = useCallback(async () => {
    try {
      await fetch('http://localhost:8080/api/v1/auth/logout', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      storeLogout();
      router.push('/login');
      toast.info('Logged out successfully');
    }
  }, [router, storeLogout]);

  return {
    user,
    token,
    isAuthenticated,
    isLoading,
    error,
    login,
    logout,
  };
};