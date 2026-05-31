import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Copy, Check, AlertTriangle } from "lucide-react";

interface SecretDisplayProps {
  open: boolean;
  onClose: () => void;
  clientId: string;
  clientSecret: string | null;
  clientName: string;
}

export function SecretDisplay({
  open,
  onClose,
  clientId,
  clientSecret,
  clientName,
}: SecretDisplayProps) {
  const [copiedId, setCopiedId] = useState(false);
  const [copiedSecret, setCopiedSecret] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  const copyToClipboard = async (text: string, field: "id" | "secret") => {
    await navigator.clipboard.writeText(text);
    if (field === "id") {
      setCopiedId(true);
      setTimeout(() => setCopiedId(false), 2000);
    } else {
      setCopiedSecret(true);
      setTimeout(() => setCopiedSecret(false), 2000);
    }
  };

  const handleClose = () => {
    setConfirmed(false);
    setCopiedId(false);
    setCopiedSecret(false);
    onClose();
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => !isOpen && confirmed && handleClose()}
    >
      <DialogContent
        className="sm:max-w-lg"
        onInteractOutside={(e) => e.preventDefault()}
      >
        <DialogHeader>
          <DialogTitle>Client Credentials — {clientName}</DialogTitle>
          <DialogDescription>
            Save these credentials now. They cannot be retrieved later.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          {/* Warning */}
          <div className="flex items-start gap-3 border border-black bg-yellow-100 p-3">
            <AlertTriangle className="h-5 w-5 text-yellow-700 shrink-0 mt-0.5" />
            <div className="text-sm text-yellow-900 font-['Times_New_Roman',Times,serif]">
              <p className="font-bold font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] uppercase tracking-[1px]">
                Important
              </p>
              {clientSecret ? (
                <p>
                  The client secret is shown only once. Copy and store it
                  securely before closing this dialog.
                </p>
              ) : (
                <p>
                  This is a public client — no client secret is generated. Use
                  PKCE for authorization flows.
                </p>
              )}
            </div>
          </div>

          {/* Client ID */}
          <div className="space-y-1.5">
            <Label className="text-sm font-medium">Client ID</Label>
            <div className="flex gap-2">
              <Input
                value={clientId}
                readOnly
                className="font-mono text-sm bg-gray-100"
              />
              <Button
                variant="outline"
                size="icon"
                onClick={() => copyToClipboard(clientId, "id")}
              >
                {copiedId ? (
                  <Check className="h-4 w-4 text-green-700" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            </div>
          </div>

          {/* Client Secret (confidential only) */}
          {clientSecret && (
            <div className="space-y-1.5">
              <Label className="text-sm font-medium">Client Secret</Label>
              <div className="flex gap-2">
                <Input
                  value={clientSecret}
                  readOnly
                  className="font-mono text-sm bg-gray-100"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => copyToClipboard(clientSecret, "secret")}
                >
                  {copiedSecret ? (
                    <Check className="h-4 w-4 text-green-700" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
          )}

          {/* Confirmation */}
          <div className="flex items-center gap-2 pt-2">
            <input
              type="checkbox"
              id="confirm-saved"
              checked={confirmed}
              onChange={(e) => setConfirmed(e.target.checked)}
              className="h-4 w-4 border border-black"
            />
            <label
              htmlFor="confirm-saved"
              className="text-sm cursor-pointer font-['Times_New_Roman',Times,serif]"
            >
              I have saved{" "}
              {clientSecret ? "the client secret" : "the client ID"} securely
            </label>
          </div>
        </div>

        <div className="flex justify-end">
          <Button onClick={handleClose} disabled={!confirmed}>
            Done
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
