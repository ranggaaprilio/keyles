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
    <div className="bg-background rounded-lg shadow-card border border-border p-6">
      <h3 className="text-lg font-medium tracking-tight text-foreground mb-4">
        User Roles
      </h3>

      <div className="mb-4">
        {userEmail && <p className="text-sm text-muted-foreground mt-1">{userEmail}</p>}
        <p className="text-xs text-muted-foreground mt-1">User ID: {userId}</p>
      </div>

      {/* Status Messages */}
      {error && (
        <div className="mb-4 p-3 bg-destructive/5 border border-destructive text-destructive text-sm">
          {error}
        </div>
      )}
      {success && (
        <div className="mb-4 p-3 bg-success/10 border border-success/30 text-success text-sm">
          {success}
        </div>
      )}

      {loading && !roles.length ? (
        <div className="text-center py-8 text-muted-foreground text-sm">Loading roles...</div>
      ) : roles.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground text-sm">
          No roles assigned to this user.
        </div>
      ) : (
        <>
          {/* Active Roles */}
          {activeRoles.length > 0 && (
            <div className="mb-6">
              <h4 className="font-medium tracking-tight text-sm mb-3">
                Active Roles ({activeRoles.length})
              </h4>
              <div className="space-y-3">
                {activeRoles.map((role) => (
                  <div
                    key={role.id}
                    className="flex items-center justify-between p-4 rounded-md border border-border bg-success/10"
                  >
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-foreground text-sm">
                          {role.client_name || role.client_id}
                        </span>
                        <span className="inline-flex items-center rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-foreground">
                          {role.role}
                        </span>
                      </div>
                      <div className="mt-1 text-sm text-muted-foreground">
                        Client ID: {role.client_id}
                      </div>
                      <div className="mt-1 text-xs text-muted-foreground">
                        Assigned{" "}
                        {new Date(role.assigned_at).toLocaleDateString()} by{" "}
                        {role.assigned_by}
                      </div>
                    </div>
                    <button
                      onClick={() => handleRevokeRole(role)}
                      disabled={loading}
                      className="ml-4 px-4 py-2 text-sm font-medium border border-destructive bg-destructive text-primary-foreground hover:bg-destructive/90 disabled:opacity-50 transition-colors"
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
              <h4 className="font-medium tracking-tight text-sm mb-3">
                Revoked Roles ({inactiveRoles.length})
              </h4>
              <div className="space-y-2">
                {inactiveRoles.map((role) => (
                  <div
                    key={role.id}
                    className="flex items-center justify-between p-3 rounded-md border border-border bg-muted"
                  >
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="text-muted-foreground text-sm">
                          {role.client_name || role.client_id}
                        </span>
                        <span className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-muted text-muted-foreground">
                          {role.role}
                        </span>
                        <span className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-destructive/10 text-destructive">
                          Revoked
                        </span>
                      </div>
                      <div className="mt-1 text-xs text-muted-foreground">
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
        <div className="mt-6 pt-4 border-t border-border text-sm text-muted-foreground">
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