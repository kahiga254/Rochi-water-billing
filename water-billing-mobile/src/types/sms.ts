// SMS log model
export interface SMSLog {
  id: string;
  customer_id: string;
  bill_id?: string;
  meter_number: string;
  phone_number: string;
  customer_name?: string;
  message_type: 'bill_notification' | 'payment_confirmation' | 'reminder' | 'disconnection_warning';
  message: string;
  status: 'sent' | 'failed' | 'delivered' | 'pending';
  provider?: 'twilio' | 'africas_talking' | 'nexmo';
  message_id?: string;
  cost?: number;
  error?: string;
  sent_at: string;
}

// SMS send request
export interface SMSRequest {
  phone: string;
  bill_id?: string;
  message?: string;
}

// Bulk SMS request
export interface BulkSMSRequest {
  bill_ids?: string[];
  meter_numbers?: string[];
  message?: string;
}

// SMS logs filter
export interface SMSLogsFilter {
  meter_number?: string;
  customer_id?: string;
  message_type?: string;
  status?: string;
  start_date?: string;
  end_date?: string;
  limit?: number;
}