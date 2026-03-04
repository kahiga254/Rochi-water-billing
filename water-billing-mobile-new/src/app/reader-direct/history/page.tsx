"use client";

import { useState, useEffect } from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/lib/hooks/useAuth';
import { readerApi } from '@/lib/api/reader';
import { Calendar, Droplets, Search, Eye, AlertCircle, RefreshCw } from 'lucide-react';
import { format } from 'date-fns';
import { toast } from 'sonner';

interface Reading {
  id: string;
  meter_number: string;
  customer_name: string;
  reading_date: string;
  previous_reading: number;
  current_reading: number;
  consumption: number;
  notes?: string;
}

function ReaderHistoryContent() {
  const { user } = useAuth();
  const [readings, setReadings] = useState<Reading[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [apiResponse, setApiResponse] = useState<any>(null);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    fetchMyReadings();
  }, [currentPage]);

  const fetchMyReadings = async () => {
    try {
      setLoading(true);
      setError(null);

    console.log('Current user:', user);
    console.log('User ID:', user?.id);

      
      const response = await readerApi.getMyReadings(currentPage, 50);
      console.log('🔥 RAW API RESPONSE:', JSON.stringify(response, null, 2));
      setApiResponse(response);
      
      if (response && response.success) {
        // Log the structure to debug
        console.log('Response data structure:', {
          hasData: !!response.data,
          dataType: response.data ? typeof response.data : 'undefined',
          isArray: response.data ? Array.isArray(response.data) : false,
          hasReadings: response.data?.readings ? true : false,
          readingsType: response.data?.readings ? typeof response.data.readings : 'undefined',
          isReadingsArray: response.data?.readings ? Array.isArray(response.data.readings) : false
        });
        
        // Extract readings array based on your API structure
        let readingsData: Reading[] = [];
        let total = 0;
        let pages = 1;
        
        // Case 1: response.data.readings is an array (your backend should return this)
        if (response.data?.readings && Array.isArray(response.data.readings)) {
          readingsData = response.data.readings;
          total = response.data.total || readingsData.length;
          pages = Math.ceil(total / 50);
          console.log('✅ Found readings in response.data.readings:', readingsData.length);
        }
        // Case 2: response.data is an array
        else if (response.data && Array.isArray(response.data)) {
          readingsData = response.data;
          total = readingsData.length;
          pages = Math.ceil(total / 50);
          console.log('✅ Found readings in response.data (array):', readingsData.length);
        }
        // Case 3: response itself is an array
        else if (Array.isArray(response)) {
          readingsData = response;
          total = readingsData.length;
          pages = Math.ceil(total / 50);
          console.log('✅ Found readings in response (array):', readingsData.length);
        }
        else {
          console.warn('❌ No readings array found in response');
          readingsData = [];
          total = 0;
          pages = 1;
        }
        
        setReadings(readingsData);
        setTotalCount(total);
        setTotalPages(pages);
        
        if (readingsData.length === 0) {
          console.log('No readings found for this reader');
          toast.info('No readings found');
        }
      } else {
        const errorMsg = response?.message || 'Failed to fetch reading history';
        setError(errorMsg);
        setReadings([]);
        toast.error(errorMsg);
      }
    } catch (error) {
      console.error('Failed to fetch readings:', error);
      const errorMsg = error instanceof Error ? error.message : 'Network error. Please check your connection.';
      setError(errorMsg);
      setReadings([]);
      toast.error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const handleRetry = () => {
    fetchMyReadings();
  };

  const handleNextPage = () => {
    if (currentPage < totalPages) {
      setCurrentPage(prev => prev + 1);
    }
  };

  const handlePrevPage = () => {
    if (currentPage > 1) {
      setCurrentPage(prev => prev - 1);
    }
  };

  // Safely filter readings (ensure readings is array)
  const filteredReadings = Array.isArray(readings) && readings.length > 0
    ? readings.filter(r => 
        r.meter_number?.toLowerCase().includes(search.toLowerCase()) ||
        r.customer_name?.toLowerCase().includes(search.toLowerCase())
      )
    : [];

  // Safe calculations
  const totalReadings = Array.isArray(readings) ? readings.length : 0;
  const totalConsumption = Array.isArray(readings) && readings.length > 0
    ? readings.reduce((sum, r) => sum + (r.consumption || 0), 0).toFixed(1)
    : '0';
  const uniqueCustomers = Array.isArray(readings) && readings.length > 0
    ? new Set(readings.map(r => r.meter_number).filter(Boolean)).size
    : 0;

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading your readings...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto space-y-6 p-4 md:p-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-gray-900">My Reading History</h1>
          <p className="text-gray-600 mt-1">View all meter readings you've submitted</p>
        </div>
        {error && (
          <button
            onClick={handleRetry}
            className="mt-4 sm:mt-0 flex items-center gap-2 px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors text-sm"
          >
            <RefreshCw size={16} />
            Retry
          </button>
        )}
      </div>

      {/* Error Message */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm font-medium text-red-800">Failed to load readings</p>
              <p className="text-sm text-red-600 mt-1">{error}</p>
              {apiResponse && (
                <details className="mt-2">
                  <summary className="text-xs text-red-600 cursor-pointer">View API Response</summary>
                  <pre className="mt-2 p-2 bg-red-100 rounded text-xs text-red-800 overflow-auto max-h-32">
                    {JSON.stringify(apiResponse, null, 2)}
                  </pre>
                </details>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Search Bar - Only show if we have data and no error */}
      {!error && totalReadings > 0 && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search by meter number or customer name..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
            />
          </div>
        </div>
      )}

      {/* Readings Table */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        {!error && totalReadings === 0 ? (
          <div className="text-center py-16 px-4">
            <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <Droplets className="w-8 h-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">No readings found</h3>
            <p className="text-gray-500 max-w-md mx-auto">
              You haven't submitted any meter readings yet. Use the meter reading form to record your first reading.
            </p>
            <button
              onClick={() => window.location.href = '/reader-direct/dashboard'}
              className="mt-6 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors inline-flex items-center gap-2"
            >
              Go to Meter Reading
            </button>
          </div>
        ) : !error && totalReadings > 0 && (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50 border-b border-gray-200">
                  <tr>
                    <th className="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
                    <th className="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Meter</th>
                    <th className="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Customer</th>
                    <th className="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Previous</th>
                    <th className="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Current</th>
                    <th className="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Consumption</th>
                    <th className="px-6 py-4 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {filteredReadings.length > 0 ? (
                    filteredReadings.map((reading) => (
                      <tr key={reading.id} className="hover:bg-gray-50 transition-colors">
                        <td className="px-6 py-4 text-sm text-gray-700 whitespace-nowrap">
                          <div className="flex items-center gap-2">
                            <Calendar size={14} className="text-gray-400" />
                            {format(new Date(reading.reading_date), 'dd MMM yyyy')}
                          </div>
                          <span className="text-xs text-gray-500 ml-6">
                            {format(new Date(reading.reading_date), 'HH:mm')}
                          </span>
                        </td>
                        <td className="px-6 py-4 font-mono text-sm text-gray-900 whitespace-nowrap">
                          {reading.meter_number}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-900">
                          {reading.customer_name}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-600 whitespace-nowrap">
                          {reading.previous_reading.toFixed(1)} m³
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-600 whitespace-nowrap">
                          {reading.current_reading.toFixed(1)} m³
                        </td>
                        <td className="px-6 py-4 text-sm font-medium text-blue-600 whitespace-nowrap">
                          <div className="flex items-center gap-1">
                            <Droplets size={14} />
                            {reading.consumption.toFixed(1)} m³
                          </div>
                        </td>
                        <td className="px-6 py-4 text-sm whitespace-nowrap">
                          <button 
                            className="text-blue-600 hover:text-blue-800 hover:bg-blue-50 p-2 rounded-lg transition-colors"
                            title="View details"
                          >
                            <Eye size={18} />
                          </button>
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={7} className="px-6 py-12 text-center text-gray-500">
                        <div className="flex flex-col items-center gap-2">
                          <Search size={32} className="text-gray-300" />
                          <p>No matching readings found</p>
                          <p className="text-sm">Try adjusting your search terms</p>
                        </div>
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
                <p className="text-sm text-gray-600">
                  Page {currentPage} of {totalPages} ({totalCount} total readings)
                </p>
                <div className="flex gap-2">
                  <button
                    onClick={handlePrevPage}
                    disabled={currentPage === 1}
                    className="px-3 py-1 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Previous
                  </button>
                  <button
                    onClick={handleNextPage}
                    disabled={currentPage === totalPages}
                    className="px-3 py-1 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* Summary Stats - Only show if we have data */}
      {!error && totalReadings > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-200">
            <p className="text-sm text-gray-500 mb-1">Total Readings</p>
            <p className="text-3xl font-bold text-gray-900">{totalReadings}</p>
          </div>
          <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-200">
            <p className="text-sm text-gray-500 mb-1">Total Consumption</p>
            <p className="text-3xl font-bold text-blue-600">{totalConsumption} m³</p>
          </div>
          <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-200">
            <p className="text-sm text-gray-500 mb-1">Unique Customers</p>
            <p className="text-3xl font-bold text-green-600">{uniqueCustomers}</p>
          </div>
        </div>
      )}
    </div>
  );
}

export default function ReaderHistoryPage() {
  return (
    <ProtectedRoute allowedRoles={['reader']}>
      <ReaderHistoryContent />
    </ProtectedRoute>
  );
}