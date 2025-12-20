/**
 * useAuth Hook Unit Tests
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';
import { useAuth } from '../../../src/hooks/useAuth';
import * as authApi from '../../../src/services/api/auth';

// Mock auth service
vi.mock('../../../src/services/api/auth', () => ({
  login: vi.fn(),
  getDashboard: vi.fn(),
  logout: vi.fn(),
  isAuthenticated: vi.fn(),
  storeAuthData: vi.fn(),
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

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>{children}</BrowserRouter>
    </QueryClientProvider>
  );
}

describe('useAuth', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (authApi.isAuthenticated as any).mockReturnValue(false);
  });

  it('returns initial auth state', () => {
    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.user).toBeNull();
    expect(result.current.isLoggingIn).toBe(false);
  });

  it('returns authenticated state when token exists', () => {
    (authApi.isAuthenticated as any).mockReturnValue(true);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    expect(result.current.isAuthenticated).toBe(true);
  });

  it('handles successful login', async () => {
    const mockLoginResponse = {
      token: 'test-jwt-token',
      expires_in: 86400,
      user: {
        id: 'user-123',
        email: 'admin@example.com',
        full_name: 'Admin User',
        role: 'admin',
      },
      tenant: {
        id: 'tenant-123',
        organization_name: 'Test Org',
        status: 'active',
      },
    };

    (authApi.login as any).mockResolvedValue(mockLoginResponse);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    result.current.login({
      email: 'admin@example.com',
      password: 'SecurePassword123!',
    });

    await waitFor(() => {
      expect(result.current.isLoggingIn).toBe(false);
    });

    expect(authApi.login).toHaveBeenCalledWith('admin@example.com', 'SecurePassword123!');
    expect(authApi.storeAuthData).toHaveBeenCalledWith(
      mockLoginResponse.token,
      mockLoginResponse.user
    );
  });

  it('handles login error', async () => {
    const error = new Error('Invalid credentials');
    (authApi.login as any).mockRejectedValue(error);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    result.current.login({
      email: 'admin@example.com',
      password: 'WrongPassword',
    });

    await waitFor(() => {
      expect(result.current.isLoggingIn).toBe(false);
    });

    expect(authApi.login).toHaveBeenCalled();
    expect(authApi.storeAuthData).not.toHaveBeenCalled();
  });

  it('fetches dashboard data when authenticated', async () => {
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

    (authApi.isAuthenticated as any).mockReturnValue(true);
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.dashboardQuery.isSuccess).toBe(true);
    });

    expect(authApi.getDashboard).toHaveBeenCalled();
    expect(result.current.dashboardQuery.data).toEqual(mockDashboardData);
  });

  it('does not fetch dashboard when not authenticated', () => {
    (authApi.isAuthenticated as any).mockReturnValue(false);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    expect(result.current.dashboardQuery.isLoading).toBe(false);
    expect(authApi.getDashboard).not.toHaveBeenCalled();
  });

  it('handles logout', () => {
    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    result.current.logout();

    expect(authApi.logout).toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/login');
  });

  it('clears dashboard data on logout', async () => {
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

    (authApi.isAuthenticated as any).mockReturnValue(true);
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.dashboardQuery.isSuccess).toBe(true);
    });

    (authApi.isAuthenticated as any).mockReturnValue(false);
    result.current.logout();

    expect(authApi.logout).toHaveBeenCalled();
  });
});
