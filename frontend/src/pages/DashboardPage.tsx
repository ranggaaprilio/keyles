/**
 * Dashboard Page Component
 */

import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { TenantDashboard } from '../components/dashboard/TenantDashboard';
import { Loader2 } from 'lucide-react';

export function DashboardPage() {
  const navigate = useNavigate();
  const { isAuthenticated, dashboardQuery, logout } = useAuth();

  // Redirect if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  // Handle logout
  const handleLogout = () => {
    logout();
  };

  // Loading state
  if (dashboardQuery.isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-12 h-12 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (dashboardQuery.isError) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
          <div className="text-red-600 mb-4">
            <svg className="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-gray-900 mb-2">
            Failed to Load Dashboard
          </h2>
          <p className="text-gray-600 mb-6">
            {dashboardQuery.error.message}
          </p>
          <button
            onClick={handleLogout}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Back to Login
          </button>
        </div>
      </div>
    );
  }

  // Success state
  return (
    <div className="min-h-screen bg-gray-50">
      {dashboardQuery.data && (
        <TenantDashboard
          tenant={dashboardQuery.data.tenant}
          user={dashboardQuery.data.user}
          onLogout={handleLogout}
        />
      )}
    </div>
  );
}
