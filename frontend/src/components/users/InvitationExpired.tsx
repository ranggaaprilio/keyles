/**
 * InvitationExpired
 */

import { AlertCircle } from "lucide-react";
import { AuthShell } from "../layout/AuthShell";

export function InvitationExpired() {
  return (
    <AuthShell
      title="Invitation expired"
      description="This invitation link has expired or has already been used."
    >
      <div className="flex items-start gap-3 rounded-md bg-muted p-4">
        <AlertCircle className="h-5 w-5 shrink-0 text-destructive" />
        <p className="text-sm text-muted-foreground">
          Please contact your administrator to request a new invitation.
        </p>
      </div>
    </AuthShell>
  );
}
