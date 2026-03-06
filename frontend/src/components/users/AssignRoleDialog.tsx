/**
 * AssignRoleDialog — modal for assigning a role to a user
 */

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "../ui/dialog";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Button } from "../ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { useAssignRole } from "../../hooks/useRoles";
import { useClients } from "../../hooks/useClients";
import type { Client } from "../../types/client";
import axios from "axios";

const schema = z.object({
  role_name: z
    .string()
    .min(1, "Role name is required")
    .max(100, "Role name must be 100 characters or less")
    .refine((v) => v.trim().length > 0, "Role name cannot be only whitespace"),
});

type FormData = z.infer<typeof schema>;

interface AssignRoleDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  userId: string;
}

export function AssignRoleDialog({
  open,
  onOpenChange,
  userId,
}: AssignRoleDialogProps) {
  const assignRole = useAssignRole();
  const { data: clientsData } = useClients();
  const [clientId, setClientId] = useState("");
  const [apiError, setApiError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormData>({ resolver: zodResolver(schema) });

  const clients = clientsData?.clients ?? [];

  const onSubmit = async (data: FormData) => {
    if (!clientId) {
      setApiError("Please select a client application.");
      return;
    }
    setApiError(null);
    try {
      await assignRole.mutateAsync({
        userId,
        req: { client_id: clientId, role_name: data.role_name },
      });
      reset();
      setClientId("");
      onOpenChange(false);
    } catch (err) {
      if (axios.isAxiosError(err)) {
        setApiError(err.response?.data?.error ?? "Failed to assign role.");
      } else {
        setApiError("Something went wrong.");
      }
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Assign Role</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label>Client Application *</Label>
            <Select value={clientId} onValueChange={setClientId}>
              <SelectTrigger>
                <SelectValue placeholder="Select a client" />
              </SelectTrigger>
              <SelectContent>
                {clients.map((c: Client) => (
                  <SelectItem key={c.client_id} value={c.client_id}>
                    {c.client_name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="role_name">Role Name *</Label>
            <Input
              id="role_name"
              placeholder="e.g. editor, viewer, admin"
              {...register("role_name")}
            />
            {errors.role_name && (
              <p className="text-sm text-red-600">{errors.role_name.message}</p>
            )}
          </div>
          {apiError && <p className="text-sm text-red-600">{apiError}</p>}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={assignRole.isPending}>
              {assignRole.isPending ? "Assigning..." : "Assign Role"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
