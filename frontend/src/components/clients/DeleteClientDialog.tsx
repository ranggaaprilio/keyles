import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Loader2 } from "lucide-react";
import { useDeleteClient } from "@/hooks/useClients";

interface DeleteClientDialogProps {
  open: boolean;
  onClose: () => void;
  clientId: string;
  clientName: string;
  onDeleted: () => void;
}

export function DeleteClientDialog({
  open,
  onClose,
  clientId,
  clientName,
  onDeleted,
}: DeleteClientDialogProps) {
  const mutation = useDeleteClient();

  const handleConfirm = () => {
    mutation.mutate(clientId, {
      onSuccess: () => {
        onClose();
        onDeleted();
      },
    });
  };

  return (
    <AlertDialog open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Client Application</AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            <p>
              Are you sure you want to delete <strong>{clientName}</strong>?
            </p>
            <p className="text-red-700 font-medium font-['Times_New_Roman',Times,serif]">
              This action is irreversible. All associated tokens will be
              immediately revoked and any applications using this client will
              stop working.
            </p>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={mutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleConfirm}
            disabled={mutation.isPending}
            className="bg-red-700 text-white hover:bg-red-800 border border-red-700"
          >
            {mutation.isPending && (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            )}
            Delete Application
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
