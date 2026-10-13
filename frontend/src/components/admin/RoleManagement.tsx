/**
 * RoleManagement Component
 * Admin interface for managing user role assignments across clients
 * Implements FR-006: Role-based Access Control
 */

import React, { useState, useEffect } from "react";
import {
  assignRole,
  revokeRole,
  listUserRoles,
  listClientRoles,
} from "../../services/roleService";
import type {
  UserRole,
  AssignRoleRequest,
  RevokeRoleRequest,
} from "../../types/role";

interface RoleManagementProps {
  tenantId: string;
}

interface RoleFormState {
  userId: string;
  clientId: string;
  role: string;
}

export const RoleManagement: React.FC<RoleManagementProps> = () => {
  const [roles, setRoles] = useState<UserRole[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [showAssignForm, setShowAssignForm] = useState(false);
  const [formData, setFormData] = useState<RoleFormState>({
    userId: "",
    clientId: "",
    role: "user",
  });
  const [filterBy, setFilterBy] = useState<"user" | "client">("user");
  const [filterId, setFilterId] = useState("");

  useEffect(() => {
    if (filterId) {
      loadRoles();
    }
  }, [filterId, filterBy]);

  const loadRoles = async () => {
    try {
      setLoading(true);
      setError(null);

      let response;
      if (filterBy === "user") {
        response = await listUserRoles(filterId);
      } else {
        response = await listClientRoles(filterId);
      }

      setRoles(response.roles);
    } catch (err: unknown) {
      const axiosErr = err as unknown as { response?: { data?: { message?: string } } };
      setError(axiosErr.response?.data?.message || (err instanceof Error ? err.message : "Failed to load roles"));
    } finally {
      setLoading(false);
    }
  };

  const handleAssignRole = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    try {
      setLoading(true);
      const request: AssignRoleRequest = {
        user_id: formData.userId,
        client_id: formData.clientId,
        role: formData.role,
      };

      await assignRole(request);
      setSuccess("Role assigned successfully");
      setShowAssignForm(false);
      setFormData({ userId: "", clientId: "", role: "user" });

      // Reload roles if currently viewing this user or client
      if (
        (filterBy === "user" && filterId === formData.userId) ||
        (filterBy === "client" && filterId === formData.clientId)
      ) {
        loadRoles();
      }
    } catch (err: unknown) {
      const axiosErr = err as unknown as { response?: { data?: { message?: string } } };
      setError(axiosErr.response?.data?.message || (err instanceof Error ? err.message : "Failed to assign role"));
    } finally {
      setLoading(false);
    }
  };

  const handleRevokeRole = async (role: UserRole) => {
    if (
      !confirm(
        `Revoke ${role.role} role for this user? This will also revoke all their refresh tokens for this client.`,
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
      setSuccess("Role revoked successfully (refresh tokens invalidated)");
      loadRoles();
    } catch (err: unknown) {
      const axiosErr = err as unknown as { response?: { data?: { message?: string } } };
      setError(axiosErr.response?.data?.message || (err instanceof Error ? err.message : "Failed to revoke role"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 bg-background rounded-lg shadow-card border border-border">
      <div className="flex items-start justify-between gap-4 mb-6">
        <h2 className="text-lg font-medium tracking-tight text-foreground">
          Role Management
        </h2>
        <button
          onClick={() => setShowAssignForm(!showAssignForm)}
          className="inline-flex h-10 items-center justify-center rounded-full bg-primary px-5 text-sm font-medium text-primary-foreground hover:bg-primary/85 transition-colors duration-fast"
        >
          {showAssignForm ? "Cancel" : "Assign Role"}
        </button>
      </div>

      {/* Filter Section */}
      <div className="mb-6 p-4 bg-muted rounded-md border border-border">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">
              Filter By
            </label>
            <select
              value={filterBy}
              onChange={(e) => setFilterBy(e.target.value as "user" | "client")}
              className="w-full px-3 py-2 rounded-md border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="user">User</option>
              <option value="client">Client</option>
            </select>
          </div>
          <div className="md:col-span-2">
            <label className="block text-sm font-medium mb-2">
              {filterBy === "user" ? "User ID" : "Client ID"}
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={filterId}
                onChange={(e) => setFilterId(e.target.value)}
                placeholder={`Enter ${filterBy} ID...`}
                className="flex-1 px-3 py-2 rounded-md border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <button
                onClick={loadRoles}
                disabled={!filterId || loading}
                className="px-4 py-2 rounded-full bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/85 disabled:opacity-50 transition-colors"
              >
                Load Roles
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Assign Role Form */}
      {showAssignForm && (
        <form
          onSubmit={handleAssignRole}
          className="mb-6 p-4 rounded-md border border-border bg-muted"
        >
          <h3 className="font-medium tracking-tight text-sm mb-4">
            Assign New Role
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">
                User ID *
              </label>
              <input
                type="text"
                required
                value={formData.userId}
                onChange={(e) =>
                  setFormData({ ...formData, userId: e.target.value })
                }
                className="w-full px-3 py-2 rounded-md border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="user-uuid"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">
                Client ID *
              </label>
              <input
                type="text"
                required
                value={formData.clientId}
                onChange={(e) =>
                  setFormData({ ...formData, clientId: e.target.value })
                }
                className="w-full px-3 py-2 rounded-md border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="client-id"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">
                Role *
              </label>
              <select
                required
                value={formData.role}
                onChange={(e) =>
                  setFormData({ ...formData, role: e.target.value })
                }
                className="w-full px-3 py-2 rounded-md border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="user">user</option>
                <option value="admin">admin</option>
                <option value="viewer">viewer</option>
                <option value="editor">editor</option>
              </select>
            </div>
          </div>
          <div className="mt-4 flex justify-end">
            <button
              type="submit"
              disabled={loading}
              className="px-6 py-2 rounded-full bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/85 disabled:opacity-50 transition-colors"
            >
              {loading ? "Assigning..." : "Assign Role"}
            </button>
          </div>
        </form>
      )}

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

      {/* Roles Table */}
      {loading && !roles.length ? (
        <div className="text-center py-8 text-muted-foreground text-sm">Loading roles...</div>
      ) : roles.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground text-sm">
          No roles found.{" "}
          {filterId ? "Try a different search." : "Enter a filter to search."}
        </div>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-border">
          <table className="min-w-full">
            <thead className="bg-muted">
              <tr className="border-b border-border">
                <th className="px-6 py-3 text-left text-sm font-medium">
                  User ID
                </th>
                <th className="px-6 py-3 text-left text-sm font-medium">
                  Client ID
                </th>
                <th className="px-6 py-3 text-left text-sm font-medium">
                  Role
                </th>
                <th className="px-6 py-3 text-left text-sm font-medium">
                  Assigned At
                </th>
                <th className="px-6 py-3 text-left text-sm font-medium">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-sm font-medium">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {roles.map((role) => (
                <tr key={role.id} className="border-b border-border hover:bg-muted">
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {role.user_email || role.user_id.slice(0, 8)}...
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {role.client_name || role.client_id}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="inline-flex items-center rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-foreground">
                      {role.role}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {new Date(role.assigned_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        role.is_active
                          ? "bg-success/10 text-success"
                          : "bg-destructive/10 text-destructive"
                      }`}
                    >
                      {role.is_active ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    {role.is_active && (
                      <button
                        onClick={() => handleRevokeRole(role)}
                        disabled={loading}
                        className="text-destructive hover:text-destructive underline disabled:opacity-50"
                      >
                        Revoke
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="mt-4 text-sm text-muted-foreground">
        <p>Total roles: {roles.length}</p>
        <p className="mt-2">
          <strong>Note:</strong> Revoking a role will also invalidate all
          refresh tokens for that user-client combination (FR-006e).
        </p>
      </div>
    </div>
  );
};

export default RoleManagement;