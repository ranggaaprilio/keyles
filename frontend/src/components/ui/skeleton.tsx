/**
 * Loading Skeleton Components
 */

import React from 'react';

/**
 * Generic Skeleton primitive used for loading placeholders.
 */
export function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={`animate-pulse rounded-md bg-gray-200 ${className ?? ''}`}
      {...props}
    />
  );
}

export function FormSkeleton() {
  return (
    <div className="space-y-4 animate-pulse">
      <div className="space-y-2">
        <div className="h-4 bg-gray-200 rounded w-24"></div>
        <div className="h-10 bg-gray-200 rounded"></div>
      </div>
      <div className="space-y-2">
        <div className="h-4 bg-gray-200 rounded w-32"></div>
        <div className="h-10 bg-gray-200 rounded"></div>
      </div>
      <div className="space-y-2">
        <div className="h-4 bg-gray-200 rounded w-28"></div>
        <div className="h-10 bg-gray-200 rounded"></div>
      </div>
      <div className="h-10 bg-gray-200 rounded mt-6"></div>
    </div>
  );
}

export function CardSkeleton() {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 animate-pulse">
      <div className="space-y-4">
        <div className="h-6 bg-gray-200 rounded w-2/3"></div>
        <div className="space-y-2">
          <div className="h-4 bg-gray-200 rounded"></div>
          <div className="h-4 bg-gray-200 rounded w-5/6"></div>
          <div className="h-4 bg-gray-200 rounded w-4/6"></div>
        </div>
      </div>
    </div>
  );
}

export function DashboardSkeleton() {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header Skeleton */}
      <div className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between animate-pulse">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gray-200 rounded-lg"></div>
              <div className="space-y-2">
                <div className="h-6 bg-gray-200 rounded w-32"></div>
                <div className="h-4 bg-gray-200 rounded w-24"></div>
              </div>
            </div>
            <div className="h-10 bg-gray-200 rounded w-24"></div>
          </div>
        </div>
      </div>

      {/* Main Content Skeleton */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="animate-pulse space-y-6">
          {/* Welcome Section */}
          <div className="space-y-2">
            <div className="h-8 bg-gray-200 rounded w-64"></div>
            <div className="h-4 bg-gray-200 rounded w-48"></div>
          </div>

          {/* Dashboard Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2">
              <CardSkeleton />
            </div>
            <div className="space-y-6">
              <CardSkeleton />
              <CardSkeleton />
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="animate-pulse">
      <div className="overflow-hidden border border-gray-200 rounded-lg">
        {/* Table Header */}
        <div className="bg-gray-50 border-b border-gray-200 px-6 py-3">
          <div className="flex gap-4">
            <div className="h-4 bg-gray-200 rounded w-1/4"></div>
            <div className="h-4 bg-gray-200 rounded w-1/4"></div>
            <div className="h-4 bg-gray-200 rounded w-1/4"></div>
            <div className="h-4 bg-gray-200 rounded w-1/4"></div>
          </div>
        </div>
        
        {/* Table Rows */}
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className="bg-white border-b border-gray-200 px-6 py-4">
            <div className="flex gap-4">
              <div className="h-4 bg-gray-200 rounded w-1/4"></div>
              <div className="h-4 bg-gray-200 rounded w-1/4"></div>
              <div className="h-4 bg-gray-200 rounded w-1/4"></div>
              <div className="h-4 bg-gray-200 rounded w-1/4"></div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
