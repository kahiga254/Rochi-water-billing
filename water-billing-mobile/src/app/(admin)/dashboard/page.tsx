
"use client";
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/lib/hooks/useAuth';
import { Card, CardContent, CardHeader } from '@/components/ui/Card';
import { 
  Users, 
  Droplets, 
  CreditCard, 
  TrendingUp, 
  AlertCircle,
  CheckCircle,
  UserPlus,
  FileText,
  Download,
  MoreHorizontal
} from 'lucide-react';
import { useState, useEffect } from 'react';

// Simple loading component
function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <div className="w-16 h-16 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p className="text-gray-600 dark:text-gray-400">Loading dashboard...</p>
      </div>
    </div>
  );
}

// Main dashboard content
function AdminDashboardContent() {
  const { user } = useAuth();
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
    // Simulate loading data
    const timer = setTimeout(() => {
      setStats({
        totalCustomers: 156,
        activeCustomers: 142,
        pendingBills: 23,
        overdueBills: 8,
        totalRevenue: 245000,
        collectedRevenue: 189000,
      });
      setLoading(false);
    }, 1000);

    return () => clearTimeout(timer);
  }, []);

  const collectionRate = stats.totalRevenue > 0 
    ? Math.round((stats.collectedRevenue / stats.totalRevenue) * 100) 
    : 0;

  const recentActivities = [
    { id: 1, type: 'new_customer', title: 'New customer registered', user: 'Alice Johnson', time: '5 minutes ago', status: 'success' },
    { id: 2, type: 'payment', title: 'Payment received', user: 'Bob Smith', time: '15 minutes ago', status: 'success' },
    { id: 3, type: 'meter_reading', title: 'Meter reading submitted', user: 'Carol White', time: '1 hour ago', status: 'pending' },
    { id: 4, type: 'bill_generated', title: 'Bill generated', user: 'David Brown', time: '2 hours ago', status: 'success' },
  ];

  if (loading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="space-y-6 p-6">
      {/* Welcome Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            Welcome back, {user?.first_name || 'Admin'}! 👋
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Here's what's happening with your water billing system
          </p>
        </div>
        <div className="mt-4 sm:mt-0 flex gap-3">
          <button className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg flex items-center gap-2">
            <FileText className="w-4 h-4" />
            Generate Report
          </button>
          <button className="px-4 py-2 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg flex items-center gap-2">
            <Download className="w-4 h-4" />
            Export Data
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Total Customers</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                  {stats.totalCustomers}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  {stats.activeCustomers} active
                </p>
              </div>
              <div className="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                <Users className="w-6 h-6 text-blue-600 dark:text-blue-400" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Pending Bills</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                  {stats.pendingBills}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  {stats.overdueBills} overdue
                </p>
              </div>
              <div className="p-3 bg-yellow-100 dark:bg-yellow-900/30 rounded-lg">
                <AlertCircle className="w-6 h-6 text-yellow-600 dark:text-yellow-400" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Monthly Revenue</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                  KSh {stats.totalRevenue.toLocaleString()}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  Collected: KSh {stats.collectedRevenue.toLocaleString()}
                </p>
              </div>
              <div className="p-3 bg-green-100 dark:bg-green-900/30 rounded-lg">
                <CreditCard className="w-6 h-6 text-green-600 dark:text-green-400" />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Collection Rate</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                  {collectionRate}%
                </p>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 mt-2">
                  <div 
                    className="bg-blue-600 rounded-full h-2" 
                    style={{ width: `${collectionRate}%` }}
                  />
                </div>
              </div>
              <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                <TrendingUp className="w-6 h-6 text-purple-600 dark:text-purple-400" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Quick Actions</h2>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <button className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/40">
              <UserPlus className="w-6 h-6 text-blue-600 dark:text-blue-400 mx-auto mb-2" />
              <span className="text-xs font-medium text-gray-700 dark:text-gray-300">Add Customer</span>
            </button>
            <button className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg hover:bg-green-100 dark:hover:bg-green-900/40">
              <Droplets className="w-6 h-6 text-green-600 dark:text-green-400 mx-auto mb-2" />
              <span className="text-xs font-medium text-gray-700 dark:text-gray-300">Meter Reading</span>
            </button>
            <button className="p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg hover:bg-yellow-100 dark:hover:bg-yellow-900/40">
              <CreditCard className="w-6 h-6 text-yellow-600 dark:text-yellow-400 mx-auto mb-2" />
              <span className="text-xs font-medium text-gray-700 dark:text-gray-300">Record Payment</span>
            </button>
            <button className="p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg hover:bg-purple-100 dark:hover:bg-purple-900/40">
              <FileText className="w-6 h-6 text-purple-600 dark:text-purple-400 mx-auto mb-2" />
              <span className="text-xs font-medium text-gray-700 dark:text-gray-300">Generate Bills</span>
            </button>
          </div>
        </CardContent>
      </Card>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Recent Activity</h2>
            <button className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300">
              View All
            </button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {recentActivities.map((activity) => (
              <div key={activity.id} className="flex items-start gap-4 p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                <div className={`p-2 rounded-lg ${
                  activity.status === 'success' 
                    ? 'bg-green-100 dark:bg-green-900/30' 
                    : 'bg-yellow-100 dark:bg-yellow-900/30'
                }`}>
                  {activity.type === 'new_customer' && <UserPlus className="w-4 h-4 text-green-600 dark:text-green-400" />}
                  {activity.type === 'payment' && <CreditCard className="w-4 h-4 text-green-600 dark:text-green-400" />}
                  {activity.type === 'meter_reading' && <Droplets className="w-4 h-4 text-yellow-600 dark:text-yellow-400" />}
                  {activity.type === 'bill_generated' && <FileText className="w-4 h-4 text-green-600 dark:text-green-400" />}
                </div>
                <div className="flex-1">
                  <p className="text-sm font-medium text-gray-900 dark:text-white">{activity.title}</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    {activity.user} • {activity.time}
                  </p>
                </div>
                <button className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg">
                  <MoreHorizontal className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                </button>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* System Status */}
      <Card>
        <CardHeader>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">System Status</h2>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
              <CheckCircle className="w-5 h-5 text-green-600 dark:text-green-400" />
              <div>
                <p className="text-sm font-medium text-gray-900 dark:text-white">Database</p>
                <p className="text-xs text-green-600 dark:text-green-400">Connected</p>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
              <CheckCircle className="w-5 h-5 text-green-600 dark:text-green-400" />
              <div>
                <p className="text-sm font-medium text-gray-900 dark:text-white">API Server</p>
                <p className="text-xs text-green-600 dark:text-green-400">Online</p>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
              <AlertCircle className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
              <div>
                <p className="text-sm font-medium text-gray-900 dark:text-white">SMS Service</p>
                <p className="text-xs text-yellow-600 dark:text-yellow-400">Mock Mode</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// Export with ProtectedRoute wrapper
export default function AdminDashboardPage() {
  return (
    <ProtectedRoute allowedRoles={['admin']}>
      <AdminDashboardContent />
    </ProtectedRoute>
  );
}