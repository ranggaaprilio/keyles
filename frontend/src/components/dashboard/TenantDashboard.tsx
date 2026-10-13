/**
 * Tenant Dashboard Component
 */

import { TenantInfo } from './TenantInfo';
import { LogOut, LayoutDashboard } from 'lucide-react';
import { Button } from '../ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';

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
  const isActive = tenant.status === 'active';

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border bg-background/90 backdrop-blur">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground">
                <LayoutDashboard className="w-5 h-5" />
              </span>
              <div>
                <h1 className="text-lg font-medium tracking-tight text-foreground">
                  {tenant.organization_name}
                </h1>
                <p className="text-xs text-muted-foreground">
                  Keyles Multi-Tenant SSO
                </p>
              </div>
            </div>
            <Button variant="outline" onClick={onLogout}>
              <LogOut className="w-4 h-4" />
              <span>Logout</span>
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Welcome Section */}
        <div className="mb-8">
          <h2 className="text-2xl font-medium leading-tight tracking-tight text-foreground mb-1">
            Welcome back, {user.full_name}!
          </h2>
          <p className="text-sm text-muted-foreground">
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
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Account Status
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-2.5">
                  <span className={`h-2.5 w-2.5 rounded-full ${isActive ? 'bg-success' : 'bg-warning'}`} />
                  <span className="text-lg font-medium capitalize text-foreground">
                    {tenant.status}
                  </span>
                </div>
                {isActive && (
                  <p className="mt-2 text-xs text-muted-foreground">
                    Your account is fully activated
                  </p>
                )}
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
                <p className="mt-2 text-xs text-muted-foreground">
                  Administrator access
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Quick Actions
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-1">
                  {['Manage Users', 'Settings', 'API Keys'].map((label) => (
                    <button
                      key={label}
                      className="w-full rounded-md px-3 py-2 text-sm text-left text-foreground hover:bg-muted transition-colors duration-fast"
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  );
}
