"use client";

import { useState, useEffect } from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/lib/hooks/useAuth';
import { customerApi } from '@/lib/api/customer';
import { 
  Droplets, 
  FileText, 
  CreditCard, 
  AlertCircle,
  CheckCircle,
  Clock,
  TrendingUp,
  Calendar,
  Phone,
  RefreshCw
} from 'lucide-react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer
} from 'recharts';
import { toast } from 'sonner';
import { format } from 'date-fns';

interface Bill {
  id: string;
  bill_number: string;
  billing_period: string;
  total_amount: number;
  amount_paid: number;
  balance: number;
  status: string;
  due_date: string;
  consumption: number;
}

interface Reading {
  id: string;
  reading_date: string;
  consumption: number;
  billing_period: string;
}

function CustomerDashboard() {
  const { user } = useAuth();
  const [loading, setLoading] = useState(true);
  const [customer, setCustomer] = useState<any>(null);
  const [bills, setBills] = useState<Bill[]>([]);
  const [readings, setReadings] = useState<Reading[]>([]);
  const [currentBill, setCurrentBill] = useState<Bill | null>(null);
  const [paymentModal, setPaymentModal] = useState(false);
  const [phoneNumber, setPhoneNumber] = useState('');
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [apiResponse, setApiResponse] = useState<any>(null);

  const meterNumber = user?.meter_number || '';

  useEffect(() => {
    if (meterNumber) {
      fetchCustomerData();
    } else {
      console.warn('No meter number found for user');
      setLoading(false);
      setError('No meter number associated with this account');
    }
  }, [meterNumber]);

  const fetchCustomerData = async () => {
    try {
      setLoading(true);
      setError(null);
      console.log('🚀 Fetching data for meter:', meterNumber);
      
      // Fetch customer details
      const customerRes = await customerApi.getCustomerByMeter(meterNumber);
      console.log('📥 Customer API Response:', customerRes);
      
      if (customerRes && customerRes.success) {
        setCustomer(customerRes.data);
        console.log('✅ Customer data set:', customerRes.data);
      } else {
        console.warn('❌ Customer fetch failed:', customerRes?.message);
      }

      // Fetch bills
      const billsRes = await customerApi.getBills(meterNumber);
      console.log('📥 Bills API Response:', billsRes);
      
      if (billsRes && billsRes.success) {
        let billsData: Bill[] = [];
        
        if (Array.isArray(billsRes.data)) {
          billsData = billsRes.data;
        } else if (billsRes.data && Array.isArray(billsRes.data.bills)) {
          billsData = billsRes.data.bills;
        } else {
          billsData = [];
        }
        
        setBills(billsData);
        console.log('✅ Bills data set:', billsData);
        
        // Find current pending bill
        const pending = billsData.find((b: Bill) => b.status === 'pending' || b.status === 'overdue');
        setCurrentBill(pending || null);
        console.log('💰 Current pending bill:', pending);
      }

      // Fetch reading history
      const readingsRes = await customerApi.getReadingHistory(meterNumber);
      console.log('📥 Readings API Response:', readingsRes);
      
      if (readingsRes && readingsRes.success) {
        let readingsData: Reading[] = [];
        
        if (Array.isArray(readingsRes.data)) {
          readingsData = readingsRes.data;
        } else if (readingsRes.data && Array.isArray(readingsRes.data.readings)) {
          readingsData = readingsRes.data.readings;
        } else {
          readingsData = [];
        }
        
        setReadings(readingsData);
        console.log('✅ Readings data set:', readingsData);
      }

    } catch (error) {
      console.error('💥 Error in fetchCustomerData:', error);
      setError('Failed to load dashboard data. Please try again.');
      toast.error('Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    fetchCustomerData();
  };

  const handlePayment = async () => {
    if (!currentBill) return;
    if (!phoneNumber || phoneNumber.length < 10) {
      toast.error('Please enter a valid phone number');
      return;
    }

    try {
      setProcessing(true);
      const response = await customerApi.makePayment(
        currentBill.id,
        currentBill.balance,
        phoneNumber
      );

      if (response.success) {
        toast.success('Payment initiated! Please check your phone for M-Pesa prompt.');
        setPaymentModal(false);
        setPhoneNumber('');
        // Refresh data
        fetchCustomerData();
      } else {
        toast.error(response.message || 'Payment failed');
      }
    } catch (error) {
      toast.error('Payment failed');
    } finally {
      setProcessing(false);
    }
  };

  // Calculate totals
  const totalConsumption = readings.reduce((sum, r) => sum + r.consumption, 0);
  const avgConsumption = readings.length > 0 ? totalConsumption / readings.length : 0;
  
  // Calculate balance from bills
  const totalBalance = bills.reduce((sum, bill) => sum + bill.balance, 0);
  
  // Get latest customer balance or use calculated balance
  const displayBalance = customer?.balance !== undefined ? customer.balance : totalBalance;

  // Prepare chart data
  const chartData = readings.slice(0, 12).reverse().map(r => ({
    month: r.billing_period,
    consumption: r.consumption
  }));

  const getStatusBadge = (status: string) => {
    switch(status) {
      case 'paid':
        return <span className="flex items-center gap-1 px-2 py-1 bg-green-100 text-green-700 rounded-full text-xs"><CheckCircle size={12} /> Paid</span>;
      case 'pending':
        return <span className="flex items-center gap-1 px-2 py-1 bg-yellow-100 text-yellow-700 rounded-full text-xs"><Clock size={12} /> Pending</span>;
      case 'overdue':
        return <span className="flex items-center gap-1 px-2 py-1 bg-red-100 text-red-700 rounded-full text-xs"><AlertCircle size={12} /> Overdue</span>;
      default:
        return <span className="px-2 py-1 bg-gray-100 text-gray-700 rounded-full text-xs">{status}</span>;
    }
  };

  // Format balance with proper sign and color
  const formatBalance = (balance: number) => {
    if (balance > 0) {
      return {
        text: `KSh ${balance.toLocaleString()}`,
        className: "text-red-600",
        label: "Outstanding Balance"
      };
    } else if (balance < 0) {
      return {
        text: `KSh ${Math.abs(balance).toLocaleString()}`,
        className: "text-green-600",
        label: "Credit Balance"
      };
    } else {
      return {
        text: "KSh 0",
        className: "text-gray-900",
        label: "Current Balance"
      };
    }
  };

  const balanceDisplay = formatBalance(displayBalance);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading your dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              Welcome back, {user?.first_name}!
            </h1>
            <p className="text-gray-600 mt-1">
              Meter Number: <span className="font-mono font-medium">{meterNumber}</span>
            </p>
          </div>
          <button
            onClick={handleRefresh}
            className="p-2 bg-blue-100 text-blue-700 rounded-lg hover:bg-blue-200 transition-colors"
            title="Refresh data"
          >
            <RefreshCw size={20} />
          </button>
        </div>
        
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm font-medium text-red-800">Failed to load data</p>
              <p className="text-sm text-red-600 mt-1">{error}</p>
              <button
                onClick={handleRefresh}
                className="mt-4 flex items-center gap-2 px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors text-sm"
              >
                <RefreshCw size={16} />
                Retry
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Welcome Header with Refresh Button */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            Welcome back, {customer?.first_name || user?.first_name}!
          </h1>
          <p className="text-gray-600 mt-1">
            Meter Number: <span className="font-mono font-medium">{meterNumber}</span>
          </p>
        </div>
        <button
          onClick={handleRefresh}
          className="p-2 bg-blue-100 text-blue-700 rounded-lg hover:bg-blue-200 transition-colors"
          title="Refresh data"
        >
          <RefreshCw size={20} />
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Current Balance - UPDATED WITH PROPER FORMATTING */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">{balanceDisplay.label}</p>
              <p className={`text-2xl font-bold mt-2 ${balanceDisplay.className}`}>
                {balanceDisplay.text}
              </p>
              {displayBalance < 0 && (
                <p className="text-xs text-green-600 mt-1">
                  You have credit available for future bills
                </p>
              )}
              {displayBalance > 0 && (
                <p className="text-xs text-red-600 mt-1">
                  Please pay to avoid service interruption
                </p>
              )}
            </div>
            <div className="p-3 bg-blue-100 rounded-lg">
              <Droplets className="w-6 h-6 text-blue-600" />
            </div>
          </div>
        </div>

        {/* Current Bill */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Current Bill</p>
              <p className="text-2xl font-bold text-yellow-600 mt-2">
                KSh {currentBill?.balance?.toLocaleString() || 0}
              </p>
              {currentBill && getStatusBadge(currentBill.status)}
            </div>
            <div className="p-3 bg-yellow-100 rounded-lg">
              <FileText className="w-6 h-6 text-yellow-600" />
            </div>
          </div>
          {currentBill && currentBill.status !== 'paid' && (
            <button
              onClick={() => setPaymentModal(true)}
              className="mt-4 w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 text-sm font-medium"
            >
              Pay Now
            </button>
          )}
        </div>

        {/* Total Consumption */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Total Consumption</p>
              <p className="text-2xl font-bold text-green-600 mt-2">
                {totalConsumption.toFixed(1)} m³
              </p>
              <p className="text-xs text-gray-500 mt-1">
                Avg: {avgConsumption.toFixed(1)} m³/month
              </p>
            </div>
            <div className="p-3 bg-green-100 rounded-lg">
              <TrendingUp className="w-6 h-6 text-green-600" />
            </div>
          </div>
        </div>

        {/* Total Paid */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Total Paid</p>
              <p className="text-2xl font-bold text-purple-600 mt-2">
                KSh {customer?.total_paid?.toLocaleString() || 0}
              </p>
              <p className="text-xs text-gray-500 mt-1">
                Lifetime payments
              </p>
            </div>
            <div className="p-3 bg-purple-100 rounded-lg">
              <CreditCard className="w-6 h-6 text-purple-600" />
            </div>
          </div>
        </div>
      </div>

      {/* Consumption Chart */}
      <div className="bg-white p-6 rounded-lg shadow-sm border">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Water Consumption History</h2>
        <div className="h-64 w-full">
          {chartData.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="month" />
                <YAxis />
                <Tooltip />
                <Line 
                  type="monotone" 
                  dataKey="consumption" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  dot={{ fill: '#3b82f6' }}
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-full flex items-center justify-center text-gray-500">
              No consumption data available
            </div>
          )}
        </div>
      </div>

      {/* Recent Bills (Last 5) */}
      <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
        <div className="px-6 py-4 border-b flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Recent Bills</h2>
          <a href="/my-account/bills" className="text-sm text-blue-600 hover:text-blue-800">
            View All
          </a>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Period</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Bill No.</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Paid</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Balance</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {bills.slice(0, 5).map((bill) => (
                <tr key={bill.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm">{bill.billing_period}</td>
                  <td className="px-6 py-4 text-sm font-mono">{bill.bill_number}</td>
                  <td className="px-6 py-4 text-sm">KSh {bill.total_amount.toLocaleString()}</td>
                  <td className="px-6 py-4 text-sm">KSh {bill.amount_paid.toLocaleString()}</td>
                  <td className="px-6 py-4 text-sm font-medium">
                    {bill.balance > 0 ? (
                      <span className="text-red-600">KSh {bill.balance.toLocaleString()}</span>
                    ) : (
                      <span className="text-green-600">KSh 0</span>
                    )}
                  </td>
                  <td className="px-6 py-4">{getStatusBadge(bill.status)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Payment Modal */}
      {paymentModal && currentBill && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <h3 className="text-xl font-semibold text-gray-900 mb-4">Make Payment</h3>
            
            <div className="space-y-4">
              <div>
                <p className="text-sm text-gray-600">Bill Period</p>
                <p className="font-medium">{currentBill.billing_period}</p>
              </div>
              
              <div>
                <p className="text-sm text-gray-600">Amount Due</p>
                <p className="text-2xl font-bold text-blue-600">
                  KSh {currentBill.balance.toLocaleString()}
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  M-Pesa Phone Number
                </label>
                <div className="relative">
                  <Phone className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                  <input
                    type="tel"
                    value={phoneNumber}
                    onChange={(e) => setPhoneNumber(e.target.value)}
                    placeholder="e.g., 0712345678"
                    className="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <p className="text-xs text-gray-500 mt-1">
                  You'll receive an M-Pesa prompt on this number
                </p>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={() => setPaymentModal(false)}
                  className="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handlePayment}
                  disabled={processing}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {processing ? (
                    <>
                      <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                      Processing...
                    </>
                  ) : (
                    'Pay Now'
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default function CustomerDashboardPage() {
  return (
    <ProtectedRoute allowedRoles={['customer_service']}>
      <CustomerDashboard />
    </ProtectedRoute>
  );
}