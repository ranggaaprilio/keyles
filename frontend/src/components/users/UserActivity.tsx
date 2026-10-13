/**
 * UserActivity — paginated activity log
 */

import { useState } from "react";
import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import { useUserActivity } from "../../hooks/useSessions";
import { ChevronLeft, ChevronRight } from "lucide-react";

interface UserActivityProps {
  userId: string;
}

const eventColors: Record<string, string> = {
  login_failure: "text-destructive bg-destructive/5",
  session_terminated: "text-warning bg-warning/10",
  account_disabled: "text-destructive bg-destructive/5",
  role_revoked: "text-warning bg-warning/10",
  user_deleted: "text-destructive bg-destructive/5",
};

function eventLabel(type: string): string {
  return type
    .split("_")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}

export function UserActivity({ userId }: UserActivityProps) {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useUserActivity(userId, page);

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-10 w-full" />
        ))}
      </div>
    );
  }

  const events = data?.data ?? [];
  const totalPages = data?.total_pages ?? 1;

  if (events.length === 0 && page === 1) {
    return (
      <p className="text-sm text-muted-foreground py-4">
        No activity recorded for this user.
      </p>
    );
  }

  return (
    <div>
      <div className="rounded-lg border border-border bg-background shadow-card overflow-hidden">
        <table className="min-w-full divide-y divide-border">
          <thead className="bg-muted border-b border-border">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-foreground">
                Event
              </th>
              <th className="px-6 py-3 text-left text-sm font-medium text-foreground">
                Client
              </th>
              <th className="px-6 py-3 text-left text-sm font-medium text-foreground">
                IP / Location
              </th>
              <th className="px-6 py-3 text-left text-sm font-medium text-foreground">
                Time
              </th>
            </tr>
          </thead>
          <tbody className="bg-background divide-y divide-border">
            {events.map((e) => {
              const color =
                eventColors[e.event_type] ?? "text-muted-foreground bg-background";
              return (
                <tr key={e.id}>
                  <td className="px-6 py-3">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded-md border border-border text-xs font-medium ${color}`}
                    >
                      {eventLabel(e.event_type)}
                    </span>
                  </td>
                  <td className="px-6 py-3 text-sm text-muted-foreground">
                    {e.client_name ?? e.client_id ?? "—"}
                  </td>
                  <td className="px-6 py-3 text-sm text-muted-foreground">
                    {e.ip_address ?? "—"}
                    {e.country_code ? ` (${e.country_code})` : ""}
                  </td>
                  <td
                    className="px-6 py-3 text-sm text-muted-foreground"
                    title={new Date(e.occurred_at).toLocaleString()}
                  >
                    {new Date(e.occurred_at).toLocaleString()}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between mt-4">
          <p className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
            >
              <ChevronLeft className="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= totalPages}
              onClick={() => setPage(page + 1)}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
