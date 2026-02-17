import apiClient from './client';
import { UserRole } from '@/types';

export interface CreateUserData {
  first_name: string;
  last_name: string;
  email: string;
  username: string;
  password: string;
  phone_number: string;
  role: UserRole;
  meter_number?: string;
  employee_id?: string;
  department?: string;
  assigned_zone?: string;
}

export const userApi = {
  // Create new user (admin only)
  createUser: async (userData: CreateUserData) => {
    const response = await apiClient.post('/users', userData);
    return response.data;
  },

  // Get all users (admin only)
  getUsers: async (role?: string) => {
    const url = role ? `/users?role=${role}` : '/users';
    const response = await apiClient.get(url);
    return response.data;
  },

  // Get user by ID
  getUserById: async (id: string) => {
    const response = await apiClient.get(`/users/${id}`);
    return response.data;
  },

  // Update user
  updateUser: async (id: string, updates: Partial<CreateUserData>) => {
    const response = await apiClient.put(`/users/${id}`, updates);
    return response.data;
  },

  // Delete user
  deleteUser: async (id: string) => {
    const response = await apiClient.delete(`/users/${id}`);
    return response.data;
  },

  // Toggle user active status
  toggleUserStatus: async (id: string, isActive: boolean) => {
    const response = await apiClient.patch(`/users/${id}/status`, {
      is_active: isActive
    });
    return response.data;
  }
};