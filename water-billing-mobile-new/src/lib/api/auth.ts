import apiClient from './client';
import { LoginCredentials, LoginResponse, User } from '@/types';

export const authApi = {
  login: async (credentials: LoginCredentials): Promise<LoginResponse> => {
    const response = await apiClient.post('/auth/login', credentials);
    return response.data;
  },

  logout: async (): Promise<{ success: boolean }> => {
    const response = await apiClient.post('/auth/logout');
    return response.data;
  },

  getProfile: async (): Promise<{ success: boolean; data: User }> => {
    const response = await apiClient.get('/profile');
    return response.data;
  },

  updateProfile: async (data: Partial<User>): Promise<{ success: boolean }> => {
    const response = await apiClient.put('/profile', data);
    return response.data;
  },

  changePassword: async (currentPassword: string, newPassword: string): Promise<{ success: boolean }> => {
    const response = await apiClient.post('/profile/change-password', {
      current_password: currentPassword,
      new_password: newPassword,
      confirm_password: newPassword
    });
    return response.data;
  },
};