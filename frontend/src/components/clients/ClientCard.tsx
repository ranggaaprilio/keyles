import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Copy, Check } from "lucide-react";
import { useState } from "react";
import type { Client } from "@/types/client";

interface ClientCardProps {
  client: Client;
  onClick: (clientId: string) => void;
}

export function ClientCard({ client, onClick }: ClientCardProps) {
  const [copied, setCopied] = useState(false);

  const copyClientId = async (e: React.MouseEvent) => {
    e.stopPropagation();
    await navigator.clipboard.writeText(client.client_id);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const truncatedId =
    client.client_id.length > 20
      ? `${client.client_id.slice(0, 20)}...`
      : client.client_id;

  return (
    <Card
      className="cursor-pointer transition-colors hover:bg-accent/50"
      onClick={() => onClick(client.client_id)}
    >
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="font-semibold truncate">{client.client_name}</h3>
              <Badge
                variant={
                  client.client_type === "confidential"
                    ? "default"
                    : "secondary"
                }
              >
                {client.client_type}
              </Badge>
            </div>

            <div className="flex items-center gap-1 text-sm text-muted-foreground font-mono">
              <span className="truncate">{truncatedId}</span>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 shrink-0"
                onClick={copyClientId}
              >
                {copied ? (
                  <Check className="h-3 w-3 text-green-600" />
                ) : (
                  <Copy className="h-3 w-3" />
                )}
              </Button>
            </div>

            {client.description && (
              <p className="text-sm text-muted-foreground mt-1 line-clamp-1">
                {client.description}
              </p>
            )}
          </div>

          <div className="flex flex-col items-end gap-1 shrink-0">
            <Badge variant={client.is_active ? "default" : "destructive"}>
              {client.is_active ? "Active" : "Inactive"}
            </Badge>
            <span className="text-xs text-muted-foreground">
              {new Date(client.created_at).toLocaleDateString()}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
