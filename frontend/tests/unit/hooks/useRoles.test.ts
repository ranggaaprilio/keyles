/**
 * useRoles Hook Tests
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

// Mock API service
vi.mock('@/services/api/role', () => ({
  listUserRoles: vi.fn(),
  assignRole: vi.fn(),
  revokeRole: vi.fn(),
}));

import { listUserRoles, assignRole, revokeRole } from '@/services/api/role';
import { useUserRoles, useAssignRole, useRevokeRole } from '@/hooks/useRoles';
import type { RoleAssignment } from '@/types/user';

const mockRole: RoleAssignment = {
  id: 1,
  user_id: 'u1',
  client_id: 'c1',
  client_name: 'App One',
  tenant_id: 't1',
  role: 'admin',
  is_active: true,
  granted_at: '2024-01-01T00:00:00Z',
  granted_by: 'admin@test.com',
};

function createWrapper() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return React.createElement(QueryClientProvider, { client: qc }, children);
  };
}

describe('useUserRoles', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches roles for a user', async () => {
    vi.mocked(listUserRoles).mockResolvedValue([mockRole]);

    const { result } = renderHook(() => useUserRoles('u1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([mockRole]);
    expect(listUserRoles).toHaveBeenCalledWith('u1');
  });

  it('does not fetch when userId is empty', async () => {
    const { result } = renderHook(() => useUserRoles(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(listUserRoles).not.toHaveBeenCalled();
  });
});

describe('useAssignRole', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls assignRole API', async () => {
    vi.mocked(assignRole).mockResolvedValue(mockRole);

    const { result } = renderHook(() => useAssignRole(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync({
      userId: 'u1',
      req: { client_id: 'c1', role_name: 'admin' },
    });

    expect(assignRole).toHaveBeenCalledWith('u1', { client_id: 'c1', role_name: 'admin' });
  });
});

describe('useRevokeRole', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls revokeRole API', async () => {
    vi.mocked(revokeRole).mockResolvedValue(undefined);

    const { result } = renderHook(() => useRevokeRole(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync({ userId: 'u1', assignmentId: 1 });

    expect(revokeRole).toHaveBeenCalledWith('u1', 1);
  });
});
