import api from '../api';
import type {
  RoleAssignment,
  AssignRoleRequest,
} from '@/types/user';

export async function listUserRoles(userId: string): Promise<RoleAssignment[]> {
  const { data } = await api.get(`/api/v1/admin/users/${userId}/roles`);
  return data;
}

export async function assignRole(
  userId: string,
  req: AssignRoleRequest
): Promise<RoleAssignment> {
  const { data } = await api.post(`/api/v1/admin/users/${userId}/roles`, req);
  return data;
}

export async function revokeRole(
  userId: string,
  assignmentId: number
): Promise<void> {
  await api.delete(`/api/v1/admin/users/${userId}/roles/${assignmentId}`);
}