/**
 * TypeScript types for User Management & RBAC (Feature 005)
 */

// --- Enums & Unions ---

export type UserStatus = 'pending' | 'active' | 'disabled';

export type InvitationStatus = 'pending' | 'accepted' | 'expired';

// --- Core Entities ---

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  display_name: string;
  status: UserStatus;
  last_login_at: string | null;
  created_at: string;
  updated_at: string;
  role_count?: number;
}

export interface Invitation {
  id: string;
  tenant_id: string;
  email: string;
  display_name: string;
  status: InvitationStatus;
  expires_at: string;
  created_at: string;
}

export interface RoleAssignment {
  id: number;
  user_id: string;
  client_id: string;
  client_name?: string;
  tenant_id: string;
  role: string;
  is_active: boolean;
  granted_at: string;
  granted_by: string;
  revoked_at?: string | null;
  revoked_by?: string | null;
}

export interface UserEvent {
  id: number;
  event_type: string;
  client_id?: string | null;
  client_name?: string;
  ip_address?: string | null;
  country_code?: string | null;
  details?: Record<string, unknown> | null;
  occurred_at: string;
}

export interface UserSession {
  id: number;
  client_id: string;
  client_name?: string;
  created_at: string;
  last_used_at?: string | null;
  expires_at: string;
}

// --- Pagination ---

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// --- Filters ---

export interface UserListFilters {
  search?: string;
  status?: UserStatus | '';
  page?: number;
  page_size?: number;
}

// --- Request Types ---

export interface InviteUserRequest {
  email: string;
  display_name?: string | undefined;
}

export interface AcceptInvitationRequest {
  password: string;
}

export interface AssignRoleRequest {
  client_id: string;
  role_name: string;
}
