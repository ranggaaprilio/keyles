/**
 * Role Management Service
 * API client for user role assignment and management
 * Implements FR-006: Role-based Access Control
 */

import axiosInstance from './api';
import type {
  UserRole,
  AssignRoleRequest,
  RevokeRoleRequest,
  UserRoleListResponse,
  ClientRoleListResponse,
} from '../types/role';

const BASE_URL = '/api/v1/admin/roles';

/**
 * Assign a role to a user for a specific client (FR-006a)
 * @param request - Role assignment details
 * @returns The created role assignment
 */
export const assignRole = async (request: AssignRoleRequest): Promise<UserRole> => {
  const response = await axiosInstance.post<UserRole>(`${BASE_URL}/assign`, request);
  return response.data;
};

/**
 * Revoke a user's role for a specific client (FR-006b)
 * Cascades to revoke all refresh tokens per FR-006e
 * @param request - Role revocation details
 * @returns Success confirmation
 */
export const revokeRole = async (request: RevokeRoleRequest): Promise<void> => {
  await axiosInstance.post(`${BASE_URL}/revoke`, request);
};

/**
 * List all roles for a specific user
 * @param userId - The user's ID
 * @param clientId - Optional: Filter by specific client
 * @returns List of user's role assignments
 */
export const listUserRoles = async (
  userId: string,
  clientId?: string
): Promise<UserRoleListResponse> => {
  const params = clientId ? { client_id: clientId } : undefined;
  const response = await axiosInstance.get<UserRoleListResponse>(
    `${BASE_URL}/users/${userId}`,
    { params }
  );
  return response.data;
};

/**
 * List all role assignments for a specific client
 * @param clientId - The client's ID
 * @param userId - Optional: Filter by specific user
 * @returns List of role assignments for the client
 */
export const listClientRoles = async (
  clientId: string,
  userId?: string
): Promise<ClientRoleListResponse> => {
  const params = userId ? { user_id: userId } : undefined;
  const response = await axiosInstance.get<ClientRoleListResponse>(
    `${BASE_URL}/clients/${clientId}`,
    { params }
  );
  return response.data;
};

/**
 * Check if a user has any role for a client (FR-006d)
 * @param userId - The user's ID
 * @param clientId - The client's ID
 * @returns True if user has at least one active role
 */
export const hasRole = async (userId: string, clientId: string): Promise<boolean> => {
  try {
    const response = await listUserRoles(userId, clientId);
    return response.roles.some(role => role.is_active);
  } catch (error) {
    console.error('Error checking user role:', error);
    return false;
  }
};

/**
 * Bulk assign roles to multiple users
 * @param assignments - Array of role assignments
 * @returns Array of created role assignments
 */
export const bulkAssignRoles = async (
  assignments: AssignRoleRequest[]
): Promise<UserRole[]> => {
  const promises = assignments.map(assignment => assignRole(assignment));
  return Promise.all(promises);
};

/**
 * Bulk revoke roles from multiple users
 * @param revocations - Array of role revocations
 */
export const bulkRevokeRoles = async (
  revocations: RevokeRoleRequest[]
): Promise<void> => {
  const promises = revocations.map(revocation => revokeRole(revocation));
  await Promise.all(promises);
};

export default {
  assignRole,
  revokeRole,
  listUserRoles,
  listClientRoles,
  hasRole,
  bulkAssignRoles,
  bulkRevokeRoles,
};
