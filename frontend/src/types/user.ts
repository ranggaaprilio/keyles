/**
 * User Management Types
 * Type definitions for user management with RBAC (FR-005)
 */

export type UserStatus = 'pending' | 'active' | 'disabled';

export interface User {
  id: string;
  tenantId: string;
  email: string;
  displayName: string | null;
  status: UserStatus;
  lastLoginAt: string | null;
  createdAt: string;
  updatedAt: string;
  roleCount: number;
}

export interface Invitation {
  id: string;
  tenantId: string;
  email: string;
  displayName: string | null;
  status: 'pending' | 'accepted' | 'expired';
  expiresAt: string;
  createdAt: string;
}

export interface RoleAssignment {
  id: number;
  userId: string;
  clientId: string;
  clientName: string;
  roleName: string;
  isActive: boolean;
  grantedAt: string;
  grantedBy: string;
  revokedAt: string | null;
  revokedBy: string | null;
}

export interface UserEvent {
  id: number;
  eventType: string;
  clientId: string | null;
  clientName: string | null;
  ipAddress: string | null;
  countryCode: string | null;
  details: Record<string, unknown>;
  occurredAt: string;
}

export interface UserSession {
  id: string;
  clientId: string;
  clientName: string;
  createdAt: string;
  lastUsedAt: string | null;
  expiresAt: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface UserListFilters {
  search?: string;
  status?: UserStatus;
  page?: number;
  pageSize?: number;
}

export interface InviteUserRequest {
  email: string;
  displayName?: string;
}

export interface AcceptInvitationRequest {
  password: string;
}

export interface AssignRoleRequest {
  clientId: string;
  roleName: string;
}