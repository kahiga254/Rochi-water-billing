'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { 
  FileText, 
  Download, 
  Eye, 
  ChevronRight,
  CheckCircle,
  Clock,
  AlertCircle,
  CreditCard
} from 'lucide-react';
import { Card, CardContent, CardHeader } from '@/components/ui/Card';
import { Bill } from '@/types';
import { cn } from '@/lib/utils/cn';
import { format } from 'date-fns';

interface RecentBillsProps {
  bills: Bill[];
  isLoading?: boolean;
}

const statusConfig = {
  paid: {
    icon: CheckCircle,
    text: 'Paid',
    className: 'bg-green-100 dark:bg-green-900/50 text-green-700 dark:text-green-400',
  },
  pending: {
    icon: Clock,
    text: 'Pending',
    className: 'bg-yellow-100 dark:bg-yellow-900/50 text-yellow-700 dark:text-yellow-400',
  },
  overdue: {
    icon: AlertCircle,
    text: 'Overdue',
    className: 'bg-red-100 dark:bg-red-900/50 text-red-700 dark:text-red-400',
  },
  partially_paid: {
    icon: CreditCard,
    text: 'Partial',
    className: 'bg-blue-100 dark:bg-blue-900/50 text-blue-700 dark:text-blue-400',
  },
};

export function RecentBills({ bills, isLoading }: RecentBillsProps) {
  const router = useRouter();
  const [expandedBill, setExpandedBill] = useState<string | null>(null);

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <div className="h-6 w-32 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
        </CardHeader>
        <CardContent>
          {[1, 2, 3].map((i) => (
            <div key={i} className="mb-4 p-4 bg-gray-100 dark:bg-gray-800 rounded-lg animate-pulse">
              <div className="h-5 w-40 bg-gray-200 dark:bg-gray-700 rounded mb-2" />
              <div className="h-4 w-24 bg-gray-200 dark:bg-gray-700 rounded" />
            </div>
          ))}
        </CardContent>
      </Card>
    );
  }

  if (!bills || bills.length === 0) {
    return (
      <Card>
        <CardHeader>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
            Recent Bills
          </h3>
        </CardHeader>
        <CardContent>
          <div className="text-center py-12">
            <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600 dark:text-gray-400">
              No bills found
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
          Recent Bills
        </h3>
        <button
          onClick={() => router.push('/bills')}
          className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 font-medium flex items-center gap-1"
        >
          View All
          <ChevronRight className="w-4 h-4" />
        </button>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {bills.slice(0, 5).map((bill) => {
            const status = statusConfig[bill.status as keyof typeof statusConfig] || statusConfig.pending;
            const StatusIcon = status.icon;
            const isExpanded = expandedBill === bill.id;

            return (
              <div
                key={bill.id}
                className="group border border-gray-200 dark:border-gray-700 rounded-lg hover:border-blue-200 dark:hover:border-blue-800 transition-colors"
              >
                <div
                  className="p-4 cursor-pointer"
                  onClick={() => setExpandedBill(isExpanded ? null : bill.id)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <span className={cn(
                          'inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium',
                          status.className
                        )}>
                          <StatusIcon className="w-3 h-3" />
                          {status.text}
                        </span>
                        <span className="text-sm font-medium text-gray-900 dark:text-white">
                          {bill.bill_number}
                        </span>
                      </div>
                      
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <p className="text-gray-500 dark:text-gray-400">Period</p>
                          <p className="font-medium text-gray-900 dark:text-white">
                            {bill.billing_period}
                          </p>
                        </div>
                        <div>
                          <p className="text-gray-500 dark:text-gray-400">Due Date</p>
                          <p className="font-medium text-gray-900 dark:text-white">
                            {format(new Date(bill.due_date), 'dd MMM yyyy')}
                          </p>
                        </div>
                      </div>
                    </div>
                    
                    <div className="text-right">
                      <p className="text-lg font-bold text-gray-900 dark:text-white">
                        KSh {bill.total_amount.toLocaleString()}
                      </p>
                      {bill.status === 'pending' && (
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                          Balance: KSh {bill.balance.toLocaleString()}
                        </p>
                      )}
                    </div>
                  </div>

                  {isExpanded && (
                    <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                      <div className="grid grid-cols-2 gap-4 text-sm mb-4">
                        <div>
                          <p className="text-gray-500 dark:text-gray-400">Previous Reading</p>
                          <p className="font-medium text-gray-900 dark:text-white">
                            {bill.previous_reading.toFixed(1)} m³
                          </p>
                        </div>
                        <div>
                          <p className="text-gray-500 dark:text-gray-400">Current Reading</p>
                          <p className="font-medium text-gray-900 dark:text-white">
                            {bill.current_reading.toFixed(1)} m³
                          </p>
                        </div>
                        <div>
                          <p className="text-gray-500 dark:text-gray-400">Consumption</p>
                          <p className="font-medium text-gray-900 dark:text-white">
                            {bill.consumption.toFixed(1)} m³
                          </p>
                        </div>
                        <div>
                          <p className="text-gray-500 dark:text-gray-400">Rate</p>
                          <p className="font-medium text-gray-900 dark:text-white">
                            KSh {bill.rate_per_unit}/m³
                          </p>
                        </div>
                      </div>

                      <div className="flex gap-3">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            router.push(`/bills/${bill.id}`);
                          }}
                          className="flex-1 px-3 py-2 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 transition-colors flex items-center justify-center gap-2"
                        >
                          <Eye className="w-4 h-4" />
                          View Details
                        </button>
                        {bill.status !== 'paid' && (
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              router.push(`/payments?bill=${bill.id}`);
                            }}
                            className="flex-1 px-3 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2"
                          >
                            <CreditCard className="w-4 h-4" />
                            Pay Now
                          </button>
                        )}
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            // Download PDF functionality
                          }}
                          className="px-3 py-2 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 transition-colors"
                        >
                          <Download className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}