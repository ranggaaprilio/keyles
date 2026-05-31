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
    <div className="p-6 bg-white shadow-[2px_2px_0_#000] border border-black">
      {/* Eyebrow */}
      <div className="bg-[#8c9ae0] px-4 py-2 -mt-6 -mx-6 mb-6 border-b border-black">
        <h2 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-white text-lg">
          Role Management
        </h2>
      </div>

      <div className="flex justify-end mb-6">
        <button
          onClick={() => setShowAssignForm(!showAssignForm)}
          className="px-4 py-2 border border-black bg-black text-white text-[12px] font-bold uppercase tracking-[1.5px] font-[Helvetica,Arial,system-ui,sans-serif] hover:bg-gray-800 transition-colors"
        >
          {showAssignForm ? "Cancel" : "Assign Role"}
        </button>
      </div>

      {/* Filter Section */}
      <div className="mb-6 p-4 bg-gray-100 border border-black">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] mb-2">
              Filter By
            </label>
            <select
              value={filterBy}
              onChange={(e) => setFilterBy(e.target.value as "user" | "client")}
              className="w-full px-3 py-2 border border-black bg-white font-['Times_New_Roman',Times,serif] text-sm focus:outline-none focus:ring-2 focus:ring-dell-red"
            >
              <option value="user">User</option>
              <option value="client">Client</option>
            </select>
          </div>
          <div className="md:col-span-2">
            <label className="block font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] mb-2">
              {filterBy === "user" ? "User ID" : "Client ID"}
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={filterId}
                onChange={(e) => setFilterId(e.target.value)}
                placeholder={`Enter ${filterBy} ID...`}
                className="flex-1 px-3 py-2 border border-black bg-white font-['Times_New_Roman',Times,serif] text-sm focus:outline-none focus:ring-2 focus:ring-dell-red"
              />
              <button
                onClick={loadRoles}
                disabled={!filterId || loading}
                className="px-4 py-2 border border-black bg-gray-700 text-white text-[12px] font-bold uppercase tracking-[1.5px] font-[Helvetica,Arial,system-ui,sans-serif] hover:bg-gray-600 disabled:opacity-50 transition-colors"
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
          className="mb-6 p-4 border border-black bg-gray-100"
        >
          <h3 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-sm mb-4">
            Assign New Role
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] mb-2">
                User ID *
              </label>
              <input
                type="text"
                required
                value={formData.userId}
                onChange={(e) =>
                  setFormData({ ...formData, userId: e.target.value })
                }
                className="w-full px-3 py-2 border border-black bg-white font-['Times_New_Roman',Times,serif] text-sm focus:outline-none focus:ring-2 focus:ring-dell-red"
                placeholder="user-uuid"
              />
            </div>
            <div>
              <label className="block font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] mb-2">
                Client ID *
              </label>
              <input
                type="text"
                required
                value={formData.clientId}
                onChange={(e) =>
                  setFormData({ ...formData, clientId: e.target.value })
                }
                className="w-full px-3 py-2 border border-black bg-white font-['Times_New_Roman',Times,serif] text-sm focus:outline-none focus:ring-2 focus:ring-dell-red"
                placeholder="client-id"
              />
            </div>
            <div>
              <label className="block font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] mb-2">
                Role *
              </label>
              <select
                required
                value={formData.role}
                onChange={(e) =>
                  setFormData({ ...formData, role: e.target.value })
                }
                className="w-full px-3 py-2 border border-black bg-white font-['Times_New_Roman',Times,serif] text-sm focus:outline-none focus:ring-2 focus:ring-dell-red"
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
              className="px-6 py-2 border border-black bg-black text-white text-[12px] font-bold uppercase tracking-[1.5px] font-[Helvetica,Arial,system-ui,sans-serif] hover:bg-gray-800 disabled:opacity-50 transition-colors"
            >
              {loading ? "Assigning..." : "Assign Role"}
            </button>
          </div>
        </form>
      )}

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

      {/* Roles Table */}
      {loading && !roles.length ? (
        <div className="text-center py-8 text-gray-500 font-['Times_New_Roman',Times,serif] text-sm">Loading roles...</div>
      ) : roles.length === 0 ? (
        <div className="text-center py-8 text-gray-500 font-['Times_New_Roman',Times,serif] text-sm">
          No roles found.{" "}
          {filterId ? "Try a different search." : "Enter a filter to search."}
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full border border-black">
            <thead className="bg-gray-100">
              <tr className="border-b border-black">
                <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                  User ID
                </th>
                <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                  Client ID
                </th>
                <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                  Role
                </th>
                <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                  Assigned At
                </th>
                <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                  Status
                </th>
                <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {roles.map((role) => (
                <tr key={role.id} className="border-b border-black hover:bg-gray-100">
                  <td className="px-6 py-4 whitespace-nowrap font-['Times_New_Roman',Times,serif] text-sm">
                    {role.user_email || role.user_id.slice(0, 8)}...
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap font-['Times_New_Roman',Times,serif] text-sm">
                    {role.client_name || role.client_id}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="px-2 py-1 text-xs font-bold border border-black bg-[#8c9ae0] text-white">
                      {role.role}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap font-['Times_New_Roman',Times,serif] text-sm text-gray-600">
                    {new Date(role.assigned_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 py-1 text-xs font-bold border border-black ${
                        role.is_active
                          ? "bg-green-700 text-white"
                          : "bg-red-700 text-white"
                      }`}
                    >
                      {role.is_active ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap font-['Times_New_Roman',Times,serif] text-sm">
                    {role.is_active && (
                      <button
                        onClick={() => handleRevokeRole(role)}
                        disabled={loading}
                        className="text-red-700 hover:text-red-900 underline disabled:opacity-50 font-['Times_New_Roman',Times,serif]"
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

      <div className="mt-4 font-['Times_New_Roman',Times,serif] text-sm text-gray-600">
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