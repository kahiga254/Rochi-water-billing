// Meter reading model
export interface MeterReading {
  id: string;
  meter_number: string;
  customer_id: string;
  account_number?: string;
  customer_name: string;
  
  // Reading details
  reading_date: string;
  previous_reading: number;
  current_reading: number;
  consumption: number;
  
  // Charges
  rate_per_unit: number;
  water_charge: number;
  fixed_charge: number;
  arrears?: number;
  penalty?: number;
  discount?: number;
  total_amount: number;
  
  // Reading metadata
  reading_type: 'manual' | 'estimated' | 'actual' | 'self-read';
  reading_method: 'mobile_app' | 'field_agent' | 'customer_portal' | 'sms';
  reader_id?: string;
  reader_name: string;
  
  // Location & verification
  location?: GeoLocation;
  meter_photo_url?: string;
  is_verified: boolean;
  verified_by?: string;
  verification_date?: string;
  
  // Additional info
  meter_condition?: 'good' | 'damaged' | 'tampered';
  notes?: string;
  
  // Time period
  month: string;
  year: number;
  billing_period: string;
  season?: string;
  
  // Status
  status: 'recorded' | 'billed' | 'verified' | 'disputed';
  dispute_reason?: string;
  resolution?: string;
  
  // Timestamps
  created_at: string;
  updated_at: string;
}

// GeoLocation for meter reading
export interface GeoLocation {
  type: string;
  coordinates: [number, number]; // [longitude, latitude]
}

// Bill model
export interface Bill {
  id: string;
  meter_number: string;
  customer_id: string;
  reading_id: string;
  account_number?: string;
  customer_name: string;
  
  // Bill identification
  bill_number: string;
  bill_date: string;
  due_date: string;
  billing_period: string;
  
  // Reading information
  previous_reading: number;
  current_reading: number;
  consumption: number;
  
  // Charges breakdown
  rate_per_unit: number;
  water_charge: number;
  fixed_charge: number;
  arrears: number;
  penalty?: number;
  discount?: number;
  tax?: number;
  other_charges?: number;
  total_amount: number;
  
  // Payment information
  amount_paid: number;
  balance: number;
  status: 'pending' | 'paid' | 'overdue' | 'partially_paid' | 'cancelled';
  payment_date?: string;
  payment_method?: string;
  transaction_id?: string;
  receipt_number?: string;
  payment_notes?: string;
  
  // Notification status
  sms_sent: boolean;
  sms_sent_at?: string;
  email_sent: boolean;
  email_sent_at?: string;
  printed: boolean;
  printed_at?: string;
  
  // Timestamps
  created_at: string;
  updated_at: string;
}

// Payment model
export interface Payment {
  id: string;
  bill_id: string;
  meter_number: string;
  customer_id: string;
  customer_name: string;
  payment_date: string;
  amount: number;
  payment_method: 'cash' | 'mpesa' | 'bank_transfer' | 'cheque' | 'credit_card';
  transaction_id: string;
  receipt_number: string;
  payer_name?: string;
  payer_phone?: string;
  collected_by: string;
  status: 'completed' | 'pending' | 'failed' | 'refunded';
  notes?: string;
  created_at: string;
}

// Meter reading request
export interface MeterReadingRequest {
  meter_number: string;
  current_reading: number;
  reading_date: string;
  reading_type?: 'manual' | 'estimated' | 'actual' | 'self-read';
  reading_method?: 'mobile_app' | 'field_agent' | 'customer_portal' | 'sms';
  reader_id?: string;
  reader_name?: string;
  location?: GeoLocation;
  meter_photo_url?: string;
  meter_condition?: 'good' | 'damaged' | 'tampered';
  notes?: string;
}

// Bulk meter reading request
export interface BulkMeterReadingRequest {
  readings: MeterReadingRequest[];
}

// Payment request
export interface PaymentRequest {
  amount: number;
  payment_method: 'cash' | 'mpesa' | 'bank_transfer' | 'cheque' | 'credit_card';
  transaction_id?: string;
  receipt_number?: string;
  payer_name?: string;
  payer_phone?: string;
  notes?: string;
}

// Billing summary
export interface BillingSummary {
  period_start: string;
  period_end: string;
  status_breakdown: Record<string, StatusSummary>;
}

export interface StatusSummary {
  count: number;
  total_amount: number;
  total_paid: number;
}

// Dashboard statistics
export interface DashboardStats {
  total_customers: number;
  active_customers: number;
  total_bills: number;
  paid_bills: number;
  pending_bills: number;
  overdue_bills: number;
  total_revenue: number;
  collected_revenue: number;
  pending_revenue: number;
  collection_rate: number;
  average_consumption: number;
}

// Zone performance
export interface ZonePerformance {
  zone: string;
  customer_count: number;
  total_consumption: number;
  total_billed: number;
  total_collected: number;
  collection_rate: number;
  overdue_amount: number;
}