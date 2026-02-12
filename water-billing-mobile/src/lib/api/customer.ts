import apiClient from './client';
import { Customer, Bill, MeterReading } from '@/types';

export const customerApi = {
  // Get customer by meter number
  getCustomer: async (meterNumber: string) => {
    const response = await apiClient.get(`/customers/meter/${meterNumber}`);
    return response.data;
  },

  // Get customer by ID (from auth)
  getCustomerById: async (id: string) => {
    const response = await apiClient.get(`/customers/${id}`);
    return response.data;
  },

  // Get customer bills
  getBills: async (meterNumber: string, status?: string, limit: number = 12) => {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    if (limit) params.append('limit', limit.toString());
    
    const response = await apiClient.get(
      `/billing/customers/${meterNumber}/bills?${params.toString()}`
    );
    return response.data;
  },

  // Get current bill/unpaid bills
  getCurrentBill: async (meterNumber: string) => {
    const response = await apiClient.get(
      `/billing/customers/${meterNumber}/bills?status=pending&limit=1`
    );
    return response.data;
  },

  // Get meter reading history
  getReadingHistory: async (meterNumber: string, limit: number = 24) => {
    const response = await apiClient.get(
      `/billing/customers/${meterNumber}/readings?limit=${limit}`
    );
    return response.data;
  },

  // Get payment history
  getPaymentHistory: async (meterNumber: string, limit: number = 12) => {
    const response = await apiClient.get(`/payments?meter_number=${meterNumber}&limit=${limit}`);
    return response.data;
  },

  // Get consumption statistics
  getConsumptionStats: async (meterNumber: string, months: number = 12) => {
    const response = await apiClient.get(
      `/customers/meter/${meterNumber}/consumption?months=${months}`
    );
    return response.data;
  },

  // Update customer profile
  updateProfile: async (customerId: string, data: Partial<Customer>) => {
    const response = await apiClient.put(`/customers/${customerId}`, data);
    return response.data;
  },
};