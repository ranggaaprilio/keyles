/**
 * Dashboard Page Component
 */

import { useAuth } from '../hooks/useAuth';
import { TenantInfo } from '../components/dashboard/TenantInfo';
import { Loader2 } from 'lucide-react';

export function DashboardPage() {
  const { dashboardQuery } = useAuth();

  if (dashboardQuery.isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="text-center">
          <Loader2 className="w-12 h-12 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (dashboardQuery.isError) {
    return (
      <div className="flex items-center justify-center py-20 px-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
          <h2 className="text-xl font-bold text-gray-900 mb-2">
            Failed to Load Dashboard
          </h2>
          <p className="text-gray-600">
            {dashboardQuery.error.message}
          </p>
        </div>
      </div>
    );
  }

  const { tenant, user } = dashboardQuery.data!;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Welcome back, {user.full_name}!
        </h2>
        <p className="text-gray-600">
          Here's an overview of your organization
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <TenantInfo tenant={tenant} user={user} />
        </div>

        <div className="space-y-6">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-3">
              Account Status
            </h3>
            <div className="flex items-center gap-2">
              <div className={`w-3 h-3 rounded-full ${
                tenant.status === 'active' ? 'bg-green-500' : 'bg-yellow-500'
              }`} />
              <span className="text-lg font-semibold text-gray-900 capitalize">
                {tenant.status}
              </span>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-3">
              Your Role
            </h3>
            <div className="inline-flex items-center px-3 py-1 rounded-full bg-blue-100 text-blue-800 text-sm font-medium">
              {user.role.toUpperCase()}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
