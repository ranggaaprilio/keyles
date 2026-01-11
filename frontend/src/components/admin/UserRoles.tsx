/**
 * UserRoles Component
 * Displays all roles for a specific user across all clients
 * Implements FR-006: Role-based Access Control
 */

import React, { useState, useEffect } from 'react';
import { listUserRoles, revokeRole } from '../../services/roleService';
import type { UserRole, RevokeRoleRequest } from '../../types/role';

interface UserRolesProps {
  userId: string;
  userEmail?: string;
  onRoleRevoked?: () => void;
}

export const UserRoles: React.FC<UserRolesProps> = ({
  userId,
  userEmail,
  onRoleRevoked,
}) => {
  const [roles, setRoles] = useState<UserRole[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    loadUserRoles();
  }, [userId]);

  const loadUserRoles = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await listUserRoles(userId);
      setRoles(response.roles);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load user roles');
    } finally {
      setLoading(false);
    }
  };

  const handleRevokeRole = async (role: UserRole) => {
    if (
      !confirm(
        `Revoke ${role.role} role for client "${role.client_name || role.client_id}"? ` +
        `This will invalidate all refresh tokens for this user-client combination.`
      )
    ) {
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const request: RevokeRoleRequest = {
        user_id: role.user_id,
        client_id: role.client_id,
        role: role.role,
      };

      await revokeRole(request);
      setSuccess('Role revoked successfully');
      
      // Reload roles
      await loadUserRoles();
      
      // Notify parent component
      if (onRoleRevoked) {
        onRoleRevoked();
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to revoke role');
    } finally {
      setLoading(false);
    }
  };

  const activeRoles = roles.filter(r => r.is_active);
  const inactiveRoles = roles.filter(r => !r.is_active);

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="mb-4">
        <h3 className="text-xl font-semibold text-gray-800">User Roles</h3>
        {userEmail && (
          <p className="text-sm text-gray-600 mt-1">{userEmail}</p>
        )}
        <p className="text-xs text-gray-500 mt-1">User ID: {userId}</p>
      </div>

      {/* Status Messages */}
      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}
      {success && (
        <div className="mb-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded">
          {success}
        </div>
      )}

      {loading && !roles.length ? (
        <div className="text-center py-8 text-gray-500">Loading roles...</div>
      ) : roles.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          No roles assigned to this user.
        </div>
      ) : (
        <>
          {/* Active Roles */}
          {activeRoles.length > 0 && (
            <div className="mb-6">
              <h4 className="text-lg font-medium text-gray-700 mb-3">
                Active Roles ({activeRoles.length})
              </h4>
              <div className="space-y-3">
                {activeRoles.map((role) => (
                  <div
                    key={role.id}
                    className="flex items-center justify-between p-4 border rounded-lg bg-green-50 border-green-200"
                  >
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-gray-800">
                          {role.client_name || role.client_id}
                        </span>
                        <span className="px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                          {role.role}
                        </span>
                      </div>
                      <div className="mt-1 text-sm text-gray-600">
                        Client ID: {role.client_id}
                      </div>
                      <div className="mt-1 text-xs text-gray-500">
                        Assigned {new Date(role.assigned_at).toLocaleDateString()} by{' '}
                        {role.assigned_by}
                      </div>
                    </div>
                    <button
                      onClick={() => handleRevokeRole(role)}
                      disabled={loading}
                      className="ml-4 px-4 py-2 text-sm font-medium text-red-600 hover:text-red-800 hover:bg-red-100 rounded transition disabled:opacity-50"
                    >
                      Revoke
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Inactive Roles (History) */}
          {inactiveRoles.length > 0 && (
            <div>
              <h4 className="text-lg font-medium text-gray-700 mb-3">
                Revoked Roles ({inactiveRoles.length})
              </h4>
              <div className="space-y-2">
                {inactiveRoles.map((role) => (
                  <div
                    key={role.id}
                    className="flex items-center justify-between p-3 border rounded bg-gray-50 border-gray-200"
                  >
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="text-gray-600">
                          {role.client_name || role.client_id}
                        </span>
                        <span className="px-2 py-1 text-xs font-semibold rounded-full bg-gray-200 text-gray-600">
                          {role.role}
                        </span>
                        <span className="px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">
                          Revoked
                        </span>
                      </div>
                      <div className="mt-1 text-xs text-gray-500">
                        Assigned {new Date(role.assigned_at).toLocaleDateString()}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}

      {/* Summary */}
      {roles.length > 0 && (
        <div className="mt-6 pt-4 border-t text-sm text-gray-600">
          <p>
            <strong>Summary:</strong> {activeRoles.length} active role(s) across{' '}
            {new Set(activeRoles.map(r => r.client_id)).size} client(s)
          </p>
          {inactiveRoles.length > 0 && (
            <p className="mt-1">
              {inactiveRoles.length} revoked role(s) (history)
            </p>
          )}
          <p className="mt-2 text-xs">
            <strong>Note:</strong> Revoking a role will invalidate all refresh tokens for
            this user-client combination (FR-006e).
          </p>
        </div>
      )}
    </div>
  );
};

export default UserRoles;
