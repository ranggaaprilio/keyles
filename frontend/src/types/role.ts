/**
 * Role Management Types
 * Type definitions for user role assignment and management (FR-006)
 */

/**
 * UserRole represents a user's role assignment for a client
 */
export interface UserRole {
  id: string;
  user_id: string;
  client_id: string;
  role: string;
  assigned_at: string;
  assigned_by: string;
  is_active: boolean;
  user_email?: string;
  client_name?: string;
}

/**
 * AssignRoleRequest - Request body for assigning a role to a user
 */
export interface AssignRoleRequest {
  user_id: string;
  client_id: string;
  role: string;
}

/**
 * RevokeRoleRequest - Request body for revoking a user's role
 */
export interface RevokeRoleRequest {
  user_id: string;
  client_id: string;
  role: string;
}

/**
 * UserRoleListResponse - Response from listing user roles
 */
export interface UserRoleListResponse {
  roles: UserRole[];
  total: number;
}

/**
 * ClientRoleListResponse - Response from listing roles for a client
 */
export interface ClientRoleListResponse {
  roles: UserRole[];
  total: number;
}

/**
 * RoleAssignmentMatrix - For displaying user-client role relationships
 */
export interface RoleAssignmentMatrix {
  user_id: string;
  user_email: string;
  clients: {
    client_id: string;
    client_name: string;
    roles: string[];
    is_active: boolean;
  }[];
}

/**
 * Available role options (extend based on requirements)
 */
export const AVAILABLE_ROLES = [
  'admin',
  'user',
  'viewer',
  'editor',
] as const;

export type RoleType = typeof AVAILABLE_ROLES[number];
