/**
 * UserRoles — role assignments table grouped by client
 */

import { useState } from "react";
import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import { useUserRoles, useRevokeRole } from "../../hooks/useRoles";
import { AssignRoleDialog } from "./AssignRoleDialog";
import { Plus, X } from "lucide-react";
import type { RoleAssignment } from "../../types/user";

interface UserRolesProps {
  userId: string;
}

export function UserRoles({ userId }: UserRolesProps) {
  const { data: roles, isLoading } = useUserRoles(userId);
  const revokeRole = useRevokeRole();
  const [assignOpen, setAssignOpen] = useState(false);
  const [revoking, setRevoking] = useState<number | null>(null);

  const handleRevoke = async (assignment: RoleAssignment) => {
    setRevoking(assignment.id);
    try {
      await revokeRole.mutateAsync({ userId, assignmentId: assignment.id });
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

  const activeRoles = (roles ?? []).filter((r) => r.is_active);

  // Group by client
  const grouped = activeRoles.reduce<Record<string, RoleAssignment[]>>(
    (acc, r) => {
      const key = r.client_name ?? r.client_id;
      if (!acc[key]) acc[key] = [];
      acc[key].push(r);
      return acc;
    },
    {},
  );

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <div className="text-sm font-medium text-foreground rounded-md border border-border px-2 py-1 inline-block">
          Role Assignments
        </div>
        <Button size="sm" onClick={() => setAssignOpen(true)}>
          <Plus className="h-4 w-4 mr-1" />
          Assign Role
        </Button>
      </div>

      {activeRoles.length === 0 ? (
        <p className="text-sm text-muted-foreground py-4">
          No roles assigned. Assign a role to grant access to a client
          application.
        </p>
      ) : (
        <div className="space-y-6">
          {Object.entries(grouped).map(([clientName, assignments]) => (
            <div key={clientName}>
              <h4 className="text-sm font-medium text-foreground mb-2">
                {clientName}
              </h4>
              <div className="bg-background rounded-md border border-border divide-y divide-border">
                {assignments.map((a) => (
                  <div
                    key={a.id}
                    className="flex items-center justify-between px-4 py-3"
                  >
                    <div>
                      <span className="text-sm font-medium">{a.role}</span>
                      <span className="text-xs text-muted-foreground ml-3">
                        Granted {new Date(a.granted_at).toLocaleDateString()}
                      </span>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-destructive hover:text-destructive/80 hover:bg-destructive/5"
                      disabled={revoking === a.id}
                      onClick={() => handleRevoke(a)}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      <AssignRoleDialog
        open={assignOpen}
        onOpenChange={setAssignOpen}
        userId={userId}
      />
    </div>
  );
}
