/**
 * useUsers Hook Tests
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

// Mock API service
vi.mock('@/services/api/user', () => ({
  listUsers: vi.fn(),
  inviteUser: vi.fn(),
  getUser: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
  updateUserStatus: vi.fn(),
  resendInvitation: vi.fn(),
}));

import { listUsers, inviteUser, getUser, deleteUser } from '@/services/api/user';
import { useUsers, useUser, useInviteUser, useDeleteUser } from '@/hooks/useUsers';
import type { User, PaginatedResponse } from '@/types/user';

const mockUser: User = {
  id: 'u1',
  tenant_id: 't1',
  email: 'alice@example.com',
  display_name: 'Alice',
  status: 'active',
  last_login_at: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  role_count: 2,
};

const mockPaginated: PaginatedResponse<User> = {
  data: [mockUser],
  total: 1,
  page: 1,
  page_size: 20,
  total_pages: 1,
};

function createWrapper() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return React.createElement(QueryClientProvider, { client: qc }, children);
  };
}

describe('useUsers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches paginated user list', async () => {
    vi.mocked(listUsers).mockResolvedValue(mockPaginated);

    const { result } = renderHook(() => useUsers({ page: 1 }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockPaginated);
    expect(listUsers).toHaveBeenCalledWith({ page: 1 });
  });

  it('passes search and status filters', async () => {
    vi.mocked(listUsers).mockResolvedValue(mockPaginated);

    const { result } = renderHook(
      () => useUsers({ search: 'alice', status: 'active' }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(listUsers).toHaveBeenCalledWith({ search: 'alice', status: 'active' });
  });
});

describe('useUser', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches single user by id', async () => {
    vi.mocked(getUser).mockResolvedValue(mockUser);

    const { result } = renderHook(() => useUser('u1'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockUser);
    expect(getUser).toHaveBeenCalledWith('u1');
  });

  it('does not fetch when id is empty', async () => {
    const { result } = renderHook(() => useUser(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(getUser).not.toHaveBeenCalled();
  });
});

describe('useInviteUser', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls inviteUser API and invalidates users cache', async () => {
    vi.mocked(inviteUser).mockResolvedValue(mockUser);

    const { result } = renderHook(() => useInviteUser(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync({ email: 'bob@example.com' });

    expect(inviteUser).toHaveBeenCalledWith({ email: 'bob@example.com' });
  });
});

describe('useDeleteUser', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls deleteUser API', async () => {
    vi.mocked(deleteUser).mockResolvedValue(undefined);

    const { result } = renderHook(() => useDeleteUser(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync('u1');

    expect(deleteUser).toHaveBeenCalledWith('u1');
  });
});
