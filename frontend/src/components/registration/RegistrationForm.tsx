/**
 * Registration form
 */

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "react-router-dom";
import { Loader2, CheckCircle2, XCircle, AlertCircle } from "lucide-react";

import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { AuthShell } from "../layout/AuthShell";

import { registrationSchema, RegistrationFormData } from "./RegistrationSchema";
import { useTenantRegistration } from "../../hooks/useTenantRegistration";
import { checkAvailability } from "../../services/api/tenant";
import { ApiException } from "../../types/api";

export function RegistrationForm() {
  const navigate = useNavigate();
  const [availabilityStatus, setAvailabilityStatus] = useState<{
    orgName?: boolean;
    email?: boolean;
  }>({});
  const [checkingAvailability, setCheckingAvailability] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setError,
  } = useForm<RegistrationFormData>({
    resolver: zodResolver(registrationSchema),
    mode: "onBlur",
  });

  const mutation = useTenantRegistration({
    onSuccess: (data) => {
      if (data.status === "active") {
        navigate("/login", {
          state: { message: data.message },
        });
      } else {
        navigate("/verify-otp", {
          state: {
            tenantId: data.tenant_id,
            organizationName: data.organization_name,
            message: data.message,
          },
        });
      }
    },
    onError: (err: ApiException) => {
      setError("root", { message: err.message });
    },
  });

  const orgName = watch("organization_name");
  const email = watch("email");
  const orgNameRegistration = register("organization_name");
  const emailRegistration = register("email");

  const checkFieldAvailability = async () => {
    if (!orgName && !email) return;
    setCheckingAvailability(true);
    try {
      const result = await checkAvailability({
        organization_name: orgName ?? "",
        email: email ?? "",
      });
      setAvailabilityStatus({
        orgName: result.organization_name_available,
        email: result.email_available,
      });
    } catch {
      // Silently fail — availability check is non-critical
    } finally {
      setCheckingAvailability(false);
    }
  };

  const onSubmit = async (data: RegistrationFormData) => {
    mutation.mutate(data);
  };

  const getAvailabilityIcon = (available?: boolean) => {
    if (available === true) return <CheckCircle2 className="h-4 w-4 text-green-700" />;
    if (available === false) return <XCircle className="h-4 w-4 text-red-700" />;
    return null;
  };

  return (
    <AuthShell
      title="Create your organization"
      description="Register for Keyles SSO. You will be the first administrator."
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
        {/* Organization Name */}
        <div className="space-y-1.5">
          <Label htmlFor="organization_name">Organization Name</Label>
          <div className="relative">
            <Input
              id="organization_name"
              placeholder="Acme Corporation"
              {...orgNameRegistration}
              onBlur={(event) => {
                orgNameRegistration.onBlur(event);
                void checkFieldAvailability();
              }}
              className={errors.organization_name ? "border-destructive pr-9" : "pr-9"}
            />
            {checkingAvailability && (
              <div className="absolute right-3 top-3">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            )}
            {!checkingAvailability &&
              availabilityStatus.orgName !== undefined && (
                <div className="absolute right-3 top-3">
                  {getAvailabilityIcon(availabilityStatus.orgName)}
                </div>
              )}
          </div>
          {errors.organization_name && (
            <p className="text-sm text-red-800 flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              {errors.organization_name.message}
            </p>
          )}
          {availabilityStatus.orgName === false && (
            <p className="text-sm text-red-800">
              This organization name is already taken
            </p>
          )}
        </div>

        {/* Email */}
        <div className="space-y-1.5">
          <Label htmlFor="email">Admin Email</Label>
          <div className="relative">
            <Input
              id="email"
              type="email"
              placeholder="admin@acme.com"
              {...emailRegistration}
              onBlur={(event) => {
                emailRegistration.onBlur(event);
                void checkFieldAvailability();
              }}
              className={errors.email ? "border-destructive pr-9" : "pr-9"}
            />
            {checkingAvailability && (
              <div className="absolute right-3 top-3">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            )}
            {!checkingAvailability &&
              availabilityStatus.email !== undefined && (
                <div className="absolute right-3 top-3">
                  {getAvailabilityIcon(availabilityStatus.email)}
                </div>
              )}
          </div>
          {errors.email && (
            <p className="text-sm text-red-800 flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              {errors.email.message}
            </p>
          )}
          {availabilityStatus.email === false && (
            <p className="text-sm text-red-800">
              This email is already registered
            </p>
          )}
        </div>

        {/* Password */}
        <div className="space-y-1.5">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            placeholder="••••••••"
            {...register("password")}
            className={errors.password ? "border-destructive" : ""}
          />
          {errors.password && (
            <p className="text-sm text-red-800 flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              {errors.password.message}
            </p>
          )}
          <p className="text-xs text-muted-foreground">
            Must be 8+ characters with uppercase, lowercase, number, and
            special character
          </p>
        </div>

        {/* Full Name */}
        <div className="space-y-1.5">
          <Label htmlFor="full_name">Full Name</Label>
          <Input
            id="full_name"
            placeholder="John Doe"
            {...register("full_name")}
            className={errors.full_name ? "border-destructive" : ""}
          />
          {errors.full_name && (
            <p className="text-sm text-red-800 flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              {errors.full_name.message}
            </p>
          )}
        </div>

        {/* Global Error */}
        {errors.root && (
          <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
            <p className="flex items-center gap-2">
              <AlertCircle className="h-4 w-4" />
              {errors.root.message}
            </p>
          </div>
        )}

        {/* Submit Button */}
        <Button
          type="submit"
          className="w-full"
          disabled={mutation.isPending}
        >
          {mutation.isPending ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              Creating organization...
            </>
          ) : (
            "Create organization"
          )}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Already have an account?{" "}
        <a href="/login" className="font-medium text-foreground underline underline-offset-4 hover:text-muted-foreground">
          Sign in
        </a>
      </p>
    </AuthShell>
  );
}
