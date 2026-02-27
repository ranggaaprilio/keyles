import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ArrowLeft, Copy, Check, Pencil, RotateCw, Trash2 } from "lucide-react";
import { useState } from "react";
import { useClient } from "@/hooks/useClients";

interface ClientDetailProps {
  clientId: string;
  onBack: () => void;
  onEdit: (clientId: string) => void;
  onRotateSecret: (clientId: string, clientName: string) => void;
  onDelete: (clientId: string, clientName: string) => void;
}

export function ClientDetail({
  clientId,
  onBack,
  onEdit,
  onRotateSecret,
  onDelete,
}: ClientDetailProps) {
  const { data: client, isLoading, isError } = useClient(clientId);
  const [copiedId, setCopiedId] = useState(false);

  const copyId = async () => {
    if (!client) return;
    await navigator.clipboard.writeText(client.client_id);
    setCopiedId(true);
    setTimeout(() => setCopiedId(false), 2000);
  };

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-64 w-full rounded-lg" />
      </div>
    );
  }

  if (isError || !client) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <ArrowLeft className="h-4 w-4 mr-1" /> Back
        </Button>
        <div className="text-center py-12 text-destructive">
          Client not found or failed to load.
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={onBack}>
            <ArrowLeft className="h-4 w-4 mr-1" /> Back
          </Button>
          <h2 className="text-xl font-semibold">{client.client_name}</h2>
          <Badge
            variant={
              client.client_type === "confidential" ? "default" : "secondary"
            }
          >
            {client.client_type}
          </Badge>
          <Badge variant={client.is_active ? "default" : "destructive"}>
            {client.is_active ? "Active" : "Inactive"}
          </Badge>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => onEdit(clientId)}>
            <Pencil className="h-4 w-4 mr-1" /> Edit
          </Button>
          {client.client_type === "confidential" && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => onRotateSecret(clientId, client.client_name)}
            >
              <RotateCw className="h-4 w-4 mr-1" /> Rotate Secret
            </Button>
          )}
          <Button
            variant="destructive"
            size="sm"
            onClick={() => onDelete(clientId, client.client_name)}
          >
            <Trash2 className="h-4 w-4 mr-1" /> Delete
          </Button>
        </div>
      </div>

      {/* Client Info */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Client Information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Client ID */}
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-1">
              Client ID
            </p>
            <div className="flex items-center gap-2">
              <code className="font-mono text-sm bg-muted px-2 py-1 rounded">
                {client.client_id}
              </code>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={copyId}
              >
                {copiedId ? (
                  <Check className="h-3 w-3 text-green-600" />
                ) : (
                  <Copy className="h-3 w-3" />
                )}
              </Button>
            </div>
          </div>

          {/* Description */}
          {client.description && (
            <div>
              <p className="text-sm font-medium text-muted-foreground mb-1">
                Description
              </p>
              <p className="text-sm">{client.description}</p>
            </div>
          )}

          {/* Redirect URIs */}
          <div>
            <p className="text-sm font-medium text-muted-foreground mb-1">
              Redirect URIs
            </p>
            <ul className="space-y-1">
              {client.redirect_uris.map((uri, i) => (
                <li
                  key={i}
                  className="font-mono text-sm bg-muted px-2 py-1 rounded"
                >
                  {uri}
                </li>
              ))}
            </ul>
          </div>

          {/* Timestamps */}
          <div className="grid grid-cols-2 gap-4 pt-2 border-t">
            <div>
              <p className="text-sm font-medium text-muted-foreground">
                Created
              </p>
              <p className="text-sm">
                {new Date(client.created_at).toLocaleString()}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">
                Last Updated
              </p>
              <p className="text-sm">
                {new Date(client.updated_at).toLocaleString()}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
