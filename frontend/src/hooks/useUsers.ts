import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as userApi from '@/services/api/user';
import type { UserListFilters } from '@/types/user';

const USERS_KEY = 'users';

export function useUsers(filters: UserListFilters) {
  return useQuery({
    queryKey: [USERS_KEY, filters],
    queryFn: () => userApi.listUsers(filters),
    staleTime: 30_000,
  });
}

export function useUser(id: string) {
  return useQuery({
    queryKey: [USERS_KEY, id],
    queryFn: () => userApi.getUser(id),
    enabled: !!id,
  });
}

export function useInviteUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: userApi.inviteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [USERS_KEY] });
    },
  });
}

export function useUpdateUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, displayName }: { id: string; displayName: string }) =>
      userApi.updateUser(id, { displayName }),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: [USERS_KEY] });
      queryClient.invalidateQueries({ queryKey: [USERS_KEY, id] });
    },
  });
}

export function useDeleteUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: userApi.deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [USERS_KEY] });
    },
  });
}

export function useUpdateUserStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'active' | 'disabled' }) =>
      userApi.updateUserStatus(id, status),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: [USERS_KEY] });
      queryClient.invalidateQueries({ queryKey: [USERS_KEY, id] });
    },
  });
}

export function useResendInvitation() {
  return useMutation({
    mutationFn: userApi.resendInvitation,
  });
}