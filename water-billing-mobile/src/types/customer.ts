// Address structure matching Go backend
export interface Address {
  street_address: string;
  city: string;
  state?: string;
  postal_code?: string;
  country: string;
  landmark?: string;
}

// Customer types
export type CustomerType = 'residential' | 'commercial' | 'industrial' | 'institutional';
export type ConnectionType = 'metered' | 'unmetered';
export type MeterType = 'digital' | 'analog' | 'smart';
export type CustomerStatus = 'active' | 'inactive' | 'disconnected' | 'pending' | 'suspended';

// Customer model matching Go backend
export interface Customer {
  id: string;
  meter_number: string;
  account_number?: string;
  first_name: string;
  last_name: string;
  full_name?: string; // Derived from first_name + last_name
  phone_number: string;
  email?: string;
  id_number?: string;
  address: Address;
  customer_type: CustomerType;
  connection_type: ConnectionType;
  meter_type?: MeterType;
  zone: string;
  subzone?: string;
  tariff_code?: string;
  rate_per_unit: number;
  fixed_charge: number;
  
  // Meter information
  meter_brand?: string;
  meter_size?: string;
  meter_installation_date?: string;
  meter_location?: string;
  
  // Reading information
  initial_reading?: number;
  connection_date: string;
  last_reading_date?: string;
  last_reading?: number;
  average_consumption?: number;
  
  // Financial information
  balance: number; // Positive = credit, Negative = arrears
  total_paid?: number;
  total_consumed?: number;
  
  // Status information
  status: CustomerStatus;
  disconnection_reason?: string;
  reconnection_date?: string;
  
  // Additional information
  emergency_contact?: string;
  emergency_phone?: string;
  property_owner?: string;
  property_type?: string;
  number_of_occupants?: number;
  notes?: string;
  
  // Timestamps
  created_at: string;
  updated_at: string;
}

// Customer statistics response
export interface CustomerStatistics {
  total: number;
  active: number;
  inactive: number;
  disconnected: number;
  customer_types: Record<CustomerType, number>;
  top_zones: Record<string, number>;
}

// Customer search params
export interface CustomerSearchParams {
  q?: string;
  zone?: string;
  status?: CustomerStatus;
  customer_type?: CustomerType;
  page?: number;
  limit?: number;
}

// Customer create/update request
export interface CustomerCreateRequest {
  meter_number: string;
  first_name: string;
  last_name: string;
  phone_number: string;
  email?: string;
  address: Address;
  zone: string;
  customer_type: CustomerType;
  connection_type: ConnectionType;
  meter_type?: MeterType;
  connection_date: string;
  status?: CustomerStatus;
  rate_per_unit?: number;
  fixed_charge?: number;
  initial_reading?: number;
  notes?: string;
}

// Bulk customer create request
export interface BulkCustomerCreateRequest {
  customers: CustomerCreateRequest[];
}