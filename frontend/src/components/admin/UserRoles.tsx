/**
 * UserRoles Component
 * Displays all roles for a specific user across all clients
 * Implements FR-006: Role-based Access Control
 */

import React, { useState, useEffect } from "react";
import { listUserRoles, revokeRole } from "../../services/roleService";
import type { UserRole, RevokeRoleRequest } from "../../types/role";

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
    } catch (err: unknown) {
      const axiosErr = err as unknown as { response?: { data?: { message?: string } } };
      setError(axiosErr.response?.data?.message || (err instanceof Error ? err.message : "Failed to load user roles"));
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
      setSuccess("Role revoked successfully");

      // Reload roles
      await loadUserRoles();

      // Notify parent component
      if (onRoleRevoked) {
        onRoleRevoked();
      }
    } catch (err: unknown) {
      const axiosErr = err as unknown as { response?: { data?: { message?: string } } };
      setError(axiosErr.response?.data?.message || (err instanceof Error ? err.message : "Failed to revoke role"));
    } finally {
      setLoading(false);
    }
  };

  const activeRoles = roles.filter((r) => r.is_active);
  const inactiveRoles = roles.filter((r) => !r.is_active);

  return (
    <div className="bg-white shadow-[2px_2px_0_#000] border border-black p-6">
      {/* Eyebrow */}
      <div className="bg-[#8e8a25] px-4 py-2 -mt-6 -mx-6 mb-6 border-b border-black">
        <h3 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-white text-lg">
          User Roles
        </h3>
      </div>

      <div className="mb-4">
        {userEmail && <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-600 mt-1">{userEmail}</p>}
        <p className="font-['Times_New_Roman',Times,serif] text-xs text-gray-500 mt-1">User ID: {userId}</p>
      </div>

      {/* Status Messages */}
      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-700 text-red-800 font-['Times_New_Roman',Times,serif] text-sm">
          {error}
        </div>
      )}
      {success && (
        <div className="mb-4 p-3 bg-green-100 border border-green-700 text-green-800 font-['Times_New_Roman',Times,serif] text-sm">
          {success}
        </div>
      )}

      {loading && !roles.length ? (
        <div className="text-center py-8 text-gray-500 font-['Times_New_Roman',Times,serif] text-sm">Loading roles...</div>
      ) : roles.length === 0 ? (
        <div className="text-center py-8 text-gray-500 font-['Times_New_Roman',Times,serif] text-sm">
          No roles assigned to this user.
        </div>
      ) : (
        <>
          {/* Active Roles */}
          {activeRoles.length > 0 && (
            <div className="mb-6">
              <h4 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-sm mb-3">
                Active Roles ({activeRoles.length})
              </h4>
              <div className="space-y-3">
                {activeRoles.map((role) => (
                  <div
                    key={role.id}
                    className="flex items-center justify-between p-4 border border-black bg-green-50"
                  >
                    <div className="flex-1 font-['Times_New_Roman',Times,serif]">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-gray-800 text-sm">
                          {role.client_name || role.client_id}
                        </span>
                        <span className="px-2 py-1 text-xs font-bold border border-black bg-[#8c9ae0] text-white">
                          {role.role}
                        </span>
                      </div>
                      <div className="mt-1 text-sm text-gray-600">
                        Client ID: {role.client_id}
                      </div>
                      <div className="mt-1 text-xs text-gray-500">
                        Assigned{" "}
                        {new Date(role.assigned_at).toLocaleDateString()} by{" "}
                        {role.assigned_by}
                      </div>
                    </div>
                    <button
                      onClick={() => handleRevokeRole(role)}
                      disabled={loading}
                      className="ml-4 px-4 py-2 text-[12px] font-bold uppercase tracking-[1.5px] font-[Helvetica,Arial,system-ui,sans-serif] border border-red-700 bg-red-700 text-white hover:bg-red-800 disabled:opacity-50 transition-colors"
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
              <h4 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-sm mb-3">
                Revoked Roles ({inactiveRoles.length})
              </h4>
              <div className="space-y-2">
                {inactiveRoles.map((role) => (
                  <div
                    key={role.id}
                    className="flex items-center justify-between p-3 border border-black bg-gray-100"
                  >
                    <div className="flex-1 font-['Times_New_Roman',Times,serif]">
                      <div className="flex items-center gap-2">
                        <span className="text-gray-600 text-sm">
                          {role.client_name || role.client_id}
                        </span>
                        <span className="px-2 py-1 text-xs font-bold border border-black bg-gray-400 text-white">
                          {role.role}
                        </span>
                        <span className="px-2 py-1 text-xs font-bold border border-black bg-red-700 text-white">
                          Revoked
                        </span>
                      </div>
                      <div className="mt-1 text-xs text-gray-500">
                        Assigned{" "}
                        {new Date(role.assigned_at).toLocaleDateString()}
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
        <div className="mt-6 pt-4 border-t border-black font-['Times_New_Roman',Times,serif] text-sm text-gray-600">
          <p>
            <strong>Summary:</strong> {activeRoles.length} active role(s) across{" "}
            {new Set(activeRoles.map((r) => r.client_id)).size} client(s)
          </p>
          {inactiveRoles.length > 0 && (
            <p className="mt-1">
              {inactiveRoles.length} revoked role(s) (history)
            </p>
          )}
          <p className="mt-2 text-xs">
            <strong>Note:</strong> Revoking a role will invalidate all refresh
            tokens for this user-client combination (FR-006e).
          </p>
        </div>
      )}
    </div>
  );
};

export default UserRoles;