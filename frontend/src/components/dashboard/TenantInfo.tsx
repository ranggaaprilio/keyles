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

function InfoRow({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-start gap-3">
      <span className="mt-0.5 text-muted-foreground">{icon}</span>
      <div className="flex-1 min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className="text-sm text-foreground break-words">{value}</p>
      </div>
    </div>
  );
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
    <div className="rounded-lg border border-border bg-card shadow-card">
      {/* Header */}
      <div className="flex items-center gap-3 border-b border-border px-6 py-5">
        <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-muted text-foreground">
          <Building2 className="w-5 h-5" />
        </span>
        <div className="min-w-0">
          <h2 className="text-lg font-medium leading-tight tracking-tight text-foreground truncate">
            {tenant.organization_name}
          </h2>
          <p className="text-xs text-muted-foreground">
            Organization ID: {tenant.id.slice(0, 8)}...
          </p>
        </div>
      </div>

      {/* Content */}
      <div className="p-6 space-y-6">
        {/* Tenant Details */}
        <div>
          <h3 className="text-sm font-medium text-foreground mb-3">
            Organization Details
          </h3>
          <div className="space-y-3">
            <InfoRow
              icon={<Calendar className="w-4 h-4" />}
              label="Created"
              value={formatDate(tenant.created_at)}
            />
            {tenant.verified_at && (
              <InfoRow
                icon={<CheckCircle className="w-4 h-4 text-success" />}
                label="Verified"
                value={formatDate(tenant.verified_at)}
              />
            )}
          </div>
        </div>

        {/* Admin User Details */}
        <div className="pt-6 border-t border-border">
          <h3 className="text-sm font-medium text-foreground mb-3">
            Administrator Information
          </h3>
          <div className="space-y-3">
            <InfoRow icon={<User className="w-4 h-4" />} label="Full Name" value={user.full_name} />
            <InfoRow icon={<Mail className="w-4 h-4" />} label="Email" value={user.email} />
          </div>
        </div>

        {/* Info Banner */}
        <div className="rounded-md bg-muted p-4">
          <p className="text-sm text-muted-foreground">
            <strong className="font-medium text-foreground">Welcome to Keyles SSO!</strong> Your organization is successfully
            registered and verified. You can now manage users, configure authentication
            settings, and integrate with your applications.
          </p>
        </div>
      </div>
    </div>
  );
}
