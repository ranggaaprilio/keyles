/**
 * Client Service - API client for OAuth client management
 */

import axios from 'axios';
import type {
  Client,
  CreateClientRequest,
  CreateClientResponse,
  UpdateClientRequest,
  RotateSecretResponse,
  ListClientsResponse,
} from '../types/client';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Create axios instance with default config
const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1/admin`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token interceptor
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

/**
 * Client Service API
 */
export const clientService = {
  /**
   * Create a new OAuth client
   * @param data - Client creation data
   * @returns Created client with client_secret (only returned once)
   */
  async create(data: CreateClientRequest): Promise<CreateClientResponse> {
    const response = await api.post<CreateClientResponse>('/clients', data);
    return response.data;
  },

  /**
   * Get all clients for the current tenant
   * @returns List of clients
   */
  async list(): Promise<ListClientsResponse> {
    const response = await api.get<ListClientsResponse>('/clients');
    return response.data;
  },

  /**
   * Get a single client by ID
   * @param clientId - The client_id to fetch
   * @returns Client details
   */
  async get(clientId: string): Promise<Client> {
    const response = await api.get<Client>(`/clients/${clientId}`);
    return response.data;
  },

  /**
   * Update an existing client
   * @param clientId - The client_id to update
   * @param data - Update data
   * @returns Updated client
   */
  async update(clientId: string, data: UpdateClientRequest): Promise<Client> {
    const response = await api.put<Client>(`/clients/${clientId}`, data);
    return response.data;
  },

  /**
   * Delete (deactivate) a client
   * @param clientId - The client_id to delete
   */
  async delete(clientId: string): Promise<void> {
    await api.delete(`/clients/${clientId}`);
  },

  /**
   * Rotate client secret
   * @param clientId - The client_id to rotate secret for
   * @returns New client secret (only returned once)
   */
  async rotateSecret(clientId: string): Promise<RotateSecretResponse> {
    const response = await api.post<RotateSecretResponse>(
      `/clients/${clientId}/rotate-secret`
    );
    return response.data;
  },
};

export default clientService;
