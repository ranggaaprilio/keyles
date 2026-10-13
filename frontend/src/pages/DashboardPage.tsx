/**
 * Dashboard Page Component
 */

import { useAuth } from '../hooks/useAuth';
import { TenantInfo } from '../components/dashboard/TenantInfo';
import { PageHeader } from '../components/ui/page-header';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card';
import { Loader2 } from 'lucide-react';

export function DashboardPage() {
  const { dashboardQuery } = useAuth();

  if (dashboardQuery.isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="text-center">
          <Loader2 className="w-8 h-8 animate-spin text-muted-foreground mx-auto mb-3" />
          <p className="text-sm text-muted-foreground">
            Loading dashboard...
          </p>
        </div>
      </div>
    );
  }

  if (dashboardQuery.isError) {
    return (
      <div className="flex items-center justify-center py-20 px-4">
        <div className="max-w-md w-full rounded-lg border border-border bg-card p-6 shadow-card">
          <h2 className="text-lg font-medium text-foreground mb-2">
            Failed to Load Dashboard
          </h2>
          <p className="text-sm text-muted-foreground">
            {dashboardQuery.error.message}
          </p>
        </div>
      </div>
    );
  }

  if (!dashboardQuery.data) {
    return null;
  }

  const { tenant, user } = dashboardQuery.data;
  const isActive = tenant.status === 'active';

  return (
    <div className="mx-auto max-w-6xl px-4 sm:px-6 py-8">
      <PageHeader title="Welcome back" description={user.full_name} />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <TenantInfo tenant={tenant} user={user} />
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Account Status
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-2.5">
                <span
                  className={`h-2.5 w-2.5 rounded-full ${isActive ? 'bg-success' : 'bg-warning'}`}
                />
                <span className="text-lg font-medium capitalize text-foreground">
                  {tenant.status}
                </span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Your Role
              </CardTitle>
            </CardHeader>
            <CardContent>
              <span className="inline-flex items-center rounded-full bg-primary px-3 py-1 text-sm font-medium capitalize text-primary-foreground">
                {user.role}
              </span>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
