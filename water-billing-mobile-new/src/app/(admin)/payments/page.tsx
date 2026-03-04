"use client";

import { useState } from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/lib/hooks/useAuth';
import { paymentApi } from '@/lib/api/payment';
import { customerApi } from '@/lib/api/customer';
import { 
  CreditCard, 
  Search, 
  User, 
  Calendar,
  Phone,
  FileText,
  CheckCircle,
  AlertCircle,
  Download,
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
  balance: number;
}

interface Bill {
  id: string;
  bill_number: string;
  billing_period: string;
  total_amount: number;
  amount_paid: number;
  balance: number;
  status: string;
  due_date: string;
}

function AdminPaymentsPage() {
  const { user } = useAuth();
  const [searchMeter, setSearchMeter] = useState('');
  const [customer, setCustomer] = useState<Customer | null>(null);
  const [bills, setBills] = useState<Bill[]>([]);
  const [selectedBill, setSelectedBill] = useState<Bill | null>(null);
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [searchLoading, setSearchLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  
  // Payment form state
  const [paymentAmount, setPaymentAmount] = useState<number>(0);
  const [paymentMethod, setPaymentMethod] = useState<'cash' | 'bank_transfer' | 'cheque' | 'mpesa'>('cash');
  const [transactionId, setTransactionId] = useState('');
  const [paymentNotes, setPaymentNotes] = useState('');
  const [paymentDate, setPaymentDate] = useState(new Date().toISOString().split('T')[0]);

  const searchCustomer = async () => {
    if (!searchMeter.trim()) {
      toast.error('Please enter a meter number');
      return;
    }

    try {
      setSearchLoading(true);
      setCustomer(null);
      setBills([]);
      
      // Get customer details
      const customerRes = await customerApi.getCustomerByMeter(searchMeter.toUpperCase());
      if (customerRes.success) {
        setCustomer(customerRes.data);
        
        // Get unpaid bills for this customer
        const billsRes = await customerApi.getBills(searchMeter.toUpperCase());
        if (billsRes.success) {
          const unpaidBills = billsRes.data.filter((b: Bill) => 
            b.status === 'pending' || b.status === 'overdue'
          );
          setBills(unpaidBills);
        }
        
        toast.success('Customer found');
      } else {
        toast.error('Customer not found');
      }
    } catch (error) {
      console.error('Search error:', error);
      toast.error('Failed to search customer');
    } finally {
      setSearchLoading(false);
    }
  };

  const handleSelectBill = (bill: Bill) => {
    setSelectedBill(bill);
    setPaymentAmount(bill.balance);
    setShowPaymentModal(true);
  };

  const closeModal = () => {
    setShowPaymentModal(false);
    setSelectedBill(null);
    setPaymentAmount(0);
    setTransactionId('');
    setPaymentNotes('');
  };

  const handleRecordPayment = async () => {
    if (!selectedBill || !customer) return;
    
    if (paymentAmount <= 0) {
      toast.error('Payment amount must be greater than 0');
      return;
    }

    if (paymentAmount > selectedBill.balance) {
      toast.error('Payment amount cannot exceed bill balance');
      return;
    }

    try {
      setSubmitting(true);
      
      const paymentData = {
        bill_id: selectedBill.id,
        meter_number: customer.meter_number,
        customer_id: customer.id,
        customer_name: `${customer.first_name} ${customer.last_name}`,
        amount: paymentAmount,
        payment_method: paymentMethod,
        transaction_id: transactionId || `${paymentMethod.toUpperCase()}-${Date.now()}`,
        payment_date: new Date(paymentDate).toISOString(),
        collected_by: user?.username || 'admin',
        notes: paymentNotes,
        status: 'completed' as const
      };

      const response = await paymentApi.recordPayment(paymentData);
      
      if (response.success) {
        toast.success('Payment recorded successfully');
        closeModal();
        
        // Refresh bills
        const billsRes = await customerApi.getBills(customer.meter_number);
        if (billsRes.success) {
          const unpaidBills = billsRes.data.filter((b: Bill) => 
            b.status === 'pending' || b.status === 'overdue'
          );
          setBills(unpaidBills);
        }
        
        // Refresh customer balance
        const customerRes = await customerApi.getCustomerByMeter(customer.meter_number);
        if (customerRes.success) {
          setCustomer(customerRes.data);
        }
      } else {
        toast.error(response.message || 'Failed to record payment');
      }
    } catch (error) {
      console.error('Payment error:', error);
      toast.error('Failed to record payment');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="space-y-6 p-4 md:p-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl md:text-3xl font-bold text-gray-900">Record Payment</h1>
        <p className="text-gray-600 mt-1">Process cash, bank transfers, and cheque payments</p>
      </div>

      {/* Search Customer */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Search Customer by Meter Number
        </label>
        <div className="flex flex-col sm:flex-row gap-3">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              value={searchMeter}
              onChange={(e) => setSearchMeter(e.target.value.toUpperCase())}
              onKeyPress={(e) => e.key === 'Enter' && searchCustomer()}
              placeholder="Enter meter number (e.g., WMTR001)"
              className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              disabled={searchLoading}
            />
          </div>
          <button
            onClick={searchCustomer}
            disabled={searchLoading || !searchMeter}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center justify-center gap-2 transition-colors"
          >
            {searchLoading ? (
              <>
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
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

      {/* Customer Details & Bills */}
      {customer && (
        <div className="space-y-6">
          {/* Customer Info Card */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
              <User size={20} className="text-blue-600" />
              Customer Information
            </h2>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="bg-gray-50 p-3 rounded-lg">
                <p className="text-xs text-gray-500 mb-1">Name</p>
                <p className="font-medium text-gray-900">{customer.first_name} {customer.last_name}</p>
              </div>
              <div className="bg-gray-50 p-3 rounded-lg">
                <p className="text-xs text-gray-500 mb-1">Meter Number</p>
                <p className="font-mono font-medium text-gray-900">{customer.meter_number}</p>
              </div>
              <div className="bg-gray-50 p-3 rounded-lg">
                <p className="text-xs text-gray-500 mb-1">Current Balance</p>
                <p className="font-medium text-blue-600">KSh {customer.balance?.toLocaleString() || 0}</p>
              </div>
              <div className="bg-gray-50 p-3 rounded-lg">
                <p className="text-xs text-gray-500 mb-1">Phone</p>
                <p className="font-medium text-gray-900">{customer.phone_number}</p>
              </div>
              {customer.email && (
                <div className="bg-gray-50 p-3 rounded-lg">
                  <p className="text-xs text-gray-500 mb-1">Email</p>
                  <p className="font-medium text-gray-900">{customer.email}</p>
                </div>
              )}
            </div>
          </div>

          {/* Unpaid Bills */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 bg-gray-50">
              <h2 className="text-lg font-semibold text-gray-900">Unpaid Bills</h2>
            </div>
            
            {bills.length === 0 ? (
              <div className="text-center py-12">
                <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <p className="text-gray-500">No unpaid bills found</p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Bill #</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Period</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Due Date</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Balance</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Action</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {bills.map((bill) => (
                      <tr key={bill.id} className="hover:bg-gray-50 transition-colors">
                        <td className="px-6 py-4 font-mono text-sm text-gray-900">{bill.bill_number}</td>
                        <td className="px-6 py-4 text-sm text-gray-700">{bill.billing_period}</td>
                        <td className="px-6 py-4 text-sm text-gray-700">
                          {format(new Date(bill.due_date), 'dd MMM yyyy')}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-900">KSh {bill.total_amount.toLocaleString()}</td>
                        <td className="px-6 py-4 text-sm font-medium text-red-600">
                          KSh {bill.balance.toLocaleString()}
                        </td>
                        <td className="px-6 py-4">
                          <span className={`px-2 py-1 text-xs rounded-full ${
                            bill.status === 'overdue' 
                              ? 'bg-red-100 text-red-700' 
                              : 'bg-yellow-100 text-yellow-700'
                          }`}>
                            {bill.status}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-right">
                          <button
                            onClick={() => handleSelectBill(bill)}
                            className="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700 transition-colors"
                          >
                            Record Payment
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
      )}

      {/* Payment Modal */}
      {showPaymentModal && selectedBill && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl max-w-md w-full max-h-[90vh] overflow-y-auto shadow-xl">
            {/* Modal Header with visible X button */}
            <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
              <h3 className="text-xl font-semibold text-gray-900">Record Payment</h3>
              <button
                onClick={closeModal}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                aria-label="Close modal"
              >
                <X size={20} className="text-gray-600" />
              </button>
            </div>

            <div className="p-6 space-y-4">
              {/* Bill Summary */}
              <div className="bg-blue-50 p-4 rounded-lg space-y-2 border border-blue-200">
                <h4 className="font-medium text-blue-900 mb-2">Bill Details</h4>
                <div className="flex justify-between text-sm">
                  <span className="text-blue-700">Bill Number:</span>
                  <span className="font-mono font-medium text-blue-900">{selectedBill.bill_number}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-blue-700">Period:</span>
                  <span className="font-medium text-blue-900">{selectedBill.billing_period}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-blue-700">Total Amount:</span>
                  <span className="font-medium text-blue-900">KSh {selectedBill.total_amount.toLocaleString()}</span>
                </div>
                <div className="flex justify-between text-sm pt-2 border-t border-blue-200">
                  <span className="font-medium text-blue-900">Balance Due:</span>
                  <span className="font-bold text-blue-600">KSh {selectedBill.balance.toLocaleString()}</span>
                </div>
              </div>

              {/* Payment Form */}
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Payment Amount (KSh) <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="number"
                    value={paymentAmount}
                    onChange={(e) => setPaymentAmount(parseFloat(e.target.value))}
                    max={selectedBill.balance}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    placeholder="Enter amount"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Max amount: KSh {selectedBill.balance.toLocaleString()}
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Payment Method <span className="text-red-500">*</span>
                  </label>
                  <select
                    value={paymentMethod}
                    onChange={(e) => setPaymentMethod(e.target.value as any)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    <option value="cash">Cash</option>
                    <option value="bank_transfer">Bank Transfer</option>
                    <option value="cheque">Cheque</option>
                    <option value="mpesa">M-Pesa</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Transaction ID / Reference
                  </label>
                  <input
                    type="text"
                    value={transactionId}
                    onChange={(e) => setTransactionId(e.target.value)}
                    placeholder="e.g., TRX123456 or CHEQUE001"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Payment Date <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="date"
                    value={paymentDate}
                    onChange={(e) => setPaymentDate(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Notes (Optional)
                  </label>
                  <textarea
                    value={paymentNotes}
                    onChange={(e) => setPaymentNotes(e.target.value)}
                    rows={2}
                    placeholder="Additional notes about this payment"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </div>

              {/* Action Buttons - Clearly visible at the bottom */}
              <div className="sticky bottom-0 bg-white pt-4 border-t border-gray-200 mt-4 flex gap-3">
                <button
                  onClick={closeModal}
                  className="flex-1 px-4 py-3 border border-gray-300 rounded-lg text-gray-700 font-medium hover:bg-gray-50 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleRecordPayment}
                  disabled={submitting}
                  className="flex-1 px-4 py-3 bg-green-600 text-white rounded-lg font-medium hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 transition-colors"
                >
                  {submitting ? (
                    <>
                      <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                      Processing...
                    </>
                  ) : (
                    <>
                      <CheckCircle size={18} />
                      Record Payment
                    </>
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

export default function PaymentsPage() {
  return (
    <ProtectedRoute allowedRoles={['admin']}>
      <AdminPaymentsPage />
    </ProtectedRoute>
  );
}