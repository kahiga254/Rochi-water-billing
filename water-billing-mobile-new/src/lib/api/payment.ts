import apiClient from './client';

export interface PaymentData {
  bill_id: string;
  meter_number: string;
  customer_id: string;
  customer_name: string;
  amount: number;
  payment_method: 'cash' | 'bank_transfer' | 'cheque' | 'mpesa';
  transaction_id: string;
  payment_date: string;
  collected_by: string;
  notes?: string;
  status: 'completed' | 'pending' | 'failed';
}

export interface PaymentResponse {
  id: string;
  receipt_number: string;
  // ... other fields
}

export const paymentApi = {
  // Record a new payment
  recordPayment: async (paymentData: PaymentData): Promise<{ success: boolean; message?: string; data?: PaymentResponse }> => {
    try {
      const response = await apiClient.post('/payments', paymentData);
      return response.data;
    } catch (error: any) {
      console.error('Error recording payment:', error);
      throw error;
    }
  },

  // Get payment history
  getPayments: async (params?: { meter_number?: string; start_date?: string; end_date?: string }) => {
    try {
      const response = await apiClient.get('/payments', { params });
      return response.data;
    } catch (error) {
      console.error('Error fetching payments:', error);
      throw error;
    }
  },

  // Get payment by ID
  getPaymentById: async (id: string) => {
    try {
      const response = await apiClient.get(`/payments/${id}`);
      return response.data;
    } catch (error) {
      console.error('Error fetching payment:', error);
      throw error;
    }
  },

  // Generate receipt
  generateReceipt: async (paymentId: string) => {
    try {
      const response = await apiClient.get(`/payments/${paymentId}/receipt`, {
        responseType: 'blob'
      });
      return response.data;
    } catch (error) {
      console.error('Error generating receipt:', error);
      throw error;
    }
  }
};