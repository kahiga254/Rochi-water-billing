export type UserRole = 'admin' | 'reader' | 'cashier' | 'manager' | 'customer_service';

export interface User {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone_number: string;
  username: string;
  role: UserRole;
  meter_number?: string;
  employee_id?: string;
  department?: string;
  assigned_zone?: string;
  is_active: boolean;
  last_login?: string;
  created_at: string;
  updated_at: string;
}

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

export interface LoginCredentials {
  username: string;
  password: string;
}

export interface LoginResponse {
  success: boolean;
  message: string;
  data: {
    token: string;
    user: User;
  };
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}