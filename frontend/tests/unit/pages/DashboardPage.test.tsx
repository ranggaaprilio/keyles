/**
 * DashboardPage Unit Tests
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DashboardPage } from '../../../src/pages/DashboardPage';
import * as authApi from '../../../src/services/api/auth';

// Mock auth service
vi.mock('../../../src/services/api/auth', () => ({
  getDashboard: vi.fn(),
  isAuthenticated: vi.fn(),
  logout: vi.fn(),
}));

// Mock router navigation
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
    mutations: { retry: false },
  },
});

function renderDashboardPage() {
  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <DashboardPage />
      </BrowserRouter>
    </QueryClientProvider>
  );
}

const mockDashboardData = {
  tenant: {
    id: 'tenant-123',
    organization_name: 'Test Organization',
    status: 'active',
    created_at: '2024-01-15T10:00:00Z',
    verified_at: '2024-01-15T11:00:00Z',
  },
  user: {
    id: 'user-123',
    full_name: 'Admin User',
    email: 'admin@example.com',
    role: 'admin',
  },
};

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    queryClient.clear();
    (authApi.isAuthenticated as any).mockReturnValue(true);
  });

  it('redirects to login if not authenticated', () => {
    (authApi.isAuthenticated as any).mockReturnValue(false);
    
    renderDashboardPage();
    
    expect(mockNavigate).toHaveBeenCalledWith('/login', { replace: true });
  });

  it('shows loading state while fetching data', () => {
    (authApi.getDashboard as any).mockImplementation(() => new Promise(() => {})); // Never resolves
    
    renderDashboardPage();
    
    expect(screen.getByText(/Loading dashboard/i)).toBeInTheDocument();
  });

  it('displays dashboard data after successful load', async () => {
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    
    renderDashboardPage();
    
    await waitFor(() => {
      expect(screen.getByText('Test Organization')).toBeInTheDocument();
      expect(screen.getByText(/Welcome back, Admin User/i)).toBeInTheDocument();
      expect(screen.getByText('admin@example.com')).toBeInTheDocument();
    });
  });

  it('shows error message on load failure', async () => {
    (authApi.getDashboard as any).mockRejectedValue(new Error('Failed to load dashboard'));
    
    renderDashboardPage();
    
    await waitFor(() => {
      expect(screen.getByText(/Failed to Load Dashboard/i)).toBeInTheDocument();
    });
  });

  it('handles logout action', async () => {
    const user = userEvent.setup();
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    
    renderDashboardPage();
    
    await waitFor(() => {
      expect(screen.getByText('Test Organization')).toBeInTheDocument();
    });
    
    const logoutButton = screen.getByRole('button', { name: /Logout/i });
    await user.click(logoutButton);
    
    expect(authApi.logout).toHaveBeenCalled();
  });

  it('displays tenant status badge', async () => {
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    
    renderDashboardPage();
    
    await waitFor(() => {
      expect(screen.getByText(/active/i)).toBeInTheDocument();
    });
  });

  it('displays user role', async () => {
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    
    renderDashboardPage();
    
    await waitFor(() => {
      expect(screen.getByText(/ADMIN/i)).toBeInTheDocument();
    });
  });

  it('handles tenant with null verified_at', async () => {
    const dataWithNullVerified = {
      ...mockDashboardData,
      tenant: {
        ...mockDashboardData.tenant,
        verified_at: null,
      },
    };
    (authApi.getDashboard as any).mockResolvedValue(dataWithNullVerified);
    
    renderDashboardPage();
    
    await waitFor(() => {
      expect(screen.getByText('Test Organization')).toBeInTheDocument();
    });
  });

  it('formats dates correctly', async () => {
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    
    renderDashboardPage();
    
    await waitFor(() => {
      // Should display formatted date strings
      expect(screen.getByText(/January/i)).toBeInTheDocument();
    });
  });

  it('shows quick actions menu', async () => {
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    
    renderDashboardPage();
    
    await waitFor(() => {
      expect(screen.getByText(/Quick Actions/i)).toBeInTheDocument();
      expect(screen.getByText(/Manage Users/i)).toBeInTheDocument();
      expect(screen.getByText(/Settings/i)).toBeInTheDocument();
    });
  });
});
