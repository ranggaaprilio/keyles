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
import { useRotateSecret } from "@/hooks/useClients";

interface RotateSecretDialogProps {
  open: boolean;
  onClose: () => void;
  clientId: string;
  clientName: string;
  onSecretRotated: (clientId: string, newSecret: string) => void;
}

export function RotateSecretDialog({
  open,
  onClose,
  clientId,
  clientName,
  onSecretRotated,
}: RotateSecretDialogProps) {
  const mutation = useRotateSecret();

  const handleConfirm = () => {
    mutation.mutate(clientId, {
      onSuccess: (data) => {
        onClose();
        onSecretRotated(data.client_id, data.client_secret);
      },
    });
  };

  return (
    <AlertDialog open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Rotate Client Secret</AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            <p>
              Are you sure you want to rotate the secret for{" "}
              <strong>{clientName}</strong>?
            </p>
            <p className="text-destructive font-medium">
              The current secret will stop working immediately. Any applications
              using the old secret will need to be updated.
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
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {mutation.isPending && (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            )}
            Rotate Secret
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
