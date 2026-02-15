"use client";

import { useState, useEffect } from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/lib/hooks/useAuth';
import { customerApi } from '@/lib/api/customer';
import { 
  FileText, 
  Search,
  CheckCircle,
  Clock,
  AlertCircle,
  Download,
  Eye
} from 'lucide-react';
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

function CustomerBills() {
  const { user } = useAuth();
  const [bills, setBills] = useState<Bill[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState('all');

  const meterNumber = user?.meter_number || '';

  useEffect(() => {
    if (meterNumber) {
      fetchBills();
    }
  }, [meterNumber]);

  const fetchBills = async () => {
    try {
      setLoading(true);
      const response = await customerApi.getBills(meterNumber);
      if (response.success) {
        setBills(response.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch bills:', error);
    } finally {
      setLoading(false);
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

  const filteredBills = bills.filter(bill => {
    const matchesSearch = 
      bill.bill_number.toLowerCase().includes(search.toLowerCase()) ||
      bill.billing_period.toLowerCase().includes(search.toLowerCase());
    
    if (filter === 'all') return matchesSearch;
    return matchesSearch && bill.status === filter;
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">My Bills</h1>
        <p className="text-gray-600 mt-1">View your water bill history</p>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg border p-4">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search by bill number or period..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All Bills</option>
            <option value="paid">Paid</option>
            <option value="pending">Pending</option>
            <option value="overdue">Overdue</option>
          </select>
        </div>
      </div>

      {/* Bills List */}
      <div className="bg-white rounded-lg border overflow-hidden">
        {loading ? (
          <div className="text-center py-12">
            <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto"></div>
          </div>
        ) : filteredBills.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            No bills found
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Bill #</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Period</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Consumption</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Paid</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Balance</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Due Date</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {filteredBills.map((bill) => (
                  <tr key={bill.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm font-mono">{bill.bill_number}</td>
                    <td className="px-6 py-4 text-sm">{bill.billing_period}</td>
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
                    <td className="px-6 py-4 text-sm">
                      {bill.due_date ? format(new Date(bill.due_date), 'dd MMM yyyy') : 'N/A'}
                    </td>
                    <td className="px-6 py-4">{getStatusBadge(bill.status)}</td>
                    <td className="px-6 py-4 text-right">
                      <button className="text-blue-600 hover:text-blue-900 mr-3">
                        <Eye size={18} />
                      </button>
                      <button className="text-green-600 hover:text-green-900">
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

export default function CustomerBillsPage() {
  return (
    <ProtectedRoute allowedRoles={['customer_service']}>
      <CustomerBills />
    </ProtectedRoute>
  );
}