import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { X, Plus, Loader2 } from "lucide-react";
import type { ClientType } from "@/types/client";

const createClientSchema = z.object({
  client_name: z
    .string()
    .min(3, "Name must be at least 3 characters")
    .max(100, "Name must not exceed 100 characters"),
  description: z
    .string()
    .max(500, "Description must not exceed 500 characters")
    .optional()
    .default(""),
  client_type: z.enum(["confidential", "public"] as const),
  redirect_uris: z
    .array(z.string().url("Must be a valid URL"))
    .min(1, "At least one redirect URI is required"),
});

type CreateClientFormData = z.infer<typeof createClientSchema>;

interface CreateClientFormProps {
  onSubmit: (data: CreateClientFormData) => void;
  onCancel: () => void;
  isLoading?: boolean;
}

export function CreateClientForm({
  onSubmit,
  onCancel,
  isLoading,
}: CreateClientFormProps) {
  const [redirectURIs, setRedirectURIs] = useState<string[]>([""]);
  const [uriErrors, setUriErrors] = useState<Record<number, string>>({});

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<CreateClientFormData>({
    resolver: zodResolver(createClientSchema),
    defaultValues: {
      client_name: "",
      description: "",
      client_type: "confidential",
      redirect_uris: [""],
    },
  });

  const clientType = watch("client_type");

  const validateURI = (uri: string, index: number): boolean => {
    if (!uri) {
      setUriErrors((prev) => ({ ...prev, [index]: "URI is required" }));
      return false;
    }
    try {
      const parsed = new URL(uri);
      if (parsed.hash) {
        setUriErrors((prev) => ({
          ...prev,
          [index]: "URI must not contain a fragment",
        }));
        return false;
      }
      const host = parsed.hostname;
      if (
        parsed.protocol !== "https:" &&
        host !== "localhost" &&
        host !== "127.0.0.1"
      ) {
        setUriErrors((prev) => ({
          ...prev,
          [index]: "Must use HTTPS (except localhost)",
        }));
        return false;
      }
      setUriErrors((prev) => {
        const copy = { ...prev };
        delete copy[index];
        return copy;
      });
      return true;
    } catch {
      setUriErrors((prev) => ({ ...prev, [index]: "Must be a valid URL" }));
      return false;
    }
  };

  const addRedirectURI = () => {
    const updated = [...redirectURIs, ""];
    setRedirectURIs(updated);
    setValue("redirect_uris", updated);
  };

  const removeRedirectURI = (index: number) => {
    if (redirectURIs.length <= 1) return;
    const updated = redirectURIs.filter((_, i) => i !== index);
    setRedirectURIs(updated);
    setValue("redirect_uris", updated);
    setUriErrors((prev) => {
      const copy = { ...prev };
      delete copy[index];
      return copy;
    });
  };

  const updateRedirectURI = (index: number, value: string) => {
    const updated = [...redirectURIs];
    updated[index] = value;
    setRedirectURIs(updated);
    setValue("redirect_uris", updated);
    if (value) validateURI(value, index);
  };

  const onFormSubmit = (data: CreateClientFormData) => {
    // Validate all URIs
    let allValid = true;
    data.redirect_uris.forEach((uri, i) => {
      if (!validateURI(uri, i)) allValid = false;
    });
    if (!allValid) return;
    onSubmit(data);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Register New Client Application</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-6">
          {/* Client Name */}
          <div className="space-y-2">
            <Label htmlFor="client_name">Application Name *</Label>
            <Input
              id="client_name"
              placeholder="My Application"
              {...register("client_name")}
            />
            {errors.client_name && (
              <p className="text-sm text-red-700 font-['Times_New_Roman',Times,serif]">
                {errors.client_name.message}
              </p>
            )}
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Brief description of your application"
              rows={3}
              {...register("description")}
            />
            {errors.description && (
              <p className="text-sm text-red-700 font-['Times_New_Roman',Times,serif]">
                {errors.description.message}
              </p>
            )}
          </div>

          {/* Client Type */}
          <div className="space-y-3">
            <Label>Client Type *</Label>
            <RadioGroup
              value={clientType}
              onValueChange={(value: ClientType) =>
                setValue("client_type", value)
              }
              className="space-y-3"
            >
              <div className="flex items-start space-x-3 border border-black p-3">
                <RadioGroupItem
                  value="confidential"
                  id="confidential"
                  className="mt-1"
                />
                <div>
                  <Label
                    htmlFor="confidential"
                    className="font-medium cursor-pointer"
                  >
                    Confidential
                  </Label>
                  <p className="text-sm text-gray-600 font-['Times_New_Roman',Times,serif]">
                    Server-side applications that can securely store a client
                    secret. A secret will be generated for authentication.
                  </p>
                </div>
              </div>
              <div className="flex items-start space-x-3 border border-black p-3">
                <RadioGroupItem value="public" id="public" className="mt-1" />
                <div>
                  <Label
                    htmlFor="public"
                    className="font-medium cursor-pointer"
                  >
                    Public
                  </Label>
                  <p className="text-sm text-gray-600 font-['Times_New_Roman',Times,serif]">
                    SPAs, mobile, or desktop apps that cannot securely store
                    secrets. Uses PKCE for authorization. No client secret
                    generated.
                  </p>
                </div>
              </div>
            </RadioGroup>
          </div>

          {/* Redirect URIs */}
          <div className="space-y-3">
            <Label>Redirect URIs *</Label>
            <p className="text-sm text-gray-600 font-['Times_New_Roman',Times,serif]">
              HTTPS required for production. HTTP allowed for localhost only.
            </p>
            {redirectURIs.map((uri, index) => (
              <div key={index} className="flex gap-2">
                <div className="flex-1">
                  <Input
                    placeholder="https://example.com/callback"
                    value={uri}
                    onChange={(e) => updateRedirectURI(index, e.target.value)}
                    onBlur={() => uri && validateURI(uri, index)}
                  />
                  {uriErrors[index] && (
                    <p className="text-sm text-red-700 font-['Times_New_Roman',Times,serif] mt-1">
                      {uriErrors[index]}
                    </p>
                  )}
                </div>
                {redirectURIs.length > 1 && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeRedirectURI(index)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                )}
              </div>
            ))}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={addRedirectURI}
            >
              <Plus className="h-4 w-4 mr-1" /> Add Redirect URI
            </Button>
            {errors.redirect_uris && (
              <p className="text-sm text-red-700 font-['Times_New_Roman',Times,serif]">
                {errors.redirect_uris.message}
              </p>
            )}
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-4">
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Register Application
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
