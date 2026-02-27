/**
 * TypeScript interfaces for OAuth Client management
 */

// Client type enum
export type ClientType = 'confidential' | 'public';

// Client entity as returned from API
export interface Client {
  client_id: string;
  client_name: string;
  description: string | null;
  client_type: ClientType;
  redirect_uris: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Response when creating a new client (includes secret)
export interface CreateClientResponse {
  client_id: string;
  client_secret: string | null; // null for public clients
  client_name: string;
  description: string | null;
  client_type: ClientType;
  redirect_uris: string[];
  is_active: boolean;
  created_at: string;
}

// Request to create a new client
export interface CreateClientRequest {
  client_name: string;
  description?: string;
  client_type: ClientType;
  redirect_uris: string[];
}

// Request to update an existing client
export interface UpdateClientRequest {
  client_name?: string;
  description?: string | null;
  redirect_uris?: string[];
  is_active?: boolean;
}

// Response for rotating a client secret
export interface RotateSecretResponse {
  client_id: string;
  client_secret: string; // New secret - only returned once
  rotated_at: string;
}

// Paginated response wrapper
export interface PaginatedResponse<T> {
  clients: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// Response for listing clients
export type ListClientsResponse = PaginatedResponse<Client>;

// API error response
export interface ClientAPIError {
  error: string;
  message?: string;
  status?: number;
}
