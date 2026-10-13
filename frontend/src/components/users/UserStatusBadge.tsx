/**
 * UserStatusBadge — colour-coded badge for user status
 */

import { Badge } from "../ui/badge";
import type { UserStatus } from "../../types/user";

const variants: Record<UserStatus, { className: string; label: string }> = {
  active: {
    className: "bg-success/10 text-success border-transparent",
    label: "Active",
  },
  pending: {
    className: "bg-warning/10 text-warning border-transparent",
    label: "Pending",
  },
  disabled: {
    className: "bg-destructive/10 text-destructive border-transparent",
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
