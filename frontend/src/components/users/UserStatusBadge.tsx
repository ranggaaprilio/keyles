/**
 * UserStatusBadge — colour-coded badge for user status
 */

import { Badge } from "../ui/badge";
import type { UserStatus } from "../../types/user";

const variants: Record<UserStatus, { className: string; label: string }> = {
  active: {
    className: "bg-green-700 text-white border border-black hover:bg-green-700",
    label: "Active",
  },
  pending: {
    className: "bg-yellow-600 text-white border border-black hover:bg-yellow-600",
    label: "Pending",
  },
  disabled: {
    className: "bg-red-700 text-white border border-black hover:bg-red-700",
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
