/**
 * TanStack Query hooks for User Management
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  listUsers,
  inviteUser,
  getUser,
  updateUser,
  deleteUser,
  updateUserStatus,
  resendInvitation,
} from '../services/api/user';
import type { UserListFilters, UserStatus } from '../types/user';

const USERS_KEY = 'users';
const USER_KEY = 'user';

export function useUsers(filters: UserListFilters = {}) {
  return useQuery({
    queryKey: [USERS_KEY, filters],
    queryFn: () => listUsers(filters),
    staleTime: 30_000,
  });
}

export function useUser(id: string) {
  return useQuery({
    queryKey: [USER_KEY, id],
    queryFn: () => getUser(id),
    enabled: !!id,
  });
}

export function useInviteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: Parameters<typeof inviteUser>[0]) => inviteUser(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: [USERS_KEY] });
    },
  });
}

export function useUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, displayName }: { id: string; displayName: string }) =>
      updateUser(id, { display_name: displayName }),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: [USER_KEY, vars.id] });
      qc.invalidateQueries({ queryKey: [USERS_KEY] });
    },
  });
}

export function useDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: [USERS_KEY] });
    },
  });
}

export function useUpdateUserStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: UserStatus }) =>
      updateUserStatus(id, status),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: [USER_KEY, vars.id] });
      qc.invalidateQueries({ queryKey: [USERS_KEY] });
    },
  });
}

export function useResendInvitation() {
  return useMutation({
    mutationFn: resendInvitation,
  });
}
