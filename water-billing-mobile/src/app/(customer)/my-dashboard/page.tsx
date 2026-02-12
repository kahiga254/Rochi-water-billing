'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Droplets, 
  Zap, 
  TrendingUp, 
  Clock, 
  AlertCircle,
  Home,
  CreditCard
} from 'lucide-react';
import { useAuth } from '@/lib/hooks/useAuth';
import { useDashboardStore } from '@/lib/store/dashboardStore';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { StatsCard } from '@/components/customer/StatsCard';
import { ConsumptionChart } from '@/components/customer/ConsumptionChart';
import { RecentBills } from '@/components/customer/RecentBills';
import { Loading } from '@/components/common/Loading';

function CustomerDashboardContent() {
  const router = useRouter();
  const { user } = useAuth();
  const {
    customer,
    currentBill,
    recentBills,
    readingHistory,
    isLoading,
    error,
    fetchCustomerData,
    fetchBills,
    fetchReadingHistory,
  } = useDashboardStore();

  useEffect(() => {
    if (user?.meter_number) {
      fetchCustomerData(user.meter_number);
      fetchBills(user.meter_number);
      fetchReadingHistory(user.meter_number);
    }
  }, [user, fetchCustomerData, fetchBills, fetchReadingHistory]);

  if (isLoading) {
    return <Loading fullScreen />;
  }

  const getDaysUntilDue = () => {
    if (!currentBill?.due_date) return null;
    const dueDate = new Date(currentBill.due_date);
    const today = new Date();
    const diffTime = dueDate.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  };

  const getConsumptionTrend = () => {
    if (readingHistory.length < 2) return undefined;
    const lastMonth = readingHistory[0]?.consumption || 0;
    const prevMonth = readingHistory[1]?.consumption || 0;
    if (prevMonth === 0) return undefined;
    const change = ((lastMonth - prevMonth) / prevMonth) * 100;
    return {
      value: Math.round(Math.abs(change) * 10) / 10,
      isPositive: change < 0,
    };
  };

  const daysUntilDue = getDaysUntilDue();
  const consumptionTrend = getConsumptionTrend();

  return (
    <div className="space-y-6">
      {/* Welcome Section */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            Welcome back, {customer?.first_name || user?.first_name}!
          </h1>
          <div className="flex items-center gap-2 mt-1">
            <Home className="w-4 h-4 text-gray-500 dark:text-gray-400" />
            <p className="text-gray-600 dark:text-gray-400">
              Meter: <span className="font-mono font-medium">{customer?.meter_number || user?.meter_number}</span>
            </p>
            <span className="text-gray-400 dark:text-gray-600">•</span>
            <p className="text-gray-600 dark:text-gray-400">
              Zone: {customer?.zone || 'N/A'}
            </p>
          </div>
        </div>
        
        {currentBill?.status === 'pending' && (
          <button
            onClick={() => router.push(`/payments?bill=${currentBill.id}`)}
            className="mt-4 sm:mt-0 inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
          >
            <CreditCard className="w-4 h-4" />
            Pay Now
          </button>
        )}
      </div>

      {/* Error State */}
      {error && (
        <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg flex items-start gap-3">
          <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
          <div className="flex-1">
            <p className="text-sm font-medium text-red-600 dark:text-red-400">
              Failed to load dashboard data
            </p>
            <p className="text-sm text-red-500 dark:text-red-500 mt-1">
              {error}
            </p>
          </div>
        </div>
      )}

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title="Current Balance"
          value={customer?.balance || 0}
          icon={Droplets}
          color="blue"
          prefix="KSh"
        />

        <StatsCard
          title="This Month's Bill"
          value={currentBill?.total_amount || 0}
          icon={Zap}
          color="yellow"
          prefix="KSh"
        />

        <StatsCard
          title="Current Reading"
          value={customer?.last_reading?.toFixed(1) || '0.0'}
          icon={TrendingUp}
          color="green"
          suffix="m³"
          trend={consumptionTrend}
        />

        <StatsCard
          title={daysUntilDue && daysUntilDue > 0 ? 'Days Until Due' : 'Overdue'}
          value={daysUntilDue && daysUntilDue > 0 ? daysUntilDue : 'Due'}
          icon={daysUntilDue && daysUntilDue > 0 ? Clock : AlertCircle}
          color={daysUntilDue && daysUntilDue > 0 ? 'orange' : 'red'}
          suffix={daysUntilDue && daysUntilDue > 0 ? 'days' : ''}
        />
      </div>

      {/* Consumption Chart */}
      <ConsumptionChart 
        data={readingHistory} 
        isLoading={isLoading} 
      />

      {/* Recent Bills */}
      <RecentBills 
        bills={recentBills} 
        isLoading={isLoading} 
      />
    </div>
  );
}

export default function CustomerDashboardPage() {
  return (
    <ProtectedRoute allowedRoles={['customer']}>
      <CustomerDashboardContent />
    </ProtectedRoute>
  );
}