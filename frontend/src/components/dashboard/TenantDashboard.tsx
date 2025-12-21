/**
 * Tenant Dashboard Component
 */

import { TenantInfo } from './TenantInfo';
import { LogOut, LayoutDashboard } from 'lucide-react';

interface TenantDashboardProps {
  tenant: {
    id: string;
    organization_name: string;
    status: string;
    created_at: string;
    verified_at: string | null;
  };
  user: {
    id: string;
    full_name: string;
    email: string;
    role: string;
  };
  onLogout: () => void;
}

export function TenantDashboard({ tenant, user, onLogout }: TenantDashboardProps) {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center">
                <LayoutDashboard className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900">
                  {tenant.organization_name}
                </h1>
                <p className="text-sm text-gray-500">
                  Keyles Multi-Tenant SSO
                </p>
              </div>
            </div>
            <button
              onClick={onLogout}
              className="flex items-center gap-2 px-4 py-2 text-gray-700 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <LogOut className="w-4 h-4" />
              <span className="text-sm font-medium">Logout</span>
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Welcome Section */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-2">
            Welcome back, {user.full_name}!
          </h2>
          <p className="text-gray-600">
            Here's an overview of your organization
          </p>
        </div>

        {/* Dashboard Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Tenant Information */}
          <div className="lg:col-span-2">
            <TenantInfo tenant={tenant} user={user} />
          </div>

          {/* Quick Stats */}
          <div className="space-y-6">
            {/* Status Card */}
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
              {tenant.status === 'active' && (
                <p className="mt-2 text-xs text-gray-500">
                  Your account is fully activated
                </p>
              )}
            </div>

            {/* User Role Card */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h3 className="text-sm font-medium text-gray-500 mb-3">
                Your Role
              </h3>
              <div className="inline-flex items-center px-3 py-1 rounded-full bg-blue-100 text-blue-800 text-sm font-medium">
                {user.role.toUpperCase()}
              </div>
              <p className="mt-2 text-xs text-gray-500">
                Administrator access
              </p>
            </div>

            {/* Quick Actions */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h3 className="text-sm font-medium text-gray-500 mb-4">
                Quick Actions
              </h3>
              <div className="space-y-2">
                <button className="w-full px-4 py-2 text-sm text-left text-gray-700 hover:bg-gray-50 rounded-lg transition-colors">
                  Manage Users
                </button>
                <button className="w-full px-4 py-2 text-sm text-left text-gray-700 hover:bg-gray-50 rounded-lg transition-colors">
                  Settings
                </button>
                <button className="w-full px-4 py-2 text-sm text-left text-gray-700 hover:bg-gray-50 rounded-lg transition-colors">
                  API Keys
                </button>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
