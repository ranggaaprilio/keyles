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
    } catch (err: any) {
      setError(err.response?.data?.message || "Failed to load roles");
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
    } catch (err: any) {
      setError(err.response?.data?.message || "Failed to assign role");
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
    } catch (err: any) {
      setError(err.response?.data?.message || "Failed to revoke role");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 bg-white rounded-lg shadow-md">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-800">Role Management</h2>
        <button
          onClick={() => setShowAssignForm(!showAssignForm)}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
        >
          {showAssignForm ? "Cancel" : "Assign Role"}
        </button>
      </div>

      {/* Filter Section */}
      <div className="mb-6 p-4 bg-gray-50 rounded">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Filter By
            </label>
            <select
              value={filterBy}
              onChange={(e) => setFilterBy(e.target.value as "user" | "client")}
              className="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="user">User</option>
              <option value="client">Client</option>
            </select>
          </div>
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              {filterBy === "user" ? "User ID" : "Client ID"}
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={filterId}
                onChange={(e) => setFilterId(e.target.value)}
                placeholder={`Enter ${filterBy} ID...`}
                className="flex-1 px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                onClick={loadRoles}
                disabled={!filterId || loading}
                className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 disabled:opacity-50 transition"
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
          className="mb-6 p-4 border rounded bg-gray-50"
        >
          <h3 className="text-lg font-semibold mb-4">Assign New Role</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                User ID *
              </label>
              <input
                type="text"
                required
                value={formData.userId}
                onChange={(e) =>
                  setFormData({ ...formData, userId: e.target.value })
                }
                className="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="user-uuid"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Client ID *
              </label>
              <input
                type="text"
                required
                value={formData.clientId}
                onChange={(e) =>
                  setFormData({ ...formData, clientId: e.target.value })
                }
                className="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="client-id"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Role *
              </label>
              <select
                required
                value={formData.role}
                onChange={(e) =>
                  setFormData({ ...formData, role: e.target.value })
                }
                className="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
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
              className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 transition"
            >
              {loading ? "Assigning..." : "Assign Role"}
            </button>
          </div>
        </form>
      )}

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

      {/* Roles Table */}
      {loading && !roles.length ? (
        <div className="text-center py-8 text-gray-500">Loading roles...</div>
      ) : roles.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          No roles found.{" "}
          {filterId ? "Try a different search." : "Enter a filter to search."}
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  User ID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Client ID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Role
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Assigned At
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {roles.map((role) => (
                <tr key={role.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {role.user_email || role.user_id.slice(0, 8)}...
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {role.client_name || role.client_id}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800">
                      {role.role}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(role.assigned_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        role.is_active
                          ? "bg-green-100 text-green-800"
                          : "bg-red-100 text-red-800"
                      }`}
                    >
                      {role.is_active ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    {role.is_active && (
                      <button
                        onClick={() => handleRevokeRole(role)}
                        disabled={loading}
                        className="text-red-600 hover:text-red-900 disabled:opacity-50"
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

      <div className="mt-4 text-sm text-gray-500">
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
