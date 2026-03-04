"use client";

import { useState, useEffect } from 'react';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { dashboardApi } from '@/lib/api/dashboard';
import { Users, FileText, CreditCard, AlertCircle } from 'lucide-react';

function AdminDashboard() {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({
    totalCustomers: 0,
    activeCustomers: 0,
    pendingBills: 0,
    overdueBills: 0,
    totalRevenue: 0,
    collectedRevenue: 0,
  });

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      
      // Fetch customer statistics
      const customerStats = await dashboardApi.getCustomerStats();
      if (customerStats.success) {
        setStats(prev => ({
          ...prev,
          totalCustomers: customerStats.data.total || 0,
          activeCustomers: customerStats.data.active || 0,
        }));
      }

      // Fetch unpaid bills
      const unpaidBills = await dashboardApi.getUnpaidBills();
      if (unpaidBills.success) {
        const bills = unpaidBills.data || [];
        const pending = bills.filter((b: any) => b.status === 'pending');
        const overdue = bills.filter((b: any) => b.status === 'overdue');
        
        setStats(prev => ({
          ...prev,
          pendingBills: pending.length,
          overdueBills: overdue.length,
        }));
      }

      // Fetch billing summary for revenue
      const billingSummary = await dashboardApi.getBillingSummary();
      if (billingSummary.success) {
        // Calculate total revenue from summary
        const summary = billingSummary.data;
        let total = 0;
        let collected = 0;
        
        if (summary?.status_breakdown) {
          Object.values(summary.status_breakdown).forEach((item: any) => {
            total += item.total_amount || 0;
            collected += item.total_paid || 0;
          });
        }
        
        setStats(prev => ({
          ...prev,
          totalRevenue: total,
          collectedRevenue: collected,
        }));
      }

    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const collectionRate = stats.totalRevenue > 0 
    ? Math.round((stats.collectedRevenue / stats.totalRevenue) * 100) 
    : 0;

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Admin Dashboard</h1>
        <p className="text-gray-600 mt-2">Welcome to Rochi Pure water billing system</p>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Customers Card */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Total Customers</p>
              <p className="text-3xl font-bold text-blue-600 mt-2">{stats.totalCustomers}</p>
              <p className="text-xs text-gray-500 mt-1">{stats.activeCustomers} active</p>
            </div>
            <div className="p-3 bg-blue-100 rounded-lg">
              <Users className="w-6 h-6 text-blue-600" />
            </div>
          </div>
        </div>

        {/* Pending Bills Card */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Pending Bills</p>
              <p className="text-3xl font-bold text-yellow-600 mt-2">{stats.pendingBills}</p>
              <p className="text-xs text-gray-500 mt-1">{stats.overdueBills} overdue</p>
            </div>
            <div className="p-3 bg-yellow-100 rounded-lg">
              <FileText className="w-6 h-6 text-yellow-600" />
            </div>
          </div>
        </div>

        {/* Revenue Card */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Monthly Revenue</p>
              <p className="text-3xl font-bold text-green-600 mt-2">KSh {stats.totalRevenue.toLocaleString()}</p>
              <p className="text-xs text-gray-500 mt-1">Collected: KSh {stats.collectedRevenue.toLocaleString()}</p>
            </div>
            <div className="p-3 bg-green-100 rounded-lg">
              <CreditCard className="w-6 h-6 text-green-600" />
            </div>
          </div>
        </div>

        {/* Collection Rate Card */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Collection Rate</p>
              <p className="text-3xl font-bold text-purple-600 mt-2">{collectionRate}%</p>
              <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                <div 
                  className="bg-purple-600 rounded-full h-2" 
                  style={{ width: `${collectionRate}%` }}
                />
              </div>
            </div>
            <div className="p-3 bg-purple-100 rounded-lg">
              <AlertCircle className="w-6 h-6 text-purple-600" />
            </div>
          </div>
        </div>
      </div>

      {/* Recent Activity Section */}
      <div className="bg-white p-6 rounded-lg shadow-sm border mt-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h2>
        {stats.overdueBills > 0 && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <p className="text-red-700">
              ⚠️ {stats.overdueBills} overdue bill{stats.overdueBills > 1 ? 's' : ''} require attention
            </p>
          </div>
        )}
        {stats.pendingBills === 0 && stats.overdueBills === 0 && (
          <p className="text-gray-500">No pending activity</p>
        )}
      </div>
    </div>
  );
}

export default function AdminDashboardPage() {
  return (
    <ProtectedRoute allowedRoles={['admin']}>
      <AdminDashboard />
    </ProtectedRoute>
  );
}