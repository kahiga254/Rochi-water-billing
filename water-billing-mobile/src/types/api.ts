import { User } from './user';

// Standard API response wrapper
export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
}

// Login response data
export interface LoginData {
  token: string;
  refresh_token?: string;
  user: User;
}

// Token refresh response
export interface TokenRefreshData {
  token: string;
}

// Paginated response
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

// Error response
export interface ApiError {
  success: false;
  message: string;
  error: string;
  status?: number;
}

// Query parameters for list endpoints
export interface ListQueryParams {
  page?: number;
  limit?: number;
  sort?: string;
  order?: 'asc' | 'desc';
  search?: string;
}

// Date range filter
export interface DateRangeFilter {
  start_date: string;
  end_date: string;
}