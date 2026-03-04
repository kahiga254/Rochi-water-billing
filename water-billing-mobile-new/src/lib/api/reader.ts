import apiClient from './client';

export interface ReadingResponse {
  success: boolean;
  message: string;
  data: {
    readings: Reading[];
    total: number;
    page: number;
    limit: number;
  };
}

export interface Reading {
  id: string;
  meter_number: string;
  customer_name: string;
  reading_date: string;
  previous_reading: number;
  current_reading: number;
  consumption: number;
  notes?: string;
}

export const readerApi = {
  // Get customer by meter number
  getCustomerByMeter: async (meterNumber: string) => {
    try {
      const response = await apiClient.get(`/customers/meter/${meterNumber}`);
      return response.data;
    } catch (error) {
      console.error('Error fetching customer:', error);
      throw error;
    }
  },

  // Submit meter reading
  submitReading: async (data: {
    meter_number: string;
    current_reading: number;
    notes?: string;
  }) => {
    try {
      const response = await apiClient.post('/billing/readings', {
        ...data,
        reading_date: new Date().toISOString(),
        reading_type: 'manual',
        reading_method: 'mobile_app'
      });
      return response.data;
    } catch (error) {
      console.error('Error submitting reading:', error);
      throw error;
    }
  },

  // Get reader's reading history
  getMyReadings: async (page: number = 1, limit: number = 50) => {
    try {
      const response = await apiClient.get(`/billing/readings/my-readings?page=${page}&limit=${limit}`);
      return response.data;
    } catch (error: any) {
      console.error('Error fetching readings:', error);
      
      // If endpoint doesn't exist (404), throw a specific error
      if (error.response?.status === 404) {
        throw new Error('Reading history endpoint not available. Please check backend configuration.');
      }
      
      // If unauthorized
      if (error.response?.status === 401) {
        throw new Error('Unauthorized. Please login again.');
      }
      
      // Generic error
      throw new Error(error.response?.data?.message || 'Failed to fetch reading history');
    }
  },

  // Get reading by ID
  getReadingById: async (readingId: string) => {
    try {
      const response = await apiClient.get(`/billing/readings/${readingId}`);
      return response.data;
    } catch (error) {
      console.error('Error fetching reading:', error);
      throw error;
    }
  },

  // Send SMS notification
  sendNotification: async (billId: string) => {
    try {
      const response = await apiClient.post(`/sms/bills/${billId}/notify`);
      return response.data;
    } catch (error) {
      console.error('Error sending notification:', error);
      throw error;
    }
  }
};