'use client';

import { useState } from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { Card, CardContent, CardHeader } from '@/components/ui/Card';
import { Calendar, Droplets, TrendingDown, TrendingUp } from 'lucide-react';
import { MeterReading } from '@/types/billing';
import { cn } from '@/lib/utils/cn';

interface ConsumptionChartProps {
  data: MeterReading[];
  isLoading?: boolean;
}

export function ConsumptionChart({ data, isLoading }: ConsumptionChartProps) {
  const [timeRange, setTimeRange] = useState<'6' | '12' | '24'>('12');

  // Prepare chart data
  const chartData = data
    .slice(0, parseInt(timeRange))
    .reverse()
    .map(reading => ({
      period: reading.billing_period,
      consumption: Number(reading.consumption.toFixed(1)),
      amount: reading.total_amount,
      month: new Date(reading.reading_date).toLocaleString('default', { month: 'short' }),
      year: new Date(reading.reading_date).getFullYear(),
    }));

  // Calculate statistics
  const averageConsumption = chartData.length > 0
    ? chartData.reduce((acc, curr) => acc + curr.consumption, 0) / chartData.length
    : 0;
  
  const totalConsumption = chartData.reduce((acc, curr) => acc + curr.consumption, 0);
  const totalAmount = chartData.reduce((acc, curr) => acc + curr.amount, 0);

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-white dark:bg-gray-900 p-4 rounded-lg shadow-lg border border-gray-200 dark:border-gray-800">
          <p className="text-sm font-medium text-gray-900 dark:text-white mb-2">
            {label}
          </p>
          <div className="space-y-1">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Consumption: <span className="font-semibold text-blue-600 dark:text-blue-400">
                {payload[0].value} m³
              </span>
            </p>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Amount: <span className="font-semibold text-green-600 dark:text-green-400">
                KSh {payload[1].value.toLocaleString()}
              </span>
            </p>
          </div>
        </div>
      );
    }
    return null;
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <div className="h-6 w-48 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
        </CardHeader>
        <CardContent>
          <div className="h-80 w-full bg-gray-100 dark:bg-gray-800 rounded animate-pulse" />
        </CardContent>
      </Card>
    );
  }

  if (!data || data.length === 0) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Droplets className="w-5 h-5 text-blue-600 dark:text-blue-400" />
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Water Consumption History
            </h3>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-80 flex flex-col items-center justify-center text-center">
            <Droplets className="w-12 h-12 text-gray-400 mb-4" />
            <p className="text-gray-600 dark:text-gray-400 font-medium">
              No consumption data available
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-500 mt-1">
              Meter readings will appear here once submitted
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-2">
          <Droplets className="w-5 h-5 text-blue-600 dark:text-blue-400" />
          <div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Water Consumption History
            </h3>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Last {timeRange} months
            </p>
          </div>
        </div>
        
        <div className="flex items-center gap-4">
          {/* Stats Summary */}
          <div className="hidden md:flex items-center gap-4 text-sm">
            <div className="flex items-center gap-1">
              <TrendingDown className="w-4 h-4 text-blue-500" />
              <span className="text-gray-600 dark:text-gray-400">Avg:</span>
              <span className="font-medium text-gray-900 dark:text-white">
                {averageConsumption.toFixed(1)} m³
              </span>
            </div>
            <div className="flex items-center gap-1">
              <TrendingUp className="w-4 h-4 text-green-500" />
              <span className="text-gray-600 dark:text-gray-400">Total:</span>
              <span className="font-medium text-gray-900 dark:text-white">
                KSh {totalAmount.toLocaleString()}
              </span>
            </div>
          </div>

          {/* Time Range Selector */}
          <div className="flex items-center gap-2">
            <Calendar className="w-4 h-4 text-gray-500 dark:text-gray-400" />
            <select
              value={timeRange}
              onChange={(e) => setTimeRange(e.target.value as any)}
              className="text-sm bg-transparent border-0 text-gray-700 dark:text-gray-300 focus:ring-0 cursor-pointer pr-6"
              aria-label="Select time range"
            >
              <option value="6">Last 6 months</option>
              <option value="12">Last 12 months</option>
              <option value="24">Last 24 months</option>
            </select>
          </div>
        </div>
      </CardHeader>

      <CardContent>
        <div className="h-80 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart
              data={chartData}
              margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
            >
              <defs>
                <linearGradient id="consumptionGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="amountGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                </linearGradient>
              </defs>
              
              <CartesianGrid 
                strokeDasharray="3 3" 
                className="stroke-gray-200 dark:stroke-gray-700"
                vertical={false}
              />
              
              <XAxis 
                dataKey="period" 
                className="text-xs text-gray-600 dark:text-gray-400"
                tickMargin={10}
                axisLine={false}
                tickLine={false}
              />
              
              <YAxis 
                yAxisId="left"
                className="text-xs text-gray-600 dark:text-gray-400"
                tickMargin={10}
                axisLine={false}
                tickLine={false}
                width={60}
                label={{
                  value: 'Consumption (m³)',
                  angle: -90,
                  position: 'insideLeft',
                  style: { 
                    textAnchor: 'middle',
                    fill: '#6b7280',
                    fontSize: '12px',
                  },
                }}
              />
              
              <YAxis 
                yAxisId="right"
                orientation="right"
                className="text-xs text-gray-600 dark:text-gray-400"
                tickMargin={10}
                axisLine={false}
                tickLine={false}
                width={80}
                tickFormatter={(value) => `KSh ${(value / 1000).toFixed(0)}k`}
                label={{
                  value: 'Amount (KSh)',
                  angle: 90,
                  position: 'insideRight',
                  style: { 
                    textAnchor: 'middle',
                    fill: '#6b7280',
                    fontSize: '12px',
                  },
                }}
              />
              
              <Tooltip content={<CustomTooltip />} />
              
              <Legend 
                verticalAlign="top"
                height={36}
                iconType="circle"
                iconSize={8}
                formatter={(value) => (
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    {value}
                  </span>
                )}
              />
              
              <Area
                yAxisId="left"
                type="monotone"
                dataKey="consumption"
                name="Consumption"
                stroke="#3b82f6"
                strokeWidth={2}
                fill="url(#consumptionGradient)"
                dot={{ r: 3, fill: '#3b82f6', strokeWidth: 0 }}
                activeDot={{ r: 5, stroke: '#3b82f6', strokeWidth: 2, fill: 'white' }}
              />
              
              <Area
                yAxisId="right"
                type="monotone"
                dataKey="amount"
                name="Amount"
                stroke="#10b981"
                strokeWidth={2}
                fill="url(#amountGradient)"
                dot={{ r: 3, fill: '#10b981', strokeWidth: 0 }}
                activeDot={{ r: 5, stroke: '#10b981', strokeWidth: 2, fill: 'white' }}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Mobile Stats Summary */}
        <div className="mt-4 grid grid-cols-2 gap-4 md:hidden">
          <div className="p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
            <p className="text-xs text-gray-500 dark:text-gray-400">Average Consumption</p>
            <p className="text-lg font-semibold text-gray-900 dark:text-white">
              {averageConsumption.toFixed(1)} m³
            </p>
          </div>
          <div className="p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
            <p className="text-xs text-gray-500 dark:text-gray-400">Total Spent</p>
            <p className="text-lg font-semibold text-gray-900 dark:text-white">
              KSh {totalAmount.toLocaleString()}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}