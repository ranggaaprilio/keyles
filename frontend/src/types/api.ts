/**
 * API-related types and error handling
 */

export type TenantStatus = 'pending_verification' | 'active' | 'suspended' | 'deleted';

export interface ApiError {
  error: string;
  code?: string;
  details?: Record<string, unknown>;
}

export interface ApiResponse<T> {
  data?: T;
  error?: ApiError;
  status: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export class ApiException extends Error {
  constructor(
    public status: number,
    public error: string,
    public code?: string,
    public details?: Record<string, unknown>
  ) {
    super(error);
    this.name = 'ApiException';
  }
}
