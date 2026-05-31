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
    <div className="min-h-screen bg-gray-100 font-['Times_New_Roman',Times,serif]">
      {/* Header - Black bar */}
      <header className="bg-black border-b border-black">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 border border-white bg-white flex items-center justify-center">
                <LayoutDashboard className="w-6 h-6 text-black" />
              </div>
              <div>
                <h1 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-white text-lg">
                  {tenant.organization_name}
                </h1>
                <p className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-gray-400">
                  Keyles Multi-Tenant SSO
                </p>
              </div>
            </div>
            <button
              onClick={onLogout}
              className="flex items-center gap-2 px-4 py-2 border border-white text-white hover:bg-white hover:text-black transition-colors font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1.5px]"
            >
              <LogOut className="w-4 h-4" />
              <span>Logout</span>
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Welcome Section */}
        <div className="mb-8">
          <h2 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-2xl text-black mb-2">
            Welcome back, {user.full_name}!
          </h2>
          <p className="text-sm text-gray-600">
            Here&apos;s an overview of your organization
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
            {/* Status Card - Ribbon style */}
            <div className="bg-white border border-black shadow-[2px_2px_0_#000]">
              <div className="bg-white border-b border-black px-4 py-2">
                <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-gray-500">
                  Account Status
                </h3>
              </div>
              <div className="p-6">
                <div className="flex items-center gap-2">
                  <div className={`w-3 h-3 border border-black ${
                    tenant.status === 'active' ? 'bg-green-700' : 'bg-yellow-600'
                  }`} />
                  <span className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-lg capitalize">
                    {tenant.status}
                  </span>
                </div>
                {tenant.status === 'active' && (
                  <p className="mt-2 text-xs text-gray-500">
                    Your account is fully activated
                  </p>
                )}
              </div>
            </div>

            {/* User Role Card - Ribbon style */}
            <div className="bg-white border border-black shadow-[2px_2px_0_#000]">
              <div className="bg-white border-b border-black px-4 py-2">
                <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-gray-500">
                  Your Role
                </h3>
              </div>
              <div className="p-6 bg-[#8c9ae0]/15">
                <span className="inline-flex items-center px-3 py-1 border border-black bg-[#8c9ae0] text-white text-sm font-bold font-[Helvetica,Arial,system-ui,sans-serif]">
                  {user.role.toUpperCase()}
                </span>
                <p className="mt-2 text-xs text-gray-500">
                  Administrator access
                </p>
              </div>
            </div>

            {/* Quick Actions - Ribbon style */}
            <div className="bg-white border border-black shadow-[2px_2px_0_#000]">
              <div className="bg-white border-b border-black px-4 py-2">
                <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-gray-500">
                  Quick Actions
                </h3>
              </div>
              <div className="p-4">
                <div className="space-y-1">
                  <button className="w-full px-4 py-2 text-sm text-left text-black hover:bg-gray-100 border border-transparent hover:border-black transition-colors font-['Times_New_Roman',Times,serif]">
                    Manage Users
                  </button>
                  <button className="w-full px-4 py-2 text-sm text-left text-black hover:bg-gray-100 border border-transparent hover:border-black transition-colors font-['Times_New_Roman',Times,serif]">
                    Settings
                  </button>
                  <button className="w-full px-4 py-2 text-sm text-left text-black hover:bg-gray-100 border border-transparent hover:border-black transition-colors font-['Times_New_Roman',Times,serif]">
                    API Keys
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}