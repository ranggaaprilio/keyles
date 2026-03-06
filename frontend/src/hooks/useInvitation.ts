/**
 * TanStack Query hooks for public invitation flow (no auth)
 */

import { useQuery, useMutation } from '@tanstack/react-query';
import {
  validateInvitation,
  acceptInvitation,
} from '../services/api/invitation';
import type { AcceptInvitationRequest } from '../types/user';

export function useValidateInvitation(token: string) {
  return useQuery({
    queryKey: ['invitation', token],
    queryFn: () => validateInvitation(token),
    enabled: !!token,
    retry: false,
  });
}

export function useAcceptInvitation(token: string) {
  return useMutation({
    mutationFn: (req: AcceptInvitationRequest) => acceptInvitation(token, req),
  });
}
