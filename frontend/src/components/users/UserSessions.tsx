/**
 * UserSessions — active sessions table with termination
 */

import { useState } from "react";
import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import { useUserSessions, useRevokeSession } from "../../hooks/useSessions";

interface UserSessionsProps {
  userId: string;
}

function formatRelative(date: string | null | undefined): string {
  if (!date) return "—";
  const diff = Date.now() - new Date(date).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "Just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export function UserSessions({ userId }: UserSessionsProps) {
  const { data: sessions, isLoading } = useUserSessions(userId);
  const revokeSession = useRevokeSession();
  const [revoking, setRevoking] = useState<number | null>(null);

  const handleRevoke = async (sessionId: number) => {
    setRevoking(sessionId);
    try {
      await revokeSession.mutateAsync({ userId, sessionId: String(sessionId) });
    } finally {
      setRevoking(null);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-10 w-full" />
        ))}
      </div>
    );
  }

  const list = sessions ?? [];

  if (list.length === 0) {
    return (
      <p className="text-sm text-gray-500 py-4">
        No active sessions for this user.
      </p>
    );
  }

  return (
    <div className="bg-white rounded-lg border shadow-sm overflow-hidden">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Client
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Created
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Last Activity
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Expires
            </th>
            <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider" />
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {list.map((s) => (
            <tr key={s.id}>
              <td className="px-6 py-3 text-sm">
                {s.client_name ?? s.client_id}
              </td>
              <td className="px-6 py-3 text-sm text-gray-500">
                {formatRelative(s.created_at)}
              </td>
              <td className="px-6 py-3 text-sm text-gray-500">
                {formatRelative(s.last_used_at)}
              </td>
              <td className="px-6 py-3 text-sm text-gray-500">
                {new Date(s.expires_at).toLocaleDateString()}
              </td>
              <td className="px-6 py-3 text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-red-600 hover:text-red-700"
                  disabled={revoking === s.id}
                  onClick={() => handleRevoke(s.id)}
                >
                  Terminate
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
