import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { X, Plus, Loader2, ArrowLeft } from 'lucide-react';
import { useClient, useUpdateClient } from '@/hooks/useClients';

const editClientSchema = z.object({
  client_name: z
    .string()
    .min(3, 'Name must be at least 3 characters')
    .max(100, 'Name must not exceed 100 characters'),
  description: z.string().max(500, 'Description must not exceed 500 characters').optional(),
  redirect_uris: z
    .array(z.string().url('Must be a valid URL'))
    .min(1, 'At least one redirect URI is required'),
});

type EditClientFormData = z.infer<typeof editClientSchema>;

interface EditClientFormProps {
  clientId: string;
  onCancel: () => void;
  onSuccess: () => void;
}

export function EditClientForm({ clientId, onCancel, onSuccess }: EditClientFormProps) {
  const { data: client, isLoading } = useClient(clientId);
  const updateMutation = useUpdateClient();
  const [redirectURIs, setRedirectURIs] = useState<string[]>([]);
  const [uriErrors, setUriErrors] = useState<Record<number, string>>({});

  const {
    register,
    handleSubmit,
    setValue,
    reset,
    formState: { errors },
  } = useForm<EditClientFormData>({
    resolver: zodResolver(editClientSchema),
  });

  // Pre-populate form when client data loads
  useEffect(() => {
    if (client) {
      reset({
        client_name: client.client_name,
        description: client.description || '',
        redirect_uris: client.redirect_uris,
      });
      setRedirectURIs(client.redirect_uris);
    }
  }, [client, reset]);

  const validateURI = (uri: string, index: number): boolean => {
    if (!uri) {
      setUriErrors((prev) => ({ ...prev, [index]: 'URI is required' }));
      return false;
    }
    try {
      const parsed = new URL(uri);
      if (parsed.hash) {
        setUriErrors((prev) => ({ ...prev, [index]: 'URI must not contain a fragment' }));
        return false;
      }
      const host = parsed.hostname;
      if (parsed.protocol !== 'https:' && host !== 'localhost' && host !== '127.0.0.1') {
        setUriErrors((prev) => ({ ...prev, [index]: 'Must use HTTPS (except localhost)' }));
        return false;
      }
      setUriErrors((prev) => {
        const copy = { ...prev };
        delete copy[index];
        return copy;
      });
      return true;
    } catch {
      setUriErrors((prev) => ({ ...prev, [index]: 'Must be a valid URL' }));
      return false;
    }
  };

  const addRedirectURI = () => {
    const updated = [...redirectURIs, ''];
    setRedirectURIs(updated);
    setValue('redirect_uris', updated);
  };

  const removeRedirectURI = (index: number) => {
    if (redirectURIs.length <= 1) return;
    const updated = redirectURIs.filter((_, i) => i !== index);
    setRedirectURIs(updated);
    setValue('redirect_uris', updated);
  };

  const updateRedirectURI = (index: number, value: string) => {
    const updated = [...redirectURIs];
    updated[index] = value;
    setRedirectURIs(updated);
    setValue('redirect_uris', updated);
    if (value) validateURI(value, index);
  };

  const onSubmit = (data: EditClientFormData) => {
    let allValid = true;
    data.redirect_uris.forEach((uri, i) => {
      if (!validateURI(uri, i)) allValid = false;
    });
    if (!allValid) return;

    updateMutation.mutate(
      {
        clientId,
        data: {
          client_name: data.client_name,
          description: data.description || null,
          redirect_uris: data.redirect_uris,
        },
      },
      { onSuccess }
    );
  };

  if (isLoading) {
    return <Skeleton className="h-96 w-full rounded-lg" />;
  }

  if (!client) {
    return <div className="text-center py-8 text-destructive">Client not found.</div>;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={onCancel}>
            <ArrowLeft className="h-4 w-4 mr-1" /> Back
          </Button>
          <CardTitle>Edit {client.client_name}</CardTitle>
          <Badge variant="outline">{client.client_type}</Badge>
        </div>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* Client Type (read-only) */}
          <div className="space-y-2">
            <Label>Client Type</Label>
            <p className="text-sm text-muted-foreground">
              <Badge variant={client.client_type === 'confidential' ? 'default' : 'secondary'}>
                {client.client_type}
              </Badge>{' '}
              — cannot be changed after creation
            </p>
          </div>

          {/* Client Name */}
          <div className="space-y-2">
            <Label htmlFor="client_name">Application Name *</Label>
            <Input id="client_name" {...register('client_name')} />
            {errors.client_name && (
              <p className="text-sm text-destructive">{errors.client_name.message}</p>
            )}
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea id="description" rows={3} {...register('description')} />
            {errors.description && (
              <p className="text-sm text-destructive">{errors.description.message}</p>
            )}
          </div>

          {/* Redirect URIs */}
          <div className="space-y-3">
            <Label>Redirect URIs *</Label>
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
                    <p className="text-sm text-destructive mt-1">{uriErrors[index]}</p>
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
            <Button type="button" variant="outline" size="sm" onClick={addRedirectURI}>
              <Plus className="h-4 w-4 mr-1" /> Add Redirect URI
            </Button>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-4">
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled={updateMutation.isPending}>
              {updateMutation.isPending && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Save Changes
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
