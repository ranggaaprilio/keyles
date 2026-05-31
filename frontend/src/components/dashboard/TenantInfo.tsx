/**
 * Tenant Info Component — Dell 1996 retro style
 */

import { Building2, Mail, Calendar, CheckCircle, User } from 'lucide-react';

interface TenantInfoProps {
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
}

export function TenantInfo({ tenant, user }: TenantInfoProps) {
  const formatDate = (dateString: string | null) => {
    if (!dateString) return 'Not verified';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="bg-white">
      {/* Header — black banner */}
      <div className="bg-black px-4 py-4">
        <div className="flex items-center gap-3">
          <Building2 className="w-5 h-5 text-white" />
          <div>
            <h2 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase tracking-[1px] text-white">
              {tenant.organization_name}
            </h2>
            <p className="font-['Times_New_Roman',Times,serif] text-[11px] text-gray-400">
              Organization ID: {tenant.id.slice(0, 8)}...
            </p>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-4 space-y-4">
        {/* Tenant Details */}
        <div>
          <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black mb-3">
            Organization Details
          </h3>
          <div className="space-y-2">
            <div className="flex items-start gap-2">
              <Calendar className="w-4 h-4 text-gray-500 mt-0.5" />
              <div className="flex-1">
                <p className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold text-black">Created</p>
                <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-600">{formatDate(tenant.created_at)}</p>
              </div>
            </div>

            {tenant.verified_at && (
              <div className="flex items-start gap-2">
                <CheckCircle className="w-4 h-4 text-green-700 mt-0.5" />
                <div className="flex-1">
                  <p className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold text-black">Verified</p>
                  <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-600">{formatDate(tenant.verified_at)}</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Admin User Details */}
        <div className="pt-3 border-t border-black">
          <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black mb-3">
            Administrator Information
          </h3>
          <div className="space-y-2">
            <div className="flex items-start gap-2">
              <User className="w-4 h-4 text-gray-500 mt-0.5" />
              <div className="flex-1">
                <p className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold text-black">Full Name</p>
                <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-600">{user.full_name}</p>
              </div>
            </div>

            <div className="flex items-start gap-2">
              <Mail className="w-4 h-4 text-gray-500 mt-0.5" />
              <div className="flex-1">
                <p className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold text-black">Email</p>
                <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-600">{user.email}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Info Banner — Dell red CTA */}
        <div className="pt-3 border-t border-black">
          <div className="bg-[#e91d2a] p-3">
            <p className="font-['Times_New_Roman',Times,serif] text-sm text-[#fffff0]">
              <strong>Welcome to Keyles SSO!</strong> Your organization is successfully
              registered and verified. You can now manage users, configure authentication
              settings, and integrate with your applications.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
