"use client";

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { billsApi } from '@/lib/api/bills';
import { toast } from 'sonner';
import { 
  ArrowLeft, 
  FileText, 
  Calendar, 
  User, 
  Zap, 
  DollarSign,
  Clock,
  CheckCircle,
  AlertCircle,
  Download,
  Send,
  Printer
} from 'lucide-react';
import { format } from 'date-fns';

interface Bill {
  id: string;
  bill_number: string;
  meter_number: string;
  customer_name: string;
  customer_id?: string;
  account_number?: string;
  billing_period: string;
  total_amount: number;
  amount_paid: number;
  balance: number;
  status: 'paid' | 'pending' | 'overdue';
  due_date: string;
  issue_date?: string;
  consumption: number;
  previous_reading?: number;
  current_reading?: number;
  rate_per_unit?: number;
  water_charge?: number;
  fixed_charge?: number;
  arrears?: number;
  created_at?: string;
}

function BillDetailsContent() {
  const params = useParams();
  const router = useRouter();
  const [bill, setBill] = useState<Bill | null>(null);
  const [loading, setLoading] = useState(true);
  const [sendingReminder, setSendingReminder] = useState(false);

  const billId = params.id as string;

  useEffect(() => {
    if (billId) {
      fetchBillDetails();
    }
  }, [billId]);

  const fetchBillDetails = async () => {
    try {
      setLoading(true);
      const response = await billsApi.getById(billId);
      
      if (response.success) {
        setBill(response.data);
      } else {
        toast.error(response.message || 'Failed to load bill details');
      }
    } catch (error) {
      console.error('Error fetching bill:', error);
      toast.error('Failed to load bill details');
    } finally {
      setLoading(false);
    }
  };

  const handleSendReminder = async () => {
    if (!bill) return;
    
    setSendingReminder(true);
    try {
      const response = await billsApi.sendNotification(bill.id);
      if (response.success) {
        toast.success(`Reminder sent for bill ${bill.bill_number}`);
      } else {
        toast.error(response.message || 'Failed to send reminder');
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to send reminder');
    } finally {
      setSendingReminder(false);
    }
  };

  const handlePrint = () => {
    window.print();
  };

  const handleDownload = () => {
    // You can implement PDF download here
    toast.info('PDF download coming soon');
  };

  const getStatusBadge = (status: string) => {
    switch(status) {
      case 'paid':
        return <span className="flex items-center gap-1 px-3 py-1 bg-green-100 text-green-700 rounded-full text-sm"><CheckCircle size={16} /> Paid</span>;
      case 'pending':
        return <span className="flex items-center gap-1 px-3 py-1 bg-yellow-100 text-yellow-700 rounded-full text-sm"><Clock size={16} /> Pending</span>;
      case 'overdue':
        return <span className="flex items-center gap-1 px-3 py-1 bg-red-100 text-red-700 rounded-full text-sm"><AlertCircle size={16} /> Overdue</span>;
      default:
        return <span className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm">{status}</span>;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading bill details...</p>
        </div>
      </div>
    );
  }

  if (!bill) {
    return (
      <div className="text-center py-12">
        <FileText className="w-16 h-16 text-gray-400 mx-auto mb-4" />
        <h2 className="text-xl font-semibold text-gray-900 mb-2">Bill Not Found</h2>
        <p className="text-gray-600 mb-4">The bill you're looking for doesn't exist.</p>
        <button
          onClick={() => router.push('/bills')}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          Back to Bills
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={() => router.back()}
            className="p-2 hover:bg-gray-100 rounded-lg"
          >
            <ArrowLeft size={20} />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Bill Details</h1>
            <p className="text-gray-600 mt-1">{bill.bill_number}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {bill.status !== 'paid' && (
            <button
              onClick={handleSendReminder}
              disabled={sendingReminder}
              className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
            >
              <Send size={18} />
              {sendingReminder ? 'Sending...' : 'Send Reminder'}
            </button>
          )}
          <button
            onClick={handlePrint}
            className="p-2 border rounded-lg hover:bg-gray-50"
            title="Print"
          >
            <Printer size={18} />
          </button>
          <button
            onClick={handleDownload}
            className="p-2 border rounded-lg hover:bg-gray-50"
            title="Download PDF"
          >
            <Download size={18} />
          </button>
        </div>
      </div>

      {/* Status Card */}
      <div className="bg-white rounded-lg border p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-500 mb-1">Bill Status</p>
            {getStatusBadge(bill.status)}
          </div>
          <div className="text-right">
            <p className="text-sm text-gray-500 mb-1">Due Date</p>
            <p className="text-lg font-semibold text-gray-900">
              {format(new Date(bill.due_date), 'dd MMMM yyyy')}
            </p>
          </div>
        </div>
      </div>

      {/* Bill Details Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Customer Information */}
        <div className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <User size={20} className="text-blue-600" />
            Customer Information
          </h2>
          <div className="space-y-3">
            <div>
              <p className="text-sm text-gray-500">Customer Name</p>
              <p className="font-medium">{bill.customer_name}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Meter Number</p>
              <p className="font-mono">{bill.meter_number}</p>
            </div>
            {bill.account_number && (
              <div>
                <p className="text-sm text-gray-500">Account Number</p>
                <p className="font-mono">{bill.account_number}</p>
              </div>
            )}
          </div>
        </div>

        {/* Billing Period */}
        <div className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <Calendar size={20} className="text-blue-600" />
            Billing Period
          </h2>
          <div className="space-y-3">
            <div>
              <p className="text-sm text-gray-500">Period</p>
              <p className="font-medium">{bill.billing_period}</p>
            </div>
            {bill.issue_date && (
              <div>
                <p className="text-sm text-gray-500">Issue Date</p>
                <p>{format(new Date(bill.issue_date), 'dd MMM yyyy')}</p>
              </div>
            )}
          </div>
        </div>

        {/* Consumption Details */}
        <div className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <Zap size={20} className="text-blue-600" />
            Consumption Details
          </h2>
          <div className="space-y-3">
            <div>
              <p className="text-sm text-gray-500">Consumption</p>
              <p className="font-medium">{bill.consumption} m³</p>
            </div>
            {bill.previous_reading && (
              <div>
                <p className="text-sm text-gray-500">Previous Reading</p>
                <p>{bill.previous_reading}</p>
              </div>
            )}
            {bill.current_reading && (
              <div>
                <p className="text-sm text-gray-500">Current Reading</p>
                <p>{bill.current_reading}</p>
              </div>
            )}
            {bill.rate_per_unit && (
              <div>
                <p className="text-sm text-gray-500">Rate per m³</p>
                <p>KSh {bill.rate_per_unit.toLocaleString()}</p>
              </div>
            )}
          </div>
        </div>

        {/* Charges Breakdown */}
        <div className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <DollarSign size={20} className="text-blue-600" />
            Charges Breakdown
          </h2>
          <div className="space-y-3">
            {bill.water_charge && (
              <div className="flex justify-between">
                <p className="text-gray-600">Water Charge</p>
                <p className="font-medium">KSh {bill.water_charge.toLocaleString()}</p>
              </div>
            )}
            {bill.fixed_charge && bill.fixed_charge > 0 && (
              <div className="flex justify-between">
                <p className="text-gray-600">Fixed Charge</p>
                <p className="font-medium">KSh {bill.fixed_charge.toLocaleString()}</p>
              </div>
            )}
            {bill.arrears && bill.arrears > 0 && (
              <div className="flex justify-between">
                <p className="text-gray-600">Arrears</p>
                <p className="font-medium text-red-600">KSh {bill.arrears.toLocaleString()}</p>
              </div>
            )}
            <div className="border-t pt-3 mt-3">
              <div className="flex justify-between font-semibold">
                <p>Total Amount</p>
                <p>KSh {bill.total_amount.toLocaleString()}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Payment Summary */}
      <div className="bg-white rounded-lg border p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Payment Summary</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div>
            <p className="text-sm text-gray-500 mb-1">Total Amount</p>
            <p className="text-xl font-bold">KSh {bill.total_amount.toLocaleString()}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">Amount Paid</p>
            <p className="text-xl font-bold text-green-600">KSh {bill.amount_paid.toLocaleString()}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500 mb-1">Balance Due</p>
            <p className={`text-xl font-bold ${bill.balance > 0 ? 'text-red-600' : 'text-green-600'}`}>
              KSh {bill.balance.toLocaleString()}
            </p>
          </div>
        </div>
      </div>

      {/* Actions */}
      {bill.status !== 'paid' && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <p className="text-yellow-800 mb-2">
            This bill is {bill.status === 'overdue' ? 'overdue' : 'pending payment'}.
          </p>
          <button
            onClick={handleSendReminder}
            disabled={sendingReminder}
            className="px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 text-sm disabled:opacity-50"
          >
            {sendingReminder ? 'Sending...' : 'Send Payment Reminder'}
          </button>
        </div>
      )}
    </div>
  );
}

export default function BillDetailsPage() {
  return (
    <ProtectedRoute allowedRoles={['admin']}>
      <BillDetailsContent />
    </ProtectedRoute>
  );
}