import axios from 'axios';
import { useAuthStore } from '@/lib/store/authStore';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
  timeout: 30000,
  withCredentials: false, // Important: Set to false if your backend doesn't use cookies
});

// Request interceptor - ONLY add auth token, nothing else
apiClient.interceptors.request.use(
  (config) => {
    console.log('📤 [apiClient] Request:', {
      method: config.method,
      url: config.url,
      baseURL: config.baseURL,
      data: config.data,
      headers: config.headers
    });
    
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    console.error('📤 [apiClient] Request Error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor - ONLY log and handle 401
apiClient.interceptors.response.use(
  (response) => {
    console.log('📥 [apiClient] Response:', {
      status: response.status,
      statusText: response.statusText,
      url: response.config.url,
      data: response.data
    });
    return response;
  },
  (error) => {
    console.error('📥 [apiClient] Response Error:', {
      status: error.response?.status,
      statusText: error.response?.statusText,
      data: error.response?.data,
      config: error.config
    });
    
    if (error.response?.status === 401) {
      console.warn('⚠️ [apiClient] 401 Unauthorized - clearing auth state');
      useAuthStore.getState().logout();
    }
    return Promise.reject(error);
  }
);

export default apiClient;