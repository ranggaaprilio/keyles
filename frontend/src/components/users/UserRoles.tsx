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
        <div className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black border border-black px-2 py-1 inline-block">
          Role Assignments
        </div>
        <Button size="sm" onClick={() => setAssignOpen(true)}>
          <Plus className="h-4 w-4 mr-1" />
          Assign Role
        </Button>
      </div>

      {activeRoles.length === 0 ? (
        <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-700 py-4">
          No roles assigned. Assign a role to grant access to a client
          application.
        </p>
      ) : (
        <div className="space-y-6">
          {Object.entries(grouped).map(([clientName, assignments]) => (
            <div key={clientName}>
              <h4 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black mb-2">
                {clientName}
              </h4>
              <div className="bg-white border border-black divide-y divide-black">
                {assignments.map((a) => (
                  <div
                    key={a.id}
                    className="flex items-center justify-between px-4 py-3"
                  >
                    <div>
                      <span className="font-['Times_New_Roman',Times,serif] text-sm font-medium">{a.role}</span>
                      <span className="font-['Times_New_Roman',Times,serif] text-xs text-gray-500 ml-3">
                        Granted {new Date(a.granted_at).toLocaleDateString()}
                      </span>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-red-600 hover:text-red-700 hover:bg-red-50"
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
