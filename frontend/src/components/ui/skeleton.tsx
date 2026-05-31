/**
 * Loading Skeleton Components
 */

import React from "react";

/**
 * Generic Skeleton primitive used for loading placeholders.
 */
export function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={`animate-pulse bg-gray-200 ${className ?? ""}`}
      {...props}
    />
  );
}

export function FormSkeleton() {
  return (
    <div className="space-y-4 animate-pulse">
      <div className="space-y-2">
        <div className="h-4 w-24 bg-gray-200"></div>
        <div className="h-10 bg-gray-200 border border-black"></div>
      </div>
      <div className="space-y-2">
        <div className="h-4 w-32 bg-gray-200"></div>
        <div className="h-10 bg-gray-200 border border-black"></div>
      </div>
      <div className="space-y-2">
        <div className="h-4 w-28 bg-gray-200"></div>
        <div className="h-10 bg-gray-200 border border-black"></div>
      </div>
      <div className="h-10 bg-gray-200 border border-black mt-6"></div>
    </div>
  );
}

export function CardSkeleton() {
  return (
    <div className="bg-white border border-black p-4 animate-pulse">
      <div className="h-4 w-3/4 bg-gray-200 mb-3"></div>
      <div className="space-y-2">
        <div className="h-3 bg-gray-200"></div>
        <div className="h-3 w-5/6 bg-gray-200"></div>
        <div className="h-3 w-2/3 bg-gray-200"></div>
      </div>
    </div>
  );
}

export function DashboardSkeleton() {
  return (
    <div className="min-h-screen bg-white">
      <div className="flex">
        <div className="w-64 border-r border-black p-4 space-y-4">
          <div className="h-6 w-24 bg-gray-200"></div>
          <div className="space-y-2">
            <div className="h-8 bg-gray-200"></div>
            <div className="h-8 bg-gray-200"></div>
            <div className="h-8 bg-gray-200"></div>
          </div>
        </div>
        <div className="flex-1 p-6 space-y-6">
          <div className="h-8 w-48 bg-gray-200"></div>
          <div className="grid grid-cols-3 gap-4">
            <CardSkeleton />
            <CardSkeleton />
            <CardSkeleton />
          </div>
        </div>
      </div>
    </div>
  );
}

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="animate-pulse">
      <div className="border border-black">
        <div className="h-10 bg-gray-100 border-b border-black flex">
          <div className="flex-1 px-4 py-2"><div className="h-4 w-20 bg-gray-200"></div></div>
          <div className="flex-1 px-4 py-2"><div className="h-4 w-24 bg-gray-200"></div></div>
          <div className="flex-1 px-4 py-2"><div className="h-4 w-16 bg-gray-200"></div></div>
          <div className="flex-1 px-4 py-2"><div className="h-4 w-20 bg-gray-200"></div></div>
        </div>
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className="h-12 border-b border-black last:border-b-0 flex">
            <div className="flex-1 px-4 py-3"><div className="h-4 w-32 bg-gray-200"></div></div>
            <div className="flex-1 px-4 py-3"><div className="h-4 w-40 bg-gray-200"></div></div>
            <div className="flex-1 px-4 py-3"><div className="h-4 w-16 bg-gray-200"></div></div>
            <div className="flex-1 px-4 py-3"><div className="h-4 w-24 bg-gray-200"></div></div>
          </div>
        ))}
      </div>
    </div>
  );
}
