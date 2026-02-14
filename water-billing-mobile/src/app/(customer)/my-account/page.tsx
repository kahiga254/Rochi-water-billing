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
  Phone
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

  // Get meter number from user (you'll need to add this to user object)
  const meterNumber = user?.meter_number || 'WMTR001'; // Fallback for demo

  useEffect(() => {
    fetchCustomerData();
  }, []);

  const fetchCustomerData = async () => {
    try {
      setLoading(true);
      
      // Fetch customer details
      const customerRes = await customerApi.getCustomerByMeter(meterNumber);
      if (customerRes.success) {
        setCustomer(customerRes.data);
      }

      // Fetch bills
      const billsRes = await customerApi.getBills(meterNumber);
      if (billsRes.success) {
        setBills(billsRes.data);
        
        // Find current pending bill
        const pending = billsRes.data.find((b: Bill) => b.status === 'pending');
        setCurrentBill(pending || null);
      }

      // Fetch reading history for chart
      const readingsRes = await customerApi.getReadingHistory(meterNumber);
      if (readingsRes.success) {
        setReadings(readingsRes.data);
      }

    } catch (error) {
      console.error('Failed to fetch customer data:', error);
      toast.error('Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
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
        // Refresh data
        fetchCustomerData();
      } else {
        toast.error('Payment failed');
      }
    } catch (error) {
      toast.error('Payment failed');
    } finally {
      setProcessing(false);
    }
  };

  // Prepare chart data
  const chartData = readings.slice(0, 12).reverse().map(r => ({
    month: r.billing_period,
    consumption: r.consumption
  }));

  const getStatusBadge = (status: string) => {
    switch(status) {
      case 'paid':
        return <span className="px-2 py-1 bg-green-100 text-green-700 rounded-full text-xs flex items-center gap-1"><CheckCircle size={12} /> Paid</span>;
      case 'pending':
        return <span className="px-2 py-1 bg-yellow-100 text-yellow-700 rounded-full text-xs flex items-center gap-1"><Clock size={12} /> Pending</span>;
      case 'overdue':
        return <span className="px-2 py-1 bg-red-100 text-red-700 rounded-full text-xs flex items-center gap-1"><AlertCircle size={12} /> Overdue</span>;
      default:
        return <span className="px-2 py-1 bg-gray-100 text-gray-700 rounded-full text-xs">{status}</span>;
    }
  };

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

  return (
    <div className="space-y-6">
      {/* Welcome Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">
          Welcome back, {customer?.first_name || user?.first_name}!
        </h1>
        <p className="text-gray-600 mt-1">
          Meter Number: <span className="font-mono font-medium">{meterNumber}</span>
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Current Balance */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Current Balance</p>
              <p className="text-3xl font-bold text-blue-600 mt-2">
                KSh {customer?.balance?.toLocaleString() || 0}
              </p>
              {currentBill && (
                <p className="text-xs text-gray-500 mt-1">
                  Due: {new Date(currentBill.due_date).toLocaleDateString()}
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
              <p className="text-3xl font-bold text-yellow-600 mt-2">
                KSh {currentBill?.balance?.toLocaleString() || 0}
              </p>
              {currentBill && getStatusBadge(currentBill.status)}
            </div>
            <div className="p-3 bg-yellow-100 rounded-lg">
              <FileText className="w-6 h-6 text-yellow-600" />
            </div>
          </div>
        </div>

        {/* Total Consumption */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Total Consumption</p>
              <p className="text-3xl font-bold text-green-600 mt-2">
                {customer?.total_consumed?.toFixed(1) || 0} m³
              </p>
              <p className="text-xs text-gray-500 mt-1">
                Avg: {((customer?.total_consumed || 0) / (bills.length || 1)).toFixed(1)} m³/month
              </p>
            </div>
            <div className="p-3 bg-green-100 rounded-lg">
              <Droplets className="w-6 h-6 text-green-600" />
            </div>
          </div>
        </div>
      </div>

      {/* Consumption Chart */}
      <div className="bg-white p-6 rounded-lg shadow-sm border">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Water Consumption History</h2>
        <div className="h-64">
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
        </div>
      </div>

      {/* Bills Table */}
      <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
        <div className="px-6 py-4 border-b">
          <h2 className="text-lg font-semibold text-gray-900">Bill History</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Period</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Bill No.</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Consumption</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Paid</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Balance</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {bills.map((bill) => (
                <tr key={bill.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm">{bill.billing_period}</td>
                  <td className="px-6 py-4 text-sm font-mono">{bill.bill_number}</td>
                  <td className="px-6 py-4 text-sm">{bill.consumption.toFixed(1)} m³</td>
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
                  <td className="px-6 py-4 text-right">
                    {bill.status === 'pending' && (
                      <button
                        onClick={() => {
                          setCurrentBill(bill);
                          setPaymentModal(true);
                        }}
                        className="text-blue-600 hover:text-blue-800 text-sm font-medium"
                      >
                        Pay Now
                      </button>
                    )}
                  </td>
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
    <ProtectedRoute allowedRoles={['customer']}>
      <CustomerDashboard />
    </ProtectedRoute>
  );
}