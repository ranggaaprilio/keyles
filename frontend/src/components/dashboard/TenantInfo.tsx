/**
 * Tenant Info Component
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
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
      {/* Header */}
      <div className="bg-gradient-to-r from-blue-600 to-blue-700 px-6 py-8">
        <div className="flex items-center gap-4">
          <div className="w-16 h-16 bg-white/20 rounded-xl flex items-center justify-center">
            <Building2 className="w-8 h-8 text-white" />
          </div>
          <div>
            <h2 className="text-2xl font-bold text-white mb-1">
              {tenant.organization_name}
            </h2>
            <p className="text-blue-100 text-sm">
              Organization ID: {tenant.id.slice(0, 8)}...
            </p>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-6 space-y-6">
        {/* Tenant Details */}
        <div>
          <h3 className="text-sm font-medium text-gray-500 mb-4">
            Organization Details
          </h3>
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <Calendar className="w-5 h-5 text-gray-400 mt-0.5" />
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-700">Created</p>
                <p className="text-sm text-gray-600">{formatDate(tenant.created_at)}</p>
              </div>
            </div>

            {tenant.verified_at && (
              <div className="flex items-start gap-3">
                <CheckCircle className="w-5 h-5 text-green-500 mt-0.5" />
                <div className="flex-1">
                  <p className="text-sm font-medium text-gray-700">Verified</p>
                  <p className="text-sm text-gray-600">{formatDate(tenant.verified_at)}</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Admin User Details */}
        <div className="pt-6 border-t border-gray-200">
          <h3 className="text-sm font-medium text-gray-500 mb-4">
            Administrator Information
          </h3>
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <User className="w-5 h-5 text-gray-400 mt-0.5" />
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-700">Full Name</p>
                <p className="text-sm text-gray-600">{user.full_name}</p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <Mail className="w-5 h-5 text-gray-400 mt-0.5" />
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-700">Email</p>
                <p className="text-sm text-gray-600">{user.email}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Info Banner */}
        <div className="pt-6 border-t border-gray-200">
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <p className="text-sm text-blue-800">
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
