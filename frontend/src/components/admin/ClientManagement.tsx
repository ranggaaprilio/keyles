/**
 * ClientManagement Component
 *
 * Displays a table of OAuth clients with CRUD operations.
 * Supports create, edit, delete, and secret rotation.
 */

import { useState, useEffect, useCallback } from "react";
import {
  Plus,
  Edit2,
  Trash2,
  Key,
  RefreshCw,
  Copy,
  Check,
  AlertCircle,
  X,
} from "lucide-react";
import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { clientService } from "../../services/clientService";
import { ClientForm } from "./ClientForm";
import type {
  Client,
  CreateClientResponse,
  RotateSecretResponse,
} from "../../types/client";

interface ClientManagementProps {
  tenantId?: string;
}

type ViewMode = "list" | "create" | "edit";

interface CredentialsModal {
  show: boolean;
  clientId: string;
  clientSecret: string | null;
  isRotation: boolean;
}

export function ClientManagement(_props: ClientManagementProps) {
  const [clients, setClients] = useState<Client[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>("list");
  const [editingClient, setEditingClient] = useState<Client | null>(null);
  const [deletingClientId, setDeletingClientId] = useState<string | null>(null);
  const [rotatingClientId, setRotatingClientId] = useState<string | null>(null);
  const [selectedClientIds, setSelectedClientIds] = useState<Set<string>>(
    new Set()
  );
  const [bulkAction, setBulkAction] = useState<string | null>(null);
  const [credentials, setCredentials] = useState<CredentialsModal>({
    show: false,
    clientId: "",
    clientSecret: "",
    isRotation: false,
  });
  const [copiedField, setCopiedField] = useState<string | null>(null);

  const fetchClients = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await clientService.list();
      setClients(response.clients || []);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to load clients");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchClients();
  }, [fetchClients]);

  const handleCreateSuccess = (response: CreateClientResponse | Client) => {
    // Type guard to check if it's a CreateClientResponse (has client_secret)
    if ("client_secret" in response) {
      setCredentials({
        show: true,
        clientId: response.client_id,
        clientSecret: response.client_secret,
        isRotation: false,
      });
    }
    setViewMode("list");
    fetchClients();
  };

  const handleUpdateSuccess = (response: CreateClientResponse | Client) => {
    // For updates, we just refresh the list - no credentials to show
    void response; // Suppress unused variable warning
    setViewMode("list");
    setEditingClient(null);
    fetchClients();
  };

  const handleEdit = (client: Client) => {
    setEditingClient(client);
    setViewMode("edit");
  };

  const handleDelete = async (clientId: string) => {
    try {
      setDeletingClientId(clientId);
      await clientService.delete(clientId);
      fetchClients();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to delete client");
    } finally {
      setDeletingClientId(null);
    }
  };

  const handleRotateSecret = async (clientId: string) => {
    try {
      setRotatingClientId(clientId);
      const response: RotateSecretResponse =
        await clientService.rotateSecret(clientId);
      setCredentials({
        show: true,
        clientId: response.client_id,
        clientSecret: response.client_secret,
        isRotation: true,
      });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to rotate secret");
    } finally {
      setRotatingClientId(null);
    }
  };

  const handleCopy = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch {
      // Clipboard API not available
    }
  };

  // Bulk operations
  const handleSelectAll = () => {
    if (selectedClientIds.size === clients.length) {
      setSelectedClientIds(new Set());
    } else {
      setSelectedClientIds(new Set(clients.map((c) => c.client_id)));
    }
  };

  const handleSelectClient = (clientId: string) => {
    const newSelected = new Set(selectedClientIds);
    if (newSelected.has(clientId)) {
      newSelected.delete(clientId);
    } else {
      newSelected.add(clientId);
    }
    setSelectedClientIds(newSelected);
  };

  const handleBulkActivate = async () => {
    try {
      setBulkAction("activating");
      for (const clientId of selectedClientIds) {
        const client = clients.find((c) => c.client_id === clientId);
        if (client && !client.is_active) {
          await clientService.update(clientId, {
            client_name: client.client_name,
            redirect_uris: client.redirect_uris,
            is_active: true,
          });
        }
      }
      setSelectedClientIds(new Set());
      fetchClients();
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : "Failed to activate clients"
      );
    } finally {
      setBulkAction(null);
    }
  };

  const handleBulkDeactivate = async () => {
    try {
      setBulkAction("deactivating");
      for (const clientId of selectedClientIds) {
        const client = clients.find((c) => c.client_id === clientId);
        if (client && client.is_active) {
          await clientService.delete(clientId);
        }
      }
      setSelectedClientIds(new Set());
      fetchClients();
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : "Failed to deactivate clients"
      );
    } finally {
      setBulkAction(null);
    }
  };

  const closeCredentialsModal = () => {
    setCredentials({
      show: false,
      clientId: "",
      clientSecret: "",
      isRotation: false,
    });
  };

  const handleCancel = () => {
    setViewMode("list");
    setEditingClient(null);
  };

  // Render credentials modal
  const renderCredentialsModal = () => {
    if (!credentials.show) return null;

    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80">
        <div className="bg-white border-8 border-black p-6 shadow-[4px_4px_0_#000] max-w-md w-full mx-4">
          <div className="flex items-start justify-between mb-4">
            <div className="flex items-center gap-2">
              <div className="w-10 h-10 border border-black bg-green-100 flex items-center justify-center">
                <Key className="w-5 h-5 text-green-700" />
              </div>
              <div>
                <h3 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-lg text-gray-900">
                  {credentials.isRotation ? "Secret Rotated" : "Client Created"}
                </h3>
                <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-500">
                  Save these credentials securely
                </p>
              </div>
            </div>
            <button
              onClick={closeCredentialsModal}
              className="text-gray-400 hover:text-black"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          <div className="bg-amber-50 border border-amber-700 p-3 mb-4">
            <div className="flex gap-2">
              <AlertCircle className="w-5 h-5 text-amber-700 flex-shrink-0 mt-0.5" />
              <p className="font-['Times_New_Roman',Times,serif] text-sm text-amber-900">
                <strong>Important:</strong> The client secret will only be shown
                once. Copy it now and store it securely.
              </p>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] mb-1">
                Client ID
              </label>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-3 py-2 bg-gray-100 border border-black font-mono text-sm break-all">
                  {credentials.clientId}
                </code>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => handleCopy(credentials.clientId, "clientId")}
                >
                  {copiedField === "clientId" ? (
                    <Check className="w-4 h-4 text-green-700" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </Button>
              </div>
            </div>

            <div>
              <label className="block font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] mb-1">
                Client Secret
              </label>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-3 py-2 bg-gray-100 border border-black font-mono text-sm break-all">
                  {credentials.clientSecret ?? "N/A (Public client)"}
                </code>
                {credentials.clientSecret && (
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() =>
                      handleCopy(credentials.clientSecret!, "clientSecret")
                    }
                  >
                    {copiedField === "clientSecret" ? (
                      <Check className="w-4 h-4 text-green-700" />
                    ) : (
                      <Copy className="w-4 h-4" />
                    )}
                  </Button>
                )}
              </div>
            </div>
          </div>

          <div className="mt-6">
            <Button onClick={closeCredentialsModal} className="w-full">
              I&apos;ve Saved the Credentials
            </Button>
          </div>
        </div>
      </div>
    );
  };

  // Render loading state
  if (loading && viewMode === "list") {
    return (
      <Card className="shadow-[2px_2px_0_#000]">
        <CardHeader>
          <CardTitle>OAuth Clients</CardTitle>
          <CardDescription>Loading clients...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-32">
            <RefreshCw className="w-6 h-6 animate-spin text-gray-400" />
          </div>
        </CardContent>
      </Card>
    );
  }

  // Render form view
  if (viewMode === "create" || viewMode === "edit") {
    return (
      <>
        <ClientForm
          client={editingClient}
          onSuccess={
            viewMode === "create" ? handleCreateSuccess : handleUpdateSuccess
          }
          onCancel={handleCancel}
        />
        {renderCredentialsModal()}
      </>
    );
  }

  // Render list view
  return (
    <>
      <Card className="shadow-[2px_2px_0_#000]">
        {/* Eyebrow */}
        <div className="bg-[#8e8a25] px-4 py-2 border-b border-black">
          <div className="flex flex-row items-center justify-between">
            <div>
              <CardTitle className="text-white">OAuth Clients</CardTitle>
              <CardDescription className="text-white/80">
                Manage client applications for OAuth authentication
              </CardDescription>
            </div>
            <Button onClick={() => setViewMode("create")}>
              <Plus className="w-4 h-4 mr-2" />
              New Client
            </Button>
          </div>
        </div>
        <CardContent className="font-['Times_New_Roman',Times,serif]">
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-700">
              <div className="flex items-center gap-2 text-red-800">
                <AlertCircle className="w-4 h-4" />
                <span className="text-sm">{error}</span>
              </div>
            </div>
          )}

          {clients.length === 0 ? (
            <div className="text-center py-12">
              <div className="w-12 h-12 mx-auto mb-4 border border-black bg-gray-100 flex items-center justify-center">
                <Key className="w-6 h-6 text-gray-400" />
              </div>
              <h3 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-lg text-gray-900 mb-1">
                No clients yet
              </h3>
              <p className="text-gray-500 mb-4 text-sm">
                Create your first OAuth client to get started
              </p>
              <Button onClick={() => setViewMode("create")}>
                <Plus className="w-4 h-4 mr-2" />
                Create Client
              </Button>
            </div>
          ) : (
            <>
              {selectedClientIds.size > 0 && (
                <div className="mb-4 p-4 bg-blue-50 border border-blue-700">
                  <div className="flex items-center justify-between">
                    <span className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-blue-900">
                      {selectedClientIds.size} client(s) selected
                    </span>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={handleBulkActivate}
                        disabled={bulkAction === "activating"}
                      >
                        {bulkAction === "activating"
                          ? "Activating..."
                          : "Activate"}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={handleBulkDeactivate}
                        disabled={bulkAction === "deactivating"}
                      >
                        {bulkAction === "deactivating"
                          ? "Deactivating..."
                          : "Deactivate"}
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => setSelectedClientIds(new Set())}
                      >
                        Clear
                      </Button>
                    </div>
                  </div>
                </div>
              )}
              <div className="overflow-x-auto">
                <table className="w-full border border-black">
                  <thead>
                    <tr className="bg-gray-100 border-b border-black">
                      <th className="px-4 py-3 text-left">
                        <input
                          type="checkbox"
                          checked={
                            selectedClientIds.size === clients.length &&
                            clients.length > 0
                          }
                          onChange={handleSelectAll}
                          className="border border-black"
                        />
                      </th>
                      <th className="px-4 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                        Name
                      </th>
                      <th className="px-4 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                        Client ID
                      </th>
                      <th className="px-4 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                        Redirect URIs
                      </th>
                      <th className="px-4 py-3 text-left font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                        Status
                      </th>
                      <th className="px-4 py-3 text-right font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {clients.map((client) => (
                      <tr key={client.client_id} className="border-b border-black hover:bg-gray-100">
                        <td className="px-4 py-3">
                          <input
                            type="checkbox"
                            checked={selectedClientIds.has(client.client_id)}
                            onChange={() =>
                              handleSelectClient(client.client_id)
                            }
                            className="border border-black"
                          />
                        </td>
                        <td className="px-4 py-3">
                          <span className="font-semibold text-gray-900 text-sm">
                            {client.client_name}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <code className="text-sm text-gray-600 bg-gray-100 px-2 py-1 border border-black font-mono">
                            {client.client_id.substring(0, 12)}...
                          </code>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex flex-col gap-1">
                            {client.redirect_uris
                              .slice(0, 2)
                              .map((uri, idx) => (
                                <span
                                  key={idx}
                                  className="text-sm text-gray-600 truncate max-w-[200px]"
                                >
                                  {uri}
                                </span>
                              ))}
                            {client.redirect_uris.length > 2 && (
                              <span className="text-xs text-gray-400">
                                +{client.redirect_uris.length - 2} more
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex items-center px-2 py-1 text-xs font-bold border border-black ${
                              client.is_active
                                ? "bg-green-700 text-white"
                                : "bg-gray-400 text-white"
                            }`}
                          >
                            {client.is_active ? "Active" : "Inactive"}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center justify-end gap-2">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleEdit(client)}
                              title="Edit client"
                            >
                              <Edit2 className="w-4 h-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() =>
                                handleRotateSecret(client.client_id)
                              }
                              disabled={rotatingClientId === client.client_id}
                              title="Rotate secret"
                            >
                              {rotatingClientId === client.client_id ? (
                                <RefreshCw className="w-4 h-4 animate-spin" />
                              ) : (
                                <Key className="w-4 h-4" />
                              )}
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDelete(client.client_id)}
                              disabled={deletingClientId === client.client_id}
                              className="text-red-600 hover:text-red-700 hover:bg-red-50"
                              title="Delete client"
                            >
                              {deletingClientId === client.client_id ? (
                                <RefreshCw className="w-4 h-4 animate-spin" />
                              ) : (
                                <Trash2 className="w-4 h-4" />
                              )}
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          )}
        </CardContent>
      </Card>
      {renderCredentialsModal()}
    </>
  );
}

export default ClientManagement;