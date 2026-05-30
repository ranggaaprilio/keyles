import { useMutation } from '@tanstack/react-query';
import * as userApi from '@/services/api/user';

export function useAcceptInvitation() {
  return useMutation({
    mutationFn: ({ token, password }: { token: string; password: string }) =>
      userApi.acceptInvitation(token, { password }),
  });
}