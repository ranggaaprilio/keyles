/**
 * TypeScript interfaces for OAuth Client management
 */

// Client entity as returned from API
export interface Client {
  client_id: string;
  client_name: string;
  redirect_uris: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Response when creating a new client (includes secret)
export interface CreateClientResponse {
  client_id: string;
  client_secret: string; // Only returned once at creation
  client_name: string;
  redirect_uris: string[];
  is_active: boolean;
  created_at: string;
}

// Request to create a new client
export interface CreateClientRequest {
  client_name: string;
  redirect_uris: string[];
}

// Request to update an existing client
export interface UpdateClientRequest {
  client_name?: string;
  redirect_uris?: string[];
  is_active?: boolean;
}

// Response for rotating a client secret
export interface RotateSecretResponse {
  client_id: string;
  client_secret: string; // New secret - only returned once
  rotated_at: string;
}

// Response for listing clients
export interface ListClientsResponse {
  clients: Client[];
  total: number;
}

// API error response
export interface ClientAPIError {
  error: string;
  message?: string;
  status?: number;
}
