"use client";

import { useState, useEffect } from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/lib/hooks/useAuth';
import { customerApi } from '@/lib/api/customer';
import { CreditCard, Download, Calendar, AlertCircle, RefreshCw } from 'lucide-react';
import { format } from 'date-fns';
import { toast } from 'sonner';

interface Payment {
  id: string;
  payment_date: string;
  amount: number;
  payment_method: string;
  transaction_id: string;
  receipt_number: string;
  status: string;
}

function CustomerPayments() {
  const { user } = useAuth();
  const [loading, setLoading] = useState(true);
  const [payments, setPayments] = useState<Payment[]>([]);
  const [error, setError] = useState<string | null>(null);

  const meterNumber = user?.meter_number;

  useEffect(() => {
    if (meterNumber) {
      fetchPayments();
    }
  }, [meterNumber]);

  const fetchPayments = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await customerApi.getPaymentHistory(meterNumber!);
      if (response.success) {
        setPayments(response.data || []);
      } else {
        setError(response.message || 'Failed to load payments');
      }
    } catch (error) {
      console.error('Error fetching payments:', error);
      setError('Failed to load payment history');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading payment history...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Payment History</h1>
          <p className="text-gray-600 mt-1">View all your water bill payments</p>
        </div>
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm font-medium text-red-800">Failed to load payments</p>
              <p className="text-sm text-red-600 mt-1">{error}</p>
              <button
                onClick={fetchPayments}
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
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Payment History</h1>
        <p className="text-gray-600 mt-1">View all your water bill payments</p>
      </div>

      {/* Payments List */}
      <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
        {payments.length === 0 ? (
          <div className="text-center py-12">
            <CreditCard className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-500">No payment history found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Receipt</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Transaction</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Method</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {payments.map((payment) => (
                  <tr key={payment.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <Calendar size={14} className="text-gray-400" />
                        {format(new Date(payment.payment_date), 'dd MMM yyyy')}
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm font-mono">{payment.receipt_number}</td>
                    <td className="px-6 py-4 text-sm font-mono">{payment.transaction_id}</td>
                    <td className="px-6 py-4 text-sm capitalize">{payment.payment_method}</td>
                    <td className="px-6 py-4 text-sm font-medium">KSh {payment.amount}</td>
                    <td className="px-6 py-4">
                      <span className={`px-2 py-1 rounded-full text-xs ${
                        payment.status === 'completed' 
                          ? 'bg-green-100 text-green-700' 
                          : 'bg-yellow-100 text-yellow-700'
                      }`}>
                        {payment.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <button className="text-blue-600 hover:text-blue-900">
                        <Download size={18} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

export default function CustomerPaymentsPage() {
  return (
    <ProtectedRoute allowedRoles={['customer_service']}>
      <CustomerPayments />
    </ProtectedRoute>
  );
}