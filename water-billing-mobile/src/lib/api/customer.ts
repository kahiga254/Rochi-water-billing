import apiClient from './client';

export const customerApi = {
  // Get all customers with pagination - ADD THIS METHOD
  getAll: async (page = 1, limit = 100) => {
    const response = await apiClient.get(`/customers?page=${page}&limit=${limit}`);
    return response.data;
  },

   create: async (customerData: any) => {
    const response = await apiClient.post('/customers', customerData);
    return response.data;
  },

  // Get customer by meter number
  getCustomerByMeter: async (meterNumber: string) => {
    const response = await apiClient.get(`/customers/meter/${meterNumber}`);
    return response.data;
  },

  // Get customer bills
  getBills: async (meterNumber: string) => {
    const response = await apiClient.get(`/billing/customers/${meterNumber}/bills?limit=12`);
    return response.data;
  },

  // Get reading history for charts
  getReadingHistory: async (meterNumber: string) => {
    const response = await apiClient.get(`/billing/customers/${meterNumber}/readings?limit=24`);
    return response.data;
  },

  // Get current bill (pending)
  getCurrentBill: async (meterNumber: string) => {
    const response = await apiClient.get(`/billing/customers/${meterNumber}/bills?status=pending&limit=1`);
    return response.data;
  },

  // Make payment (simulate M-Pesa)
  makePayment: async (billId: string, amount: number, phoneNumber: string) => {
    const response = await apiClient.post(`/billing/bills/${billId}/pay`, {
      amount,
      payment_method: 'mpesa',
      transaction_id: `MPESA${Date.now()}`,
      payer_phone: phoneNumber,
      notes: 'Payment via customer portal'
    });
    return response.data;
  },

  // Get payment history
  getPaymentHistory: async (meterNumber: string) => {
    const response = await apiClient.get(`/payments?meter_number=${meterNumber}&limit=10`);
    return response.data;
  }
};