/**
 * TanStack Query hooks for user Sessions and Activity
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { listSessions, revokeSession, listUserActivity } from '../services/api/session';

const USER_KEY = 'user';

export function useUserSessions(userId: string) {
  return useQuery({
    queryKey: [USER_KEY, userId, 'sessions'],
    queryFn: () => listSessions(userId),
    enabled: !!userId,
    staleTime: 15_000,
  });
}

export function useRevokeSession() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ userId, sessionId }: { userId: string; sessionId: string }) =>
      revokeSession(userId, sessionId),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: [USER_KEY, vars.userId, 'sessions'] });
    },
  });
}

export function useUserActivity(userId: string, page: number) {
  return useQuery({
    queryKey: [USER_KEY, userId, 'activity', page],
    queryFn: () => listUserActivity(userId, page),
    enabled: !!userId,
  });
}
