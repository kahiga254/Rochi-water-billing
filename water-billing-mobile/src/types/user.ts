// User roles in the system
export type UserRole = 
  | 'admin' 
  | 'reader' 
  | 'cashier' 
  | 'manager' 
  | 'customer_service' 
  | 'customer';

// User model matching your Go backend
export interface User {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone_number: string;
  username: string;
  role: UserRole;
  department?: string;
  employee_id?: string;
  assigned_zone?: string;
  meter_number?: string;
  permissions?: string[];
  is_active: boolean;
  last_login?: string | null;
  created_at: string;
  updated_at: string;
}

// Login request payload
export interface LoginRequest {
  username: string;
  password: string;
}

// Login response - MATCH YOUR GO BACKEND
export interface LoginResponse {
  success: boolean;
  message: string;
  data: {
    token: string;
    user: User;
  };
}

// Register request payload (admin only)
export interface RegisterRequest {
  first_name: string;
  last_name: string;
  email: string;
  username: string;
  password: string;
  phone_number?: string;
  role?: UserRole;
  employee_id?: string;
  department?: string;
  zone?: string;
}

// Auth state for frontend store
export interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

// Password change request
export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

// Profile update request
export interface UpdateProfileRequest {
  first_name?: string;
  last_name?: string;
  email?: string;
  phone_number?: string;
}