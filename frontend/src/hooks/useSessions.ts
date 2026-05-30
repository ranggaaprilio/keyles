import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as sessionApi from '@/services/api/session';

export function useUserSessions(userId: string) {
  return useQuery({
    queryKey: ['users', userId, 'sessions'],
    queryFn: () => sessionApi.listSessions(userId),
    enabled: !!userId,
    staleTime: 15_000,
  });
}

export function useRevokeSession(userId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (sessionId: string) => sessionApi.revokeSession(userId, sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users', userId, 'sessions'] });
    },
  });
}

export function useUserActivity(userId: string, page: number) {
  return useQuery({
    queryKey: ['users', userId, 'activity', page],
    queryFn: () => sessionApi.listUserActivity(userId, page),
    enabled: !!userId,
  });
}