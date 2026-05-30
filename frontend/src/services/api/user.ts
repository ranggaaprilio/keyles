import api from '../api';
import type {
  User,
  PaginatedResponse,
  UserListFilters,
  InviteUserRequest,
  AcceptInvitationRequest,
} from '@/types/user';

const BASE_URL = '/api/v1/admin/users';

export async function listUsers(
  filters: UserListFilters
): Promise<PaginatedResponse<User>> {
  const params = new URLSearchParams();
  if (filters.search) params.append('search', filters.search);
  if (filters.status) params.append('status', filters.status);
  if (filters.page) params.append('page', String(filters.page));
  if (filters.pageSize) params.append('page_size', String(filters.pageSize));
  const { data } = await api.get(`${BASE_URL}?${params.toString()}`);
  return data;
}

export async function inviteUser(req: InviteUserRequest): Promise<User> {
  const { data } = await api.post(`${BASE_URL}/invite`, req);
  return data;
}

export async function getUser(id: string): Promise<User> {
  const { data } = await api.get(`${BASE_URL}/${id}`);
  return data;
}

export async function updateUser(
  id: string,
  req: { displayName: string }
): Promise<User> {
  const { data } = await api.patch(`${BASE_URL}/${id}`, req);
  return data;
}

export async function deleteUser(id: string): Promise<void> {
  await api.delete(`${BASE_URL}/${id}`);
}

export async function updateUserStatus(
  id: string,
  status: 'active' | 'disabled'
): Promise<User> {
  const { data } = await api.patch(`${BASE_URL}/${id}/status`, { status });
  return data;
}

export async function resendInvitation(userId: string): Promise<void> {
  await api.post(`${BASE_URL}/${userId}/resend-invitation`);
}

export async function acceptInvitation(
  token: string,
  req: AcceptInvitationRequest
): Promise<void> {
  await api.post(`/api/v1/invitations/${token}/accept`, req);
}