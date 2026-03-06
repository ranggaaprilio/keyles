/**
 * EnableDisableDialog — confirmation dialog for toggling user status
 */

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { useUpdateUserStatus } from "../../hooks/useUsers";
import type { UserStatus } from "../../types/user";

interface EnableDisableDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  userId: string;
  email: string;
  currentStatus: UserStatus;
  onSuccess: () => void;
}

export function EnableDisableDialog({
  open,
  onOpenChange,
  userId,
  email,
  currentStatus,
  onSuccess,
}: EnableDisableDialogProps) {
  const mutation = useUpdateUserStatus();
  const isDisabling = currentStatus === "active";
  const targetStatus: UserStatus = isDisabling ? "disabled" : "active";

  const handleConfirm = async () => {
    await mutation.mutateAsync({ id: userId, status: targetStatus });
    onOpenChange(false);
    onSuccess();
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {isDisabling ? "Disable Account" : "Enable Account"}
          </DialogTitle>
        </DialogHeader>
        <p className="text-sm text-gray-600">
          {isDisabling
            ? `This will immediately terminate all active sessions for ${email}.`
            : `This will restore ${email}'s ability to authenticate.`}
        </p>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant={isDisabling ? "destructive" : "default"}
            disabled={mutation.isPending}
            onClick={handleConfirm}
          >
            {mutation.isPending
              ? isDisabling
                ? "Disabling..."
                : "Enabling..."
              : isDisabling
                ? "Disable Account"
                : "Enable Account"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
