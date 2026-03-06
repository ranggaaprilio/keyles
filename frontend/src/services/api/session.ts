/**
 * Session & Activity API service
 */

import api from '../api';
import type { UserSession, UserEvent, PaginatedResponse } from '../../types/user';

const BASE = '/api/v1/admin/users';

export async function listSessions(
  userId: string
): Promise<UserSession[]> {
  const { data } = await api.get(`${BASE}/${userId}/sessions`);
  return data.sessions ?? [];
}

export async function revokeSession(
  userId: string,
  sessionId: string
): Promise<void> {
  await api.delete(`${BASE}/${userId}/sessions/${sessionId}`);
}

export async function listUserActivity(
  userId: string,
  page: number = 1,
  pageSize: number = 25
): Promise<PaginatedResponse<UserEvent>> {
  const { data } = await api.get(`${BASE}/${userId}/activity`, {
    params: { page, page_size: pageSize },
  });
  return {
    data: data.events ?? [],
    total: data.pagination?.total_count ?? 0,
    page: data.pagination?.page ?? page,
    page_size: data.pagination?.page_size ?? pageSize,
    total_pages: data.pagination?.total_pages ?? 0,
  };
}
