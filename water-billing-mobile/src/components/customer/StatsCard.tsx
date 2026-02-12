'use client';

import { LucideIcon } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/Card';
import { cn } from '@/lib/utils/cn';

interface StatsCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  color: 'blue' | 'yellow' | 'green' | 'red' | 'purple' | 'orange';
  prefix?: string;
  suffix?: string;
}

const colorClasses = {
  blue: 'bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400',
  yellow: 'bg-yellow-100 dark:bg-yellow-900/50 text-yellow-600 dark:text-yellow-400',
  green: 'bg-green-100 dark:bg-green-900/50 text-green-600 dark:text-green-400',
  red: 'bg-red-100 dark:bg-red-900/50 text-red-600 dark:text-red-400',
  purple: 'bg-purple-100 dark:bg-purple-900/50 text-purple-600 dark:text-purple-400',
  orange: 'bg-orange-100 dark:bg-orange-900/50 text-orange-600 dark:text-orange-400',
};

export function StatsCard({ title, value, icon: Icon, trend, color, prefix, suffix }: StatsCardProps) {
  return (
    <Card className="hover:shadow-lg transition-shadow">
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="space-y-2">
            <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
              {title}
            </p>
            <div className="flex items-baseline gap-1">
              {prefix && (
                <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
                  {prefix}
                </span>
              )}
              <p className="text-2xl font-bold text-gray-900 dark:text-white">
                {typeof value === 'number' ? value.toLocaleString() : value}
              </p>
              {suffix && (
                <span className="text-sm font-medium text-gray-500 dark:text-gray-400 ml-1">
                  {suffix}
                </span>
              )}
            </div>
            {trend && (
              <div className="flex items-center gap-1">
                <span className={cn(
                  'text-xs font-medium',
                  trend.isPositive ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
                )}>
                  {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}%
                </span>
                <span className="text-xs text-gray-500 dark:text-gray-400">
                  vs last month
                </span>
              </div>
            )}
          </div>
          <div className={cn('p-3 rounded-xl', colorClasses[color])}>
            <Icon className="w-6 h-6" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}