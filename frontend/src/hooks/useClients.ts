/**
 * TanStack Query hooks for OAuth client management
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { clientService, type ListClientsParams } from '../services/clientService';
import type {
  CreateClientRequest,
  UpdateClientRequest,
} from '../types/client';

const CLIENTS_KEY = 'clients';

/**
 * Hook to fetch paginated list of clients
 */
export function useClients(params?: ListClientsParams) {
  return useQuery({
    queryKey: [CLIENTS_KEY, params],
    queryFn: () => clientService.list(params),
  });
}

/**
 * Hook to fetch a single client by ID
 */
export function useClient(clientId: string | undefined) {
  return useQuery({
    queryKey: [CLIENTS_KEY, clientId],
    queryFn: () => clientService.get(clientId!),
    enabled: !!clientId,
  });
}

/**
 * Hook to create a new client
 */
export function useCreateClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateClientRequest) => clientService.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [CLIENTS_KEY] });
    },
  });
}

/**
 * Hook to update an existing client
 */
export function useUpdateClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ clientId, data }: { clientId: string; data: UpdateClientRequest }) =>
      clientService.update(clientId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [CLIENTS_KEY] });
    },
  });
}

/**
 * Hook to delete a client
 */
export function useDeleteClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (clientId: string) => clientService.delete(clientId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [CLIENTS_KEY] });
    },
  });
}

/**
 * Hook to rotate client secret
 */
export function useRotateSecret() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (clientId: string) => clientService.rotateSecret(clientId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [CLIENTS_KEY] });
    },
  });
}
