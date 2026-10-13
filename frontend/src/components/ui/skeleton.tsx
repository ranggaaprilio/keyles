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
      className={`animate-pulse rounded-md bg-muted ${className ?? ""}`}
      {...props}
    />
  );
}

export function FormSkeleton() {
  return (
    <div className="space-y-4 animate-pulse">
      <div className="space-y-2">
        <div className="h-4 w-24 rounded-md bg-muted"></div>
        <div className="h-10 rounded-md bg-muted"></div>
      </div>
      <div className="space-y-2">
        <div className="h-4 w-32 rounded-md bg-muted"></div>
        <div className="h-10 rounded-md bg-muted"></div>
      </div>
      <div className="space-y-2">
        <div className="h-4 w-28 rounded-md bg-muted"></div>
        <div className="h-10 rounded-md bg-muted"></div>
      </div>
      <div className="h-10 rounded-md bg-muted mt-6"></div>
    </div>
  );
}

export function CardSkeleton() {
  return (
    <div className="rounded-lg border border-border bg-card p-4 animate-pulse">
      <div className="h-4 w-3/4 rounded-md bg-muted mb-3"></div>
      <div className="space-y-2">
        <div className="h-3 rounded-md bg-muted"></div>
        <div className="h-3 w-5/6 rounded-md bg-muted"></div>
        <div className="h-3 w-2/3 rounded-md bg-muted"></div>
      </div>
    </div>
  );
}

export function DashboardSkeleton() {
  return (
    <div className="min-h-screen bg-background">
      <div className="flex">
        <div className="w-64 border-r border-border bg-muted p-4 space-y-4">
          <div className="h-6 w-24 rounded-md bg-muted"></div>
          <div className="space-y-2">
            <div className="h-8 rounded-md bg-muted"></div>
            <div className="h-8 rounded-md bg-muted"></div>
            <div className="h-8 rounded-md bg-muted"></div>
          </div>
        </div>
        <div className="flex-1 p-6 space-y-6">
          <div className="h-8 w-48 rounded-md bg-muted"></div>
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
      <div className="rounded-lg border border-border overflow-hidden">
        <div className="h-10 bg-muted border-b border-border flex">
          <div className="flex-1 px-4 py-2"><div className="h-4 w-20 rounded-md bg-muted"></div></div>
          <div className="flex-1 px-4 py-2"><div className="h-4 w-24 rounded-md bg-muted"></div></div>
          <div className="flex-1 px-4 py-2"><div className="h-4 w-16 rounded-md bg-muted"></div></div>
          <div className="flex-1 px-4 py-2"><div className="h-4 w-20 rounded-md bg-muted"></div></div>
        </div>
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className="h-12 border-b border-border last:border-b-0 flex">
            <div className="flex-1 px-4 py-3"><div className="h-4 w-32 rounded-md bg-muted"></div></div>
            <div className="flex-1 px-4 py-3"><div className="h-4 w-40 rounded-md bg-muted"></div></div>
            <div className="flex-1 px-4 py-3"><div className="h-4 w-16 rounded-md bg-muted"></div></div>
            <div className="flex-1 px-4 py-3"><div className="h-4 w-24 rounded-md bg-muted"></div></div>
          </div>
        ))}
      </div>
    </div>
  );
}
