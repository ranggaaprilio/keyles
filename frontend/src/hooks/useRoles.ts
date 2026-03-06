/**
 * TanStack Query hooks for user-scoped Role Management
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { listUserRoles, assignRole, revokeRole } from '../services/api/role';
import type { AssignRoleRequest } from '../types/user';

const USER_KEY = 'user';

export function useUserRoles(userId: string) {
  return useQuery({
    queryKey: [USER_KEY, userId, 'roles'],
    queryFn: () => listUserRoles(userId),
    enabled: !!userId,
  });
}

export function useAssignRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ userId, req }: { userId: string; req: AssignRoleRequest }) =>
      assignRole(userId, req),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: [USER_KEY, vars.userId, 'roles'] });
    },
  });
}

export function useRevokeRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ userId, assignmentId }: { userId: string; assignmentId: number }) =>
      revokeRole(userId, assignmentId),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: [USER_KEY, vars.userId, 'roles'] });
    },
  });
}
