/**
 * Dashboard Page Component — Dell 1996 retro style
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
          <Loader2 className="w-8 h-8 animate-spin text-black mx-auto mb-3" />
          <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-600">
            Loading dashboard...
          </p>
        </div>
      </div>
    );
  }

  if (dashboardQuery.isError) {
    return (
      <div className="flex items-center justify-center py-20 px-4">
        <div className="max-w-md w-full border border-black bg-white p-4">
          <h2 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase tracking-[1.5px] text-black mb-2">
            Failed to Load Dashboard
          </h2>
          <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-600">
            {dashboardQuery.error.message}
          </p>
        </div>
      </div>
    );
  }

  const { tenant, user } = dashboardQuery.data!;

  return (
    <div className="max-w-5xl mx-auto px-4 py-6">
      {/* Section eyebrow — olive */}
      <div className="bg-[#8e8a25] px-4 py-4 mb-0">
        <h2 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[28px] font-black uppercase leading-[1.0] text-black">
          WELCOME BACK
        </h2>
        <p className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-black">
          {user.full_name}
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-0">
        <div className="lg:col-span-2 border-x border-b border-black">
          <TenantInfo tenant={tenant} user={user} />
        </div>

        <div>
          {/* Status ribbon card — sage */}
          <div className="border-x border-b border-black lg:border-l-0">
            <div className="border-b border-black bg-white px-3 py-1.5">
              <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
                Account Status
              </h3>
            </div>
            <div className="bg-[#b3bd95] px-4 py-3">
              <div className="flex items-center gap-2">
                <div className={`w-3 h-3 ${
                  tenant.status === 'active' ? 'bg-green-700' : 'bg-yellow-600'
                }`} />
                <span className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                  {tenant.status}
                </span>
              </div>
            </div>
          </div>

          {/* Role ribbon card — periwinkle */}
          <div className="border-x border-b border-black lg:border-l-0">
            <div className="border-b border-black bg-white px-3 py-1.5">
              <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
                Your Role
              </h3>
            </div>
            <div className="bg-[#8c9ae0] px-4 py-3">
              <span className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                {user.role.toUpperCase()}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
