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
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-sm font-medium text-foreground">Users</h1>
            <p className="text-sm text-muted-foreground mt-1">
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
              <p className="text-sm text-destructive">
                Failed to load users. Please try again.
              </p>
            </div>
          ) : users.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-sm text-muted-foreground">No users found.</p>
              {filters.search || filters.status ? (
                <p className="text-sm text-muted-foreground mt-1">
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
              <div className="rounded-lg border border-border bg-background shadow-card overflow-hidden">
                <table className="min-w-full divide-y divide-border">
                  <thead className="bg-muted border-b border-border">
                    <tr>
                      <th className="px-6 py-3 text-left text-sm font-medium text-foreground">
                        Name
                      </th>
                      <th className="px-6 py-3 text-left text-sm font-medium text-foreground">
                        Email
                      </th>
                      <th className="px-6 py-3 text-left text-sm font-medium text-foreground">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left text-sm font-medium text-foreground">
                        Last Login
                      </th>
                      <th className="px-6 py-3 text-left text-sm font-medium text-foreground">
                        Roles
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-background divide-y divide-border">
                    {users.map((user) => (
                      <tr
                        key={user.id}
                        className="hover:bg-muted cursor-pointer"
                        onClick={() => navigate(`/dashboard/users/${user.id}`)}
                      >
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className="text-sm font-medium text-foreground hover:underline underline-offset-4">
                            {user.display_name || user.email.split("@")[0]}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                          {user.email}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <UserStatusBadge status={user.status} />
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                          {formatRelativeTime(user.last_login_at)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
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
                  <p className="text-sm text-muted-foreground">
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
