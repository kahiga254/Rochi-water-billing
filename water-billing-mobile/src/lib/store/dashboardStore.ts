import { create } from 'zustand';
import { Bill, MeterReading, Customer } from '@/types';
import { customerApi } from '../api/customer';

interface DashboardState {
  // Customer data
  customer: Customer | null;
  meterNumber: string | null;
  
  // Billing data
  currentBill: Bill | null;
  recentBills: Bill[];
  readingHistory: MeterReading[];
  
  // UI state
  isLoading: boolean;
  error: string | null;
  
  // Actions
  setMeterNumber: (meterNumber: string) => void;
  fetchCustomerData: (meterNumber: string) => Promise<void>;
  fetchBills: (meterNumber: string) => Promise<void>;
  fetchReadingHistory: (meterNumber: string) => Promise<void>;
  clearError: () => void;
  reset: () => void;
}

export const useDashboardStore = create<DashboardState>((set, get) => ({
  // Initial state
  customer: null,
  meterNumber: null,
  currentBill: null,
  recentBills: [],
  readingHistory: [],
  isLoading: false,
  error: null,

  // Actions
  setMeterNumber: (meterNumber) => set({ meterNumber }),

  fetchCustomerData: async (meterNumber) => {
    try {
      set({ isLoading: true, error: null });
      const response = await customerApi.getCustomer(meterNumber);
      if (response.success) {
        set({ customer: response.data });
      }
    } catch (error: any) {
      set({ 
        error: error.response?.data?.message || 'Failed to fetch customer data' 
      });
    } finally {
      set({ isLoading: false });
    }
  },

  fetchBills: async (meterNumber) => {
    try {
      set({ isLoading: true, error: null });
      
      // Fetch current bill
      const currentBillRes = await customerApi.getCurrentBill(meterNumber);
      if (currentBillRes.success && currentBillRes.data && currentBillRes.data.length > 0) {
        set({ currentBill: currentBillRes.data[0] });
      }

      // Fetch recent bills
      const billsRes = await customerApi.getBills(meterNumber, '', 12);
      if (billsRes.success && billsRes.data) {
        set({ recentBills: billsRes.data });
      }
    } catch (error: any) {
      set({ 
        error: error.response?.data?.message || 'Failed to fetch bills' 
      });
    } finally {
      set({ isLoading: false });
    }
  },

  fetchReadingHistory: async (meterNumber) => {
    try {
      set({ isLoading: true, error: null });
      const response = await customerApi.getReadingHistory(meterNumber, 24);
      if (response.success && response.data) {
        set({ readingHistory: response.data });
      }
    } catch (error: any) {
      set({ 
        error: error.response?.data?.message || 'Failed to fetch reading history' 
      });
    } finally {
      set({ isLoading: false });
    }
  },

  clearError: () => set({ error: null }),
  
  reset: () => set({
    customer: null,
    meterNumber: null,
    currentBill: null,
    recentBills: [],
    readingHistory: [],
    error: null,
    isLoading: false,
  }),
}));