import apiClient from './client';

export const dashboardApi = {
  // Get customer statistics
  getCustomerStats: async () => {
    const response = await apiClient.get('/customers/statistics');
    return response.data;
  },

  // Get unpaid bills count
  getUnpaidBills: async () => {
    const response = await apiClient.get('/billing/bills/unpaid?limit=100');
    return response.data;
  },

  // Get billing summary for current month
  getBillingSummary: async () => {
    const now = new Date();
    const startDate = new Date(now.getFullYear(), now.getMonth(), 1).toISOString();
    const endDate = new Date(now.getFullYear(), now.getMonth() + 1, 0).toISOString();
    
    const response = await apiClient.get(`/billing/summary?start_date=${startDate}&end_date=${endDate}`);
    return response.data;
  },

  // Get overdue bills
  getOverdueBills: async () => {
    const response = await apiClient.get('/billing/bills/overdue');
    return response.data;
  },
};