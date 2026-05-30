import api from '../api';
import type {
  UserSession,
  UserEvent,
  PaginatedResponse,
} from '@/types/user';

export async function listSessions(userId: string): Promise<UserSession[]> {
  const { data } = await api.get(`/api/v1/admin/users/${userId}/sessions`);
  return data;
}

export async function revokeSession(
  userId: string,
  sessionId: string
): Promise<void> {
  await api.delete(`/api/v1/admin/users/${userId}/sessions/${sessionId}`);
}

export async function listUserActivity(
  userId: string,
  page: number,
  pageSize?: number
): Promise<PaginatedResponse<UserEvent>> {
  const params = new URLSearchParams();
  params.append('page', String(page));
  if (pageSize) params.append('page_size', String(pageSize));
  const { data } = await api.get(
    `/api/v1/admin/users/${userId}/activity?${params.toString()}`
  );
  return data;
}