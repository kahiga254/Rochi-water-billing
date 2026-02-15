import apiClient from './client';

export const billsApi = {
  // Get all bills with pagination
  getAll: async (page = 1, limit = 20, status?: string) => {
    let url = `/billing/bills/unpaid?page=${page}&limit=${limit}`;
    if (status && status !== 'all') {
      url += `&status=${status}`;
    }
    const response = await apiClient.get(url);
    return response.data;
  },

  // Get bill by ID
  getById: async (id: string) => {
    const response = await apiClient.get(`/billing/bills/${id}`);
    return response.data;
  },

  // Get bills by customer meter number
  getByMeter: async (meterNumber: string) => {
    const response = await apiClient.get(`/billing/customers/${meterNumber}/bills`);
    return response.data;
  },

  // Get overdue bills
  getOverdue: async () => {
    const response = await apiClient.get('/billing/bills/overdue');
    return response.data;
  },

  // Get billing summary
  getSummary: async () => {
    const now = new Date();
    const startDate = new Date(now.getFullYear(), now.getMonth(), 1).toISOString();
    const endDate = new Date(now.getFullYear(), now.getMonth() + 1, 0).toISOString();
    
    const response = await apiClient.get(`/billing/summary?start_date=${startDate}&end_date=${endDate}`);
    return response.data;
  },

  // Generate bill (admin only)
  generateBill: async (meterNumber: string, readingId: string) => {
    const response = await apiClient.post('/billing/generate', {
      meter_number: meterNumber,
      reading_id: readingId
    });
    return response.data;
  },

  // Send bill notification
  sendNotification: async (billId: string) => {
    const response = await apiClient.post(`/sms/bills/${billId}/notify`);
    return response.data;
  },

  // Export bills to CSV
  exportToCSV: async (filters?: any) => {
    const response = await apiClient.get('/billing/export', {
      params: filters,
      responseType: 'blob'
    });
    return response.data;
  }
};