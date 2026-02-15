import apiClient from './client';

export const readerApi = {
  getCustomerByMeter: async (meterNumber: string) => {
    const response = await apiClient.get(`/customers/meter/${meterNumber}`);
    return response.data;
  },

  submitReading: async (data: {
    meter_number: string;
    current_reading: number;
    notes?: string;
  }) => {
    const response = await apiClient.post('/billing/readings', {
      ...data,
      reading_date: new Date().toISOString(),
      reading_type: 'manual',
      reading_method: 'mobile_app'
    });
    return response.data;
  }
};