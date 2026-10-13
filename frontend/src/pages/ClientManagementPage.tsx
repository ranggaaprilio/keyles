/**
 * Client Management Page
 *
 * Main page for managing OAuth client applications.
 * Orchestrates list, detail, create, edit flows with modal dialogs.
 */

import { useState, useCallback } from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { ClientList } from "../components/clients/ClientList";
import { ClientDetail } from "../components/clients/ClientDetail";
import { CreateClientForm } from "../components/clients/CreateClientForm";
import { EditClientForm } from "../components/clients/EditClientForm";
import { SecretDisplay } from "../components/clients/SecretDisplay";
import { RotateSecretDialog } from "../components/clients/RotateSecretDialog";
import { DeleteClientDialog } from "../components/clients/DeleteClientDialog";
import { useCreateClient } from "../hooks/useClients";
import type { CreateClientRequest } from "../types/client";

type PageView = "list" | "detail" | "create" | "edit";

export function ClientManagementPage() {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const { clientId: selectedClientId } = useParams<{ clientId: string }>();
  const createClient = useCreateClient();

  // Derive the view from the URL. Every /dashboard/clients* route renders this
  // same component, so React keeps the instance alive across navigations and
  // local view state would go stale (e.g. header back link, browser back).
  const view: PageView = pathname.endsWith("/new")
    ? "create"
    : selectedClientId && pathname.endsWith("/edit")
      ? "edit"
      : selectedClientId
        ? "detail"
        : "list";

  // Secret display state
  const [secretDialogOpen, setSecretDialogOpen] = useState(false);
  const [createdCredentials, setCreatedCredentials] = useState<{
    clientId: string;
    clientSecret: string | null;
    clientName: string;
  } | null>(null);

  // Rotate secret dialog
  const [rotateDialogOpen, setRotateDialogOpen] = useState(false);
  const [rotateClientId, setRotateClientId] = useState("");
  const [rotateClientName, setRotateClientName] = useState("");

  // Delete dialog
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteClientId, setDeleteClientId] = useState("");
  const [deleteClientName, setDeleteClientName] = useState("");

  // Navigation handlers
  const handleSelectClient = useCallback(
    (id: string) => {
      navigate(`/dashboard/clients/${id}`);
    },
    [navigate],
  );

  const handleBackToList = useCallback(() => {
    navigate("/dashboard/clients");
  }, [navigate]);

  const handleCreateNew = useCallback(() => {
    navigate("/dashboard/clients/new");
  }, [navigate]);

  const handleEdit = useCallback(
    (id: string) => {
      navigate(`/dashboard/clients/${id}/edit`);
    },
    [navigate],
  );

  // Create form submit — calls API, then shows secret dialog on success
  const handleCreateSubmit = useCallback(
    (data: CreateClientRequest) => {
      createClient.mutate(data, {
        onSuccess: (response) => {
          setCreatedCredentials({
            clientId: response.client_id,
            clientSecret: response.client_secret,
            clientName: response.client_name,
          });
          setSecretDialogOpen(true);
        },
      });
    },
    [createClient],
  );

  const handleSecretDialogClose = useCallback(() => {
    setSecretDialogOpen(false);
    setCreatedCredentials(null);
    handleBackToList();
  }, [handleBackToList]);

  // Rotate secret handlers
  const handleRotateSecret = useCallback(
    (_clientId: string, clientName: string) => {
      setRotateClientId(_clientId);
      setRotateClientName(clientName);
      setRotateDialogOpen(true);
    },
    [],
  );

  const handleSecretRotated = useCallback(
    (rotatedClientId: string, newSecret: string) => {
      setRotateDialogOpen(false);
      setCreatedCredentials({
        clientId: rotatedClientId,
        clientSecret: newSecret,
        clientName: rotateClientName,
      });
      setSecretDialogOpen(true);
    },
    [rotateClientName],
  );

  // Delete handlers
  const handleDelete = useCallback((_clientId: string, clientName: string) => {
    setDeleteClientId(_clientId);
    setDeleteClientName(clientName);
    setDeleteDialogOpen(true);
  }, []);

  const handleDeleted = useCallback(() => {
    setDeleteDialogOpen(false);
    handleBackToList();
  }, [handleBackToList]);

  return (
    <div className="max-w-6xl mx-auto px-4 py-6">
      {/* List View */}
      {view === "list" && (
        <ClientList
          onSelectClient={handleSelectClient}
          onCreateNew={handleCreateNew}
        />
      )}

      {/* Detail View */}
      {view === "detail" && selectedClientId && (
        <ClientDetail
          clientId={selectedClientId}
          onBack={handleBackToList}
          onEdit={handleEdit}
          onRotateSecret={handleRotateSecret}
          onDelete={handleDelete}
        />
      )}

      {/* Create View */}
      {view === "create" && (
        <CreateClientForm
          onSubmit={handleCreateSubmit}
          onCancel={handleBackToList}
          isLoading={createClient.isPending}
        />
      )}

      {/* Edit View */}
      {view === "edit" && selectedClientId && (
        <EditClientForm
          clientId={selectedClientId}
          onSuccess={handleBackToList}
          onCancel={() => navigate(`/dashboard/clients/${selectedClientId}`)}
        />
      )}

      {/* Secret Display Modal */}
      {createdCredentials && (
        <SecretDisplay
          clientId={createdCredentials.clientId}
          clientSecret={createdCredentials.clientSecret}
          clientName={createdCredentials.clientName}
          open={secretDialogOpen}
          onClose={handleSecretDialogClose}
        />
      )}

      {/* Rotate Secret Confirmation Dialog */}
      <RotateSecretDialog
        clientId={rotateClientId}
        clientName={rotateClientName}
        open={rotateDialogOpen}
        onClose={() => setRotateDialogOpen(false)}
        onSecretRotated={handleSecretRotated}
      />

      {/* Delete Confirmation Dialog */}
      <DeleteClientDialog
        clientId={deleteClientId}
        clientName={deleteClientName}
        open={deleteDialogOpen}
        onClose={() => setDeleteDialogOpen(false)}
        onDeleted={handleDeleted}
      />
    </div>
  );
}
