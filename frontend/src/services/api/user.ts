/**
 * User Management API service
 */

import api from '../api';
import type {
  User,
  PaginatedResponse,
  UserListFilters,
  InviteUserRequest,
  UserStatus,
} from '../../types/user';

const BASE = '/api/v1/admin/users';

export async function listUsers(
  filters: UserListFilters = {}
): Promise<PaginatedResponse<User>> {
  const params: Record<string, string | number> = {};
  if (filters.search) params['search'] = filters.search;
  if (filters.status) params['status'] = filters.status;
  if (filters.page) params['page'] = filters.page;
  if (filters.page_size) params['page_size'] = filters.page_size;

  const { data } = await api.get(BASE, { params });
  return data;
}

export async function inviteUser(req: InviteUserRequest): Promise<User> {
  const { data } = await api.post(`${BASE}/invite`, req);
  return data;
}

export async function getUser(id: string): Promise<User> {
  const { data } = await api.get(`${BASE}/${id}`);
  return data;
}

export async function updateUser(
  id: string,
  req: { display_name: string }
): Promise<User> {
  const { data } = await api.patch(`${BASE}/${id}`, req);
  return data;
}

export async function deleteUser(id: string): Promise<void> {
  await api.delete(`${BASE}/${id}`);
}

export async function updateUserStatus(
  id: string,
  status: UserStatus
): Promise<User> {
  const { data } = await api.patch(`${BASE}/${id}/status`, { status });
  return data;
}

export async function resendInvitation(userId: string): Promise<void> {
  await api.post(`${BASE}/${userId}/resend-invitation`);
}
