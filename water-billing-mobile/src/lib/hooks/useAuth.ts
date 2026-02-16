import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/lib/store/authStore';
import { authApi } from '@/lib/api/auth';
import { LoginCredentials } from '@/types';
import { toast } from 'sonner';

export const useAuth = () => {
  const router = useRouter();
  const { user, token, isAuthenticated, isLoading, setUser, setToken, setLoading, logout } = useAuthStore();

  const login = async (credentials: LoginCredentials) => {
    try {
      setLoading(true);
      const response = await authApi.login(credentials);

      
      
      if (response.success) {
        const { token, user } = response.data;

      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
        setToken(token);
        setUser(user);
        
        
        toast.success(`Welcome back, ${user.first_name}!`);
        
        // Redirect based on role - MATCHING YOUR CURRENT FOLDER STRUCTURE
        switch (user.role) {
          case 'admin':
            router.push('/admin');
            break;
          case 'customer_service':
            router.push('/my-account');
            break;
          case 'reader':
            router.push('/reader-direct/dashboard');
            break;
          default:
            router.push('/dashboard');
        }
        
        return { success: true };
      }
      
      return { success: false, error: response.message };
    } catch (error: any) {
      const message = error.response?.data?.message || 'Login failed';
      toast.error(message);
      return { success: false, error: message };
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await authApi.logout();
      toast.info('Logged out successfully');
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      logout();
      router.push('/login');
    }
  };

  const checkAuth = async () => {
    if (!token) return;
    
    try {
      const response = await authApi.getProfile();
      if (response.success) {
        setUser(response.data);
      }
    } catch (error) {
      logout();
    }
  };

  return {
    user,
    token,
    isAuthenticated,
    isLoading,
    login,
    logout: handleLogout,
    checkAuth,
  };
};