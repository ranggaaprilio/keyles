/**
 * UserList — paginated user table with search, filter, and invite
 */

import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import { UserStatusBadge } from "./UserStatusBadge";
import { UserListFilters } from "./UserListFilters";
import { InviteUserDialog } from "./InviteUserDialog";
import { useUsers } from "../../hooks/useUsers";
import type { UserListFilters as Filters } from "../../types/user";
import { UserPlus, ChevronLeft, ChevronRight } from "lucide-react";

function formatRelativeTime(date: string | null): string {
  if (!date) return "Never";
  const diff = Date.now() - new Date(date).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "Just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  return new Date(date).toLocaleDateString();
}

export function UserList() {
  const navigate = useNavigate();
  const [filters, setFilters] = useState<Filters>({ page: 1, page_size: 25 });
  const [inviteOpen, setInviteOpen] = useState(false);
  const { data, isLoading, error } = useUsers(filters);

  const handleFiltersChange = useCallback((f: Filters) => {
    setFilters(f);
  }, []);

  const users = data?.data ?? [];
  const totalPages = data?.total_pages ?? 1;
  const currentPage = filters.page ?? 1;

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">Users</h1>
            <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-700 mt-1">
              Manage users, roles, and access
            </p>
          </div>
          <Button onClick={() => setInviteOpen(true)}>
            <UserPlus className="h-4 w-4 mr-2" />
            Invite User
          </Button>
        </div>

        {/* Filters */}
        <UserListFilters
          filters={filters}
          onFiltersChange={handleFiltersChange}
        />

        {/* Content */}
        <div className="mt-6">
          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 10 }).map((_, i) => (
                <Skeleton key={i} className="h-14 w-full" />
              ))}
            </div>
          ) : error ? (
            <div className="text-center py-12">
              <p className="font-['Times_New_Roman',Times,serif] text-sm text-red-600">
                Failed to load users. Please try again.
              </p>
            </div>
          ) : users.length === 0 ? (
            <div className="text-center py-12">
              <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-700">No users found.</p>
              {filters.search || filters.status ? (
                <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-500 mt-1">
                  Try adjusting your filters.
                </p>
              ) : (
                <Button
                  variant="outline"
                  className="mt-4"
                  onClick={() => setInviteOpen(true)}
                >
                  <UserPlus className="h-4 w-4 mr-2" />
                  Invite your first user
                </Button>
              )}
            </div>
          ) : (
            <>
              {/* Table */}
              <div className="bg-white border border-black shadow-[2px_2px_0_#000] overflow-hidden">
                <table className="min-w-full divide-y divide-black">
                  <thead className="bg-gray-100 border-b border-black">
                    <tr>
                      <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
                        Name
                      </th>
                      <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
                        Email
                      </th>
                      <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
                        Last Login
                      </th>
                      <th className="px-6 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black">
                        Roles
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-black">
                    {users.map((user) => (
                      <tr
                        key={user.id}
                        className="hover:bg-gray-100 cursor-pointer"
                        onClick={() => navigate(`/dashboard/users/${user.id}`)}
                      >
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className="font-['Times_New_Roman',Times,serif] text-sm font-medium text-[#0000ee] hover:underline">
                            {user.display_name || user.email.split("@")[0]}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap font-['Times_New_Roman',Times,serif] text-sm text-gray-700">
                          {user.email}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <UserStatusBadge status={user.status} />
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap font-['Times_New_Roman',Times,serif] text-sm text-gray-700">
                          {formatRelativeTime(user.last_login_at)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap font-['Times_New_Roman',Times,serif] text-sm text-gray-700">
                          {user.role_count ?? 0}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4">
                  <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-700">
                    Page {currentPage} of {totalPages} ({data?.total ?? 0}{" "}
                    users)
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={currentPage <= 1}
                      onClick={() =>
                        setFilters({ ...filters, page: currentPage - 1 })
                      }
                    >
                      <ChevronLeft className="h-4 w-4" />
                      Previous
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={currentPage >= totalPages}
                      onClick={() =>
                        setFilters({ ...filters, page: currentPage + 1 })
                      }
                    >
                      Next
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>

        <InviteUserDialog open={inviteOpen} onOpenChange={setInviteOpen} />
      </div>
    </div>
  );
}
