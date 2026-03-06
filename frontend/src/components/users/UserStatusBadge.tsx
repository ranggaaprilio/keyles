/**
 * UserStatusBadge — colour-coded badge for user status
 */

import { Badge } from "../ui/badge";
import type { UserStatus } from "../../types/user";

const variants: Record<UserStatus, { className: string; label: string }> = {
  active: {
    className: "bg-green-100 text-green-800 hover:bg-green-100",
    label: "Active",
  },
  pending: {
    className: "bg-yellow-100 text-yellow-800 hover:bg-yellow-100",
    label: "Pending",
  },
  disabled: {
    className: "bg-red-100 text-red-800 hover:bg-red-100",
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
