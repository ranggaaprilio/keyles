import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as roleApi from '@/services/api/role';
import type { AssignRoleRequest } from '@/types/user';

export function useUserRoles(userId: string) {
  return useQuery({
    queryKey: ['users', userId, 'roles'],
    queryFn: () => roleApi.listUserRoles(userId),
    enabled: !!userId,
  });
}

export function useAssignRole(userId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: AssignRoleRequest) => roleApi.assignRole(userId, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users', userId, 'roles'] });
    },
  });
}

export function useRevokeRole(userId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (assignmentId: number) => roleApi.revokeRole(userId, assignmentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users', userId, 'roles'] });
    },
  });
}