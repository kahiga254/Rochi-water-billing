"use client";

import { useState, useEffect } from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { 
  FileText, 
  Search, 
  Download, 
  Filter,
  ChevronLeft,
  ChevronRight,
  Eye,
  Send,
  CheckCircle,
  Clock,
  AlertCircle,
  DollarSign,
  TrendingUp,
  Calendar
} from 'lucide-react';
import { billsApi } from '@/lib/api/bills';
import { toast } from 'sonner';
import { format } from 'date-fns';

interface Bill {
  id: string;
  bill_number: string;
  meter_number: string;
  customer_name: string;
  billing_period: string;
  total_amount: number;
  amount_paid: number;
  balance: number;
  status: 'paid' | 'pending' | 'overdue';
  due_date: string;
  consumption: number;
}

function BillsContent() {
  const [bills, setBills] = useState<Bill[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [summary, setSummary] = useState({
    totalAmount: 0,
    pendingAmount: 0,
    overdueAmount: 0,
    paidAmount: 0
  });

  useEffect(() => {
    fetchBills();
    fetchSummary();
  }, [page, statusFilter]);

  const fetchBills = async () => {
    try {
      setLoading(true);
      const response = await billsApi.getAll(page, 20, statusFilter);
      if (response.success) {
        setBills(response.data || []);
        // Assuming your API returns pagination info
        setTotalPages(response.total_pages || 1);
      }
    } catch (error) {
      toast.error('Failed to load bills');
    } finally {
      setLoading(false);
    }
  };

  const fetchSummary = async () => {
    try {
      const response = await billsApi.getSummary();
      if (response.success) {
        const data = response.data;
        setSummary({
          totalAmount: data.total_amount || 0,
          pendingAmount: data.pending_amount || 0,
          overdueAmount: data.overdue_amount || 0,
          paidAmount: data.paid_amount || 0
        });
      }
    } catch (error) {
      console.error('Failed to fetch summary:', error);
    }
  };

  const handleSendNotification = async (billId: string) => {
    try {
      const response = await billsApi.sendNotification(billId);
      if (response.success) {
        toast.success('Notification sent successfully');
      }
    } catch (error) {
      toast.error('Failed to send notification');
    }
  };

  const handleExport = async () => {
    try {
      const response = await billsApi.exportToCSV({ status: statusFilter });
      // Create download link
      const url = window.URL.createObjectURL(new Blob([response]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `bills-${format(new Date(), 'yyyy-MM-dd')}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (error) {
      toast.error('Failed to export bills');
    }
  };

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

  const filteredBills = bills.filter(bill => 
    bill.bill_number?.toLowerCase().includes(search.toLowerCase()) ||
    bill.customer_name?.toLowerCase().includes(search.toLowerCase()) ||
    bill.meter_number?.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Bills Management</h1>
          <p className="text-gray-600 mt-1">View and manage all water bills</p>
        </div>
        <button
          onClick={handleExport}
          className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
        >
          <Download size={20} />
          Export CSV
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Total Bills</p>
              <p className="text-2xl font-bold text-gray-900 mt-2">
                KSh {summary.totalAmount.toLocaleString()}
              </p>
            </div>
            <div className="p-3 bg-blue-100 rounded-lg">
              <DollarSign className="w-6 h-6 text-blue-600" />
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Pending</p>
              <p className="text-2xl font-bold text-yellow-600 mt-2">
                KSh {summary.pendingAmount.toLocaleString()}
              </p>
            </div>
            <div className="p-3 bg-yellow-100 rounded-lg">
              <Clock className="w-6 h-6 text-yellow-600" />
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Overdue</p>
              <p className="text-2xl font-bold text-red-600 mt-2">
                KSh {summary.overdueAmount.toLocaleString()}
              </p>
            </div>
            <div className="p-3 bg-red-100 rounded-lg">
              <AlertCircle className="w-6 h-6 text-red-600" />
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Collected</p>
              <p className="text-2xl font-bold text-green-600 mt-2">
                KSh {summary.paidAmount.toLocaleString()}
              </p>
            </div>
            <div className="p-3 bg-green-100 rounded-lg">
              <TrendingUp className="w-6 h-6 text-green-600" />
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg border p-4">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search by bill number, customer, or meter..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          
          <div className="flex gap-2">
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="all">All Status</option>
              <option value="pending">Pending</option>
              <option value="paid">Paid</option>
              <option value="overdue">Overdue</option>
            </select>
            
            <button className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 flex items-center gap-2">
              <Calendar size={18} />
              Date Range
            </button>
          </div>
        </div>
      </div>

      {/* Bills Table */}
      <div className="bg-white rounded-lg border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Bill #</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Customer</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Period</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Paid</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Balance</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Due Date</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {loading ? (
                <tr>
                  <td colSpan={9} className="px-6 py-8 text-center">
                    <div className="flex justify-center">
                      <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                    </div>
                  </td>
                </tr>
              ) : filteredBills.length === 0 ? (
                <tr>
                  <td colSpan={9} className="px-6 py-8 text-center text-gray-500">
                    No bills found
                  </td>
                </tr>
              ) : (
                filteredBills.map((bill) => (
                  <tr key={bill.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 font-mono text-sm">{bill.bill_number}</td>
                    <td className="px-6 py-4">
                      <div className="font-medium">{bill.customer_name}</div>
                      <div className="text-xs text-gray-500">{bill.meter_number}</div>
                    </td>
                    <td className="px-6 py-4 text-sm">{bill.billing_period}</td>
                    <td className="px-6 py-4 text-sm font-medium">KSh {bill.total_amount.toLocaleString()}</td>
                    <td className="px-6 py-4 text-sm">KSh {bill.amount_paid.toLocaleString()}</td>
                    <td className="px-6 py-4 text-sm font-medium">
                      {bill.balance > 0 ? (
                        <span className="text-red-600">KSh {bill.balance.toLocaleString()}</span>
                      ) : (
                        <span className="text-green-600">KSh 0</span>
                      )}
                    </td>
                    <td className="px-6 py-4 text-sm">{format(new Date(bill.due_date), 'dd MMM yyyy')}</td>
                    <td className="px-6 py-4">{getStatusBadge(bill.status)}</td>
                    <td className="px-6 py-4 text-right">
                      <button
                        onClick={() => window.location.href = `/bills/${bill.id}`}
                        className="text-blue-600 hover:text-blue-900 mr-3"
                        title="View Details"
                      >
                        <Eye size={18} />
                      </button>
                      {bill.status !== 'paid' && (
                        <button
                          onClick={() => handleSendNotification(bill.id)}
                          className="text-green-600 hover:text-green-900"
                          title="Send Reminder"
                        >
                          <Send size={18} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        <div className="px-6 py-4 border-t flex items-center justify-between">
          <p className="text-sm text-gray-600">
            Page {page} of {totalPages}
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="p-2 border rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft size={18} />
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="p-2 border rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight size={18} />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function BillsPage() {
  return (
    <ProtectedRoute allowedRoles={['admin']}>
      <BillsContent />
    </ProtectedRoute>
  );
}