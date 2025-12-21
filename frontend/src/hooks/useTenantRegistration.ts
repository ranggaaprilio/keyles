/**
 * Custom hook for tenant registration with TanStack Query
 */

import { useMutation, UseMutationResult } from '@tanstack/react-query';
import { registerTenant } from '../services/api/tenant';
import { RegisterTenantRequest, RegisterTenantResponse } from '../types/tenant';
import { ApiException } from '../types/api';

export interface UseRegisterTenantOptions {
  onSuccess?: (data: RegisterTenantResponse) => void;
  onError?: (error: ApiException) => void;
}

export function useTenantRegistration(
  options?: UseRegisterTenantOptions
): UseMutationResult<RegisterTenantResponse, ApiException, RegisterTenantRequest> {
  return useMutation<RegisterTenantResponse, ApiException, RegisterTenantRequest>({
    mutationFn: registerTenant,
    onSuccess: (data) => {
      options?.onSuccess?.(data);
    },
    onError: (error) => {
      options?.onError?.(error);
    },
  });
}
