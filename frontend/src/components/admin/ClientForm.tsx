/**
 * ClientForm Component
 *
 * Form for creating and editing OAuth clients.
 * Supports client name and redirect URIs with validation.
 */

import { useState, useEffect } from "react";
import { Plus, Trash2, AlertCircle, ArrowLeft, Save } from "lucide-react";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { clientService } from "../../services/clientService";
import type {
  Client,
  ClientType,
  CreateClientResponse,
  CreateClientRequest,
  UpdateClientRequest,
} from "../../types/client";

interface ClientFormProps {
  client?: Client | null;
  onSuccess: (response: CreateClientResponse | Client) => void;
  onCancel: () => void;
}

interface FormData {
  clientName: string;
  clientType: ClientType;
  redirectUris: string[];
  isActive: boolean;
}

interface FormErrors {
  clientName?: string;
  redirectUris?: string[];
  general?: string;
}

// Validate a redirect URI
const validateRedirectUri = (uri: string): string | null => {
  if (!uri.trim()) {
    return "Redirect URI cannot be empty";
  }

  try {
    const url = new URL(uri);

    // Allow localhost for development
    const isLocalhost =
      url.hostname === "localhost" || url.hostname === "127.0.0.1";

    // Require HTTPS for non-localhost URIs
    if (!isLocalhost && url.protocol !== "https:") {
      return "Redirect URI must use HTTPS (except for localhost)";
    }

    // Disallow fragments
    if (url.hash) {
      return "Redirect URI cannot contain fragments (#)";
    }

    return null;
  } catch {
    return "Invalid URL format";
  }
};

export function ClientForm({ client, onSuccess, onCancel }: ClientFormProps) {
  const isEditing = !!client;

  const [formData, setFormData] = useState<FormData>({
    clientName: "",
    clientType: "confidential",
    redirectUris: [""],
    isActive: true,
  });

  const [errors, setErrors] = useState<FormErrors>({});
  const [submitting, setSubmitting] = useState(false);

  // Initialize form with client data when editing
  useEffect(() => {
    if (client) {
      setFormData({
        clientName: client.client_name,
        clientType: client.client_type,
        redirectUris:
          client.redirect_uris.length > 0 ? client.redirect_uris : [""],
        isActive: client.is_active,
      });
    }
  }, [client]);

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    // Validate client name
    if (!formData.clientName.trim()) {
      newErrors.clientName = "Client name is required";
    } else if (formData.clientName.length < 3) {
      newErrors.clientName = "Client name must be at least 3 characters";
    } else if (formData.clientName.length > 100) {
      newErrors.clientName = "Client name must not exceed 100 characters";
    }

    // Validate redirect URIs
    const uriErrors: string[] = [];
    const nonEmptyUris = formData.redirectUris.filter((uri) => uri.trim());

    if (nonEmptyUris.length === 0) {
      newErrors.general = "At least one redirect URI is required";
    } else {
      formData.redirectUris.forEach((uri, index) => {
        if (uri.trim()) {
          const error = validateRedirectUri(uri);
          if (error) {
            uriErrors[index] = error;
          }
        }
      });

      if (uriErrors.length > 0) {
        newErrors.redirectUris = uriErrors;
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      setSubmitting(true);
      setErrors({});

      // Filter out empty URIs
      const redirectUris = formData.redirectUris.filter((uri) => uri.trim());

      if (isEditing && client) {
        // Update existing client
        const updateData: UpdateClientRequest = {
          client_name: formData.clientName,
          redirect_uris: redirectUris,
          is_active: formData.isActive,
        };

        const response = await clientService.update(
          client.client_id,
          updateData
        );
        onSuccess(response);
      } else {
        // Create new client
        const createData: CreateClientRequest = {
          client_name: formData.clientName,
          client_type: formData.clientType,
          redirect_uris: redirectUris,
        };

        const response = await clientService.create(createData);
        onSuccess(response);
      }
    } catch (err) {
      setErrors({
        general: err instanceof Error ? err.message : "Failed to save client",
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleClientNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData((prev) => ({ ...prev, clientName: e.target.value }));
    if (errors.clientName) {
      const { clientName: _, ...rest } = errors;
      setErrors(rest);
    }
  };

  const handleRedirectUriChange = (index: number, value: string) => {
    setFormData((prev) => {
      const newUris = [...prev.redirectUris];
      newUris[index] = value;
      return { ...prev, redirectUris: newUris };
    });

    if (errors.redirectUris?.[index]) {
      setErrors((prev) => {
        const newUriErrors = [...(prev.redirectUris || [])];
        delete newUriErrors[index];
        return { ...prev, redirectUris: newUriErrors };
      });
    }
  };

  const addRedirectUri = () => {
    setFormData((prev) => ({
      ...prev,
      redirectUris: [...prev.redirectUris, ""],
    }));
  };

  const removeRedirectUri = (index: number) => {
    if (formData.redirectUris.length <= 1) return;

    setFormData((prev) => ({
      ...prev,
      redirectUris: prev.redirectUris.filter((_, i) => i !== index),
    }));

    if (errors.redirectUris) {
      const newUriErrors = errors.redirectUris.filter((_, i) => i !== index);
      setErrors(
        (prev) =>
          ({
            ...prev,
            redirectUris: newUriErrors.length > 0 ? newUriErrors : undefined,
          }) as FormErrors
      );
    }
  };

  const handleIsActiveChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData((prev) => ({ ...prev, isActive: e.target.checked }));
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={onCancel}>
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <CardTitle>
              {isEditing ? "Edit Client" : "Create New Client"}
            </CardTitle>
            <CardDescription>
              {isEditing
                ? "Update the client configuration"
                : "Register a new OAuth client application"}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {errors.general && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
              <div className="flex items-center gap-2 text-red-800">
                <AlertCircle className="w-4 h-4" />
                <span className="text-sm">{errors.general}</span>
              </div>
            </div>
          )}

          {/* Client Name */}
          <div className="space-y-2">
            <Label htmlFor="clientName">Client Name</Label>
            <Input
              id="clientName"
              type="text"
              placeholder="My Application"
              value={formData.clientName}
              onChange={handleClientNameChange}
              className={errors.clientName ? "border-red-500" : ""}
            />
            {errors.clientName && (
              <p className="text-sm text-red-600">{errors.clientName}</p>
            )}
            <p className="text-xs text-gray-500">
              A descriptive name for your client application
            </p>
          </div>

          {/* Client Type */}
          <div className="space-y-2">
            <Label htmlFor="clientType">Client Type</Label>
            <select
              id="clientType"
              value={formData.clientType}
              onChange={(e) =>
                setFormData((prev) => ({
                  ...prev,
                  clientType: e.target.value as ClientType,
                }))
              }
              className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <option value="confidential">Confidential</option>
              <option value="public">Public</option>
            </select>
            <p className="text-xs text-gray-500">
              Confidential clients can securely store credentials. Public clients
              cannot.
            </p>
          </div>

          {/* Redirect URIs */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label>Redirect URIs</Label>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addRedirectUri}
              >
                <Plus className="w-3 h-3 mr-1" />
                Add URI
              </Button>
            </div>

            <div className="space-y-2">
              {formData.redirectUris.map((uri, index) => (
                <div key={index} className="flex items-start gap-2">
                  <div className="flex-1">
                    <Input
                      type="text"
                      placeholder="https://example.com/callback"
                      value={uri}
                      onChange={(e) =>
                        handleRedirectUriChange(index, e.target.value)
                      }
                      className={
                        errors.redirectUris?.[index] ? "border-red-500" : ""
                      }
                    />
                    {errors.redirectUris?.[index] && (
                      <p className="text-sm text-red-600 mt-1">
                        {errors.redirectUris[index]}
                      </p>
                    )}
                  </div>
                  {formData.redirectUris.length > 1 && (
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() => removeRedirectUri(index)}
                      className="text-gray-400 hover:text-red-600"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  )}
                </div>
              ))}
            </div>

            <p className="text-xs text-gray-500">
              URIs where users will be redirected after authentication. Must use
              HTTPS (localhost allowed for development).
            </p>
          </div>

          {/* Is Active (only for editing) */}
          {isEditing && (
            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                id="isActive"
                checked={formData.isActive}
                onChange={handleIsActiveChange}
                className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <Label htmlFor="isActive" className="cursor-pointer">
                Client is active
              </Label>
            </div>
          )}

          {/* Form Actions */}
          <div className="flex items-center justify-end gap-3 pt-4 border-t">
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? (
                <>
                  <span className="animate-spin mr-2">⏳</span>
                  {isEditing ? "Saving..." : "Creating..."}
                </>
              ) : (
                <>
                  <Save className="w-4 h-4 mr-2" />
                  {isEditing ? "Save Changes" : "Create Client"}
                </>
              )}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

export default ClientForm;
