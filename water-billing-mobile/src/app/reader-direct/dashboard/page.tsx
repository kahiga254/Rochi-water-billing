"use client";

import { useState, useEffect } from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/lib/hooks/useAuth';
import { readerApi } from '@/lib/api/reader';
import { 
  Search, 
  User, 
  MapPin, 
  Calendar, 
  Droplets,
  CheckCircle,
  AlertCircle,
  Phone,
  Mail,
  Home,
  X
} from 'lucide-react';
import { toast } from 'sonner';
import { format } from 'date-fns';

interface Customer {
  id: string;
  meter_number: string;
  first_name: string;
  last_name: string;
  phone_number: string;
  email?: string;
  address: {
    street_address: string;
    city: string;
  };
  zone: string;
  last_reading?: number;
  last_reading_date?: string;
}

function ReaderDashboardContent() {
  const { user } = useAuth();
  console.log('👤 Reader user:', user); // Debug log

  const [meterNumber, setMeterNumber] = useState('');
  const [customer, setCustomer] = useState<Customer | null>(null);
  const [currentReading, setCurrentReading] = useState('');
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [searched, setSearched] = useState(false);

  const searchCustomer = async () => {
    console.log('🔍 Searching for meter:', meterNumber); // Debug log
    
    if (!meterNumber.trim()) {
      toast.error('Please enter a meter number');
      return;
    }

    try {
      setLoading(true);
      setSearched(true);
      console.log('📡 Calling API for meter:', meterNumber);
      
      const response = await readerApi.getCustomerByMeter(meterNumber);
      console.log('📥 API Response:', response); // Debug log
      
      if (response.success) {
        console.log('✅ Customer found:', response.data);
        setCustomer(response.data);
        toast.success('Customer found');
      } else {
        console.log('❌ Customer not found');
        setCustomer(null);
        toast.error('Customer not found');
      }
    } catch (error) {
      console.error('💥 Search error:', error);
      toast.error('Failed to search customer');
      setCustomer(null);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmitReading = async () => {
    console.log('📝 Submitting reading:', currentReading); // Debug log
    
    if (!customer) return;
    
    const reading = parseFloat(currentReading);
    if (isNaN(reading) || reading <= 0) {
      toast.error('Please enter a valid reading');
      return;
    }

    if (customer.last_reading && reading <= customer.last_reading) {
      toast.error('Current reading must be greater than last reading');
      return;
    }

    try {
      setSubmitting(true);
      console.log('📡 Submitting to API:', {
        meter_number: customer.meter_number,
        current_reading: reading
      });
      
      const response = await readerApi.submitReading({
        meter_number: customer.meter_number,
        current_reading: reading,
        notes: `Reading taken by ${user?.first_name} ${user?.last_name}`
      });

      console.log('📥 Submit response:', response); // Debug log

      if (response.success) {
        toast.success('Reading submitted successfully! SMS sent to customer.');
        setCurrentReading('');
        setCustomer(null);
        setMeterNumber('');
        setSearched(false);
      } else {
        toast.error(response.message || 'Failed to submit reading');
      }
    } catch (error) {
      console.error('💥 Submission error:', error);
      toast.error('Failed to submit reading');
    } finally {
      setSubmitting(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      searchCustomer();
    }
  };

  // Debug render
  console.log('🔄 Render state:', { loading, submitting, searched, hasCustomer: !!customer });

  return (
    <div className="max-w-4xl mx-auto space-y-6 p-4">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Meter Reading</h1>
        <p className="text-gray-600 mt-1">Enter meter number to record reading</p>
      </div>

      {/* Search Section */}
      <div className="bg-white rounded-lg shadow-sm border p-6">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Meter Number
        </label>
        <div className="flex gap-3">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              value={meterNumber}
              onChange={(e) => {
                console.log('📝 Meter input:', e.target.value); // Debug log
                setMeterNumber(e.target.value.toUpperCase());
              }}
              onKeyPress={handleKeyPress}
              placeholder="Enter meter number (e.g., WMTR001)"
              className="w-full pl-10 pr-4 py-3 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              disabled={loading || !!customer}
            />
          </div>
          <button
            onClick={searchCustomer}
            disabled={loading || !meterNumber || !!customer}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {loading ? (
              <>
                <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                Searching...
              </>
            ) : (
              <>
                <Search size={18} />
                Search
              </>
            )}
          </button>
        </div>
      </div>

      {/* Customer Details & Reading Form */}
      {searched && (
        <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
          {customer ? (
            <>
              {/* Customer Details */}
              <div className="p-6 border-b bg-gray-50">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                    <User size={20} className="text-blue-600" />
                    Customer Details
                  </h2>
                  <button
                    onClick={() => {
                      setCustomer(null);
                      setMeterNumber('');
                      setSearched(false);
                    }}
                    className="p-1 hover:bg-gray-200 rounded-lg"
                  >
                    <X size={18} className="text-gray-500" />
                  </button>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="flex items-start gap-3">
                    <User size={18} className="text-gray-400 mt-0.5" />
                    <div>
                      <p className="text-sm text-gray-500">Name</p>
                      <p className="font-medium">{customer.first_name} {customer.last_name}</p>
                    </div>
                  </div>
                  
                  <div className="flex items-start gap-3">
                    <MapPin size={18} className="text-gray-400 mt-0.5" />
                    <div>
                      <p className="text-sm text-gray-500">Zone</p>
                      <p className="font-medium">{customer.zone}</p>
                    </div>
                  </div>
                  
                  <div className="flex items-start gap-3">
                    <Home size={18} className="text-gray-400 mt-0.5" />
                    <div>
                      <p className="text-sm text-gray-500">Address</p>
                      <p className="font-medium">{customer.address?.street_address}, {customer.address?.city}</p>
                    </div>
                  </div>
                  
                  <div className="flex items-start gap-3">
                    <Phone size={18} className="text-gray-400 mt-0.5" />
                    <div>
                      <p className="text-sm text-gray-500">Phone</p>
                      <p className="font-medium">{customer.phone_number}</p>
                    </div>
                  </div>
                  
                  {customer.email && (
                    <div className="flex items-start gap-3">
                      <Mail size={18} className="text-gray-400 mt-0.5" />
                      <div>
                        <p className="text-sm text-gray-500">Email</p>
                        <p className="font-medium">{customer.email}</p>
                      </div>
                    </div>
                  )}
                  
                  <div className="flex items-start gap-3">
                    <Calendar size={18} className="text-gray-400 mt-0.5" />
                    <div>
                      <p className="text-sm text-gray-500">Last Reading</p>
                      <p className="font-medium">
                        {customer.last_reading ? (
                          <>
                            {customer.last_reading} m³
                            {customer.last_reading_date && (
                              <span className="text-xs text-gray-500 ml-2">
                                ({format(new Date(customer.last_reading_date), 'dd MMM yyyy')})
                              </span>
                            )}
                          </>
                        ) : (
                          'No previous reading'
                        )}
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Reading Form */}
              <div className="p-6">
                <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                  <Droplets size={20} className="text-blue-600" />
                  Enter Current Reading
                </h2>
                
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Current Reading (m³)
                    </label>
                    <input
                      type="number"
                      step="0.1"
                      value={currentReading}
                      onChange={(e) => setCurrentReading(e.target.value)}
                      placeholder="e.g., 1250.5"
                      className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-lg"
                      autoFocus
                    />
                    {customer.last_reading && (
                      <p className="mt-2 text-sm text-gray-500">
                        Previous reading: <span className="font-medium">{customer.last_reading} m³</span>
                      </p>
                    )}
                  </div>

                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div className="flex items-start gap-3">
                      <AlertCircle size={20} className="text-blue-600 mt-0.5" />
                      <div>
                        <p className="text-sm font-medium text-blue-800">SMS Notification</p>
                        <p className="text-sm text-blue-600">
                          An SMS will be sent to {customer.phone_number} with the new bill details after submission.
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="flex gap-3 pt-4">
                    <button
                      onClick={() => {
                        setCustomer(null);
                        setMeterNumber('');
                        setCurrentReading('');
                        setSearched(false);
                      }}
                      className="px-6 py-3 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={handleSubmitReading}
                      disabled={submitting || !currentReading}
                      className="flex-1 px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                    >
                      {submitting ? (
                        <>
                          <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                          Submitting...
                        </>
                      ) : (
                        <>
                          <CheckCircle size={18} />
                          Submit Reading & Send SMS
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            </>
          ) : (
            <div className="p-12 text-center">
              <AlertCircle size={48} className="mx-auto mb-4 text-gray-400" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">Customer Not Found</h3>
              <p className="text-gray-500">
                No customer found with meter number "{meterNumber}"
              </p>
              <button
                onClick={() => {
                  setSearched(false);
                  setMeterNumber('');
                }}
                className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                Try Again
              </button>
            </div>
          )}
        </div>
      )}

      {/* Quick Tips */}
      <div className="bg-white rounded-lg shadow-sm border p-4">
        <div className="flex items-start gap-3">
          <div className="p-2 bg-blue-100 rounded-lg">
            <Droplets size={16} className="text-blue-600" />
          </div>
          <div>
            <p className="text-sm font-medium text-gray-900">Quick Tips</p>
            <ul className="mt-1 text-sm text-gray-500 space-y-1">
              <li>• Enter meter number and press Enter to search</li>
              <li>• Current reading must be higher than last reading</li>
              <li>• SMS will be sent automatically to customer</li>
              <li>• Customer will receive bill details with payment info</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function ReaderDashboardPage() {
  return (
    <ProtectedRoute allowedRoles={['reader']}>
      <ReaderDashboardContent />
    </ProtectedRoute>
  );
}