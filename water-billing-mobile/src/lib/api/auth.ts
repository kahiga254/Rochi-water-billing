import apiClient from './client';
import { 
  LoginResponse, 
  User, 
  ChangePasswordRequest,
  UpdateProfileRequest,
  RegisterRequest 
} from '@/types';
import { ApiResponse } from '@/types';

export const authApi = {
  // Login
  login: async (username: string, password: string): Promise<LoginResponse> => {
    console.log('🔵 [authApi] Login attempt for:', username);
    console.log('🔵 [authApi] Request payload:', { username, password: '***' });
    
    try {
      const response = await apiClient.post<LoginResponse>('/auth/login', {
        username,
        password,
      });
      
      console.log('🟢 [authApi] Login response status:', response.status);
      console.log('🟢 [authApi] Login response data:', response.data);
      
      return response.data;
    } catch (error) {
      console.error('🔴 [authApi] Login error:', error);
      throw error;
    }
  },

  // Get current user profile
  getProfile: async (): Promise<ApiResponse<User>> => {
    console.log('🔵 [authApi] Fetching profile');
    const response = await apiClient.get<ApiResponse<User>>('/profile');
    return response.data;
  },

  // Update profile
  updateProfile: async (data: UpdateProfileRequest): Promise<ApiResponse> => {
    console.log('🔵 [authApi] Updating profile:', data);
    const response = await apiClient.put<ApiResponse>('/profile', data);
    return response.data;
  },

  // Change password
  changePassword: async (data: ChangePasswordRequest): Promise<ApiResponse> => {
    console.log('🔵 [authApi] Changing password');
    const response = await apiClient.post<ApiResponse>('/profile/change-password', data);
    return response.data;
  },

  // Logout
  logout: async (): Promise<ApiResponse> => {
    console.log('🔵 [authApi] Logging out');
    const response = await apiClient.post<ApiResponse>('/auth/logout');
    return response.data;
  },
};