/**
 * DeleteUserDialog — destructive confirmation dialog
 */

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "../ui/dialog";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Button } from "../ui/button";
import { useDeleteUser } from "../../hooks/useUsers";

interface DeleteUserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  userId: string;
  email: string;
  onDeleted: () => void;
}

export function DeleteUserDialog({
  open,
  onOpenChange,
  userId,
  email,
  onDeleted,
}: DeleteUserDialogProps) {
  const deleteUser = useDeleteUser();
  const [confirmation, setConfirmation] = useState("");

  const canConfirm = confirmation === email;

  const handleDelete = async () => {
    await deleteUser.mutateAsync(userId);
    setConfirmation("");
    onOpenChange(false);
    onDeleted();
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-red-600">Delete User</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <p className="text-sm text-gray-600">
            This will permanently delete <strong>{email}</strong> and all their
            role assignments. This action cannot be undone.
          </p>
          <div className="space-y-2">
            <Label htmlFor="confirm-email">
              Type <strong>{email}</strong> to confirm
            </Label>
            <Input
              id="confirm-email"
              value={confirmation}
              onChange={(e) => setConfirmation(e.target.value)}
              placeholder={email}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            disabled={!canConfirm || deleteUser.isPending}
            onClick={handleDelete}
          >
            {deleteUser.isPending ? "Deleting..." : "Delete User"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
