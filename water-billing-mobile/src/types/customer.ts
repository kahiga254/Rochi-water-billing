export interface Address {
  street_address: string;
  city: string;
  state?: string;
  postal_code?: string;
  country: string;
  landmark?: string;
}

export interface Customer {
  id: string;
  meter_number: string;
  account_number?: string;
  first_name: string;
  last_name: string;
  phone_number: string;
  email?: string;
  address: Address;
  customer_type: string;
  connection_type: string;
  meter_type?: string;
  zone: string;
  subzone?: string;
  tariff_code?: string;
  rate_per_unit: number;
  fixed_charge: number;
  meter_installation_date?: string;
  connection_date: string;
  last_reading_date?: string;
  last_reading?: number;
  average_consumption?: number;
  balance: number;
  total_paid?: number;
  total_consumed?: number;
  status: string;
  created_at: string;
  updated_at: string;
}