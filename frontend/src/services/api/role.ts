/**
 * User-scoped Role Management API service
 */

import api from '../api';
import type { RoleAssignment, AssignRoleRequest } from '../../types/user';

const BASE = '/api/v1/admin/users';

export async function listUserRoles(
  userId: string
): Promise<RoleAssignment[]> {
  const { data } = await api.get(`${BASE}/${userId}/roles`);
  return data.roles ?? [];
}

export async function assignRole(
  userId: string,
  req: AssignRoleRequest
): Promise<RoleAssignment> {
  const { data } = await api.post(`${BASE}/${userId}/roles`, req);
  return data;
}

export async function revokeRole(
  userId: string,
  assignmentId: number
): Promise<void> {
  await api.delete(`${BASE}/${userId}/roles/${assignmentId}`);
}
