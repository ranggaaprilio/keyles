/**
 * UserStatusBadge — colour-coded badge for user status
 */

import { Badge } from "../ui/badge";
import type { UserStatus } from "../../types/user";

const variants: Record<UserStatus, { className: string; label: string }> = {
  active: {
    className: "bg-success text-primary-foreground rounded-md border border-border hover:bg-success/90",
    label: "Active",
  },
  pending: {
    className: "bg-warning text-primary-foreground rounded-md border border-border hover:bg-warning/90",
    label: "Pending",
  },
  disabled: {
    className: "bg-destructive text-primary-foreground rounded-md border border-border hover:bg-destructive/90",
    label: "Disabled",
  },
};

interface UserStatusBadgeProps {
  status: UserStatus;
}

export function UserStatusBadge({ status }: UserStatusBadgeProps) {
  const v = variants[status] ?? variants.pending;
  return (
    <Badge variant="outline" className={v.className}>
      {v.label}
    </Badge>
  );
}
