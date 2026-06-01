/**
 * Registration form — Dell 1996 retro style
 */

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "react-router-dom";
import { Loader2, CheckCircle2, XCircle, AlertCircle } from "lucide-react";

import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Label } from "../ui/label";

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
      navigate("/verify-otp", {
        state: {
          tenantId: data.tenant_id,
          organizationName: data.organization_name,
          message: data.message,
        },
      });
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
    <div className="flex min-h-screen items-center justify-center bg-white p-4">
      <div className="w-full max-w-md">
        {/* Section eyebrow — salmon */}
        <div className="bg-[#d77a7a] px-4 py-4 mb-0">
          <h1 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[28px] font-black uppercase leading-[1.0] text-black">
            CREATE YOUR<br />ORGANIZATION
          </h1>
        </div>

        {/* Form card — ribbon card style */}
        <div className="border-x border-b border-black">
          <div className="border-b border-black bg-white px-3 py-1.5">
            <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
              REGISTER FOR KEYLES SSO
            </h3>
          </div>
          <div className="bg-[#d77a7a] px-4 py-4">
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              {/* Organization Name */}
              <div>
                <Label htmlFor="organization_name">Organization Name</Label>
                <div className="relative mt-1">
                  <Input
                    id="organization_name"
                    placeholder="Acme Corporation"
                    {...orgNameRegistration}
                    onBlur={(event) => {
                      orgNameRegistration.onBlur(event);
                      void checkFieldAvailability();
                    }}
                    className={errors.organization_name ? "border-red-700" : ""}
                  />
                  {checkingAvailability && (
                    <div className="absolute right-2 top-2">
                      <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
                    </div>
                  )}
                  {!checkingAvailability &&
                    availabilityStatus.orgName !== undefined && (
                      <div className="absolute right-2 top-2">
                        {getAvailabilityIcon(availabilityStatus.orgName)}
                      </div>
                    )}
                </div>
                {errors.organization_name && (
                  <p className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-800 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors.organization_name.message}
                  </p>
                )}
                {availabilityStatus.orgName === false && (
                  <p className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-800">
                    This organization name is already taken
                  </p>
                )}
              </div>

              {/* Email */}
              <div>
                <Label htmlFor="email">Admin Email</Label>
                <div className="relative mt-1">
                  <Input
                    id="email"
                    type="email"
                    placeholder="admin@acme.com"
                    {...emailRegistration}
                    onBlur={(event) => {
                      emailRegistration.onBlur(event);
                      void checkFieldAvailability();
                    }}
                    className={errors.email ? "border-red-700" : ""}
                  />
                  {checkingAvailability && (
                    <div className="absolute right-2 top-2">
                      <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
                    </div>
                  )}
                  {!checkingAvailability &&
                    availabilityStatus.email !== undefined && (
                      <div className="absolute right-2 top-2">
                        {getAvailabilityIcon(availabilityStatus.email)}
                      </div>
                    )}
                </div>
                {errors.email && (
                  <p className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-800 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors.email.message}
                  </p>
                )}
                {availabilityStatus.email === false && (
                  <p className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-800">
                    This email is already registered
                  </p>
                )}
              </div>

              {/* Password */}
              <div>
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  {...register("password")}
                  className={errors.password ? "border-red-700" : ""}
                />
                {errors.password && (
                  <p className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-800 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors.password.message}
                  </p>
                )}
                <p className="mt-1 font-['Times_New_Roman',Times,serif] text-[11px] text-gray-700">
                  Must be 8+ characters with uppercase, lowercase, number, and
                  special character
                </p>
              </div>

              {/* Full Name */}
              <div>
                <Label htmlFor="full_name">Full Name</Label>
                <Input
                  id="full_name"
                  placeholder="John Doe"
                  {...register("full_name")}
                  className={errors.full_name ? "border-red-700" : ""}
                />
                {errors.full_name && (
                  <p className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-800 flex items-center gap-1">
                    <AlertCircle className="h-3 w-3" />
                    {errors.full_name.message}
                  </p>
                )}
              </div>

              {/* Global Error */}
              {errors.root && (
                <div className="border border-red-700 bg-red-100 p-2">
                  <p className="font-['Times_New_Roman',Times,serif] text-sm text-red-800 flex items-center gap-2">
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
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    CREATING ORGANIZATION...
                  </>
                ) : (
                  "CREATE ORGANIZATION"
                )}
              </Button>
            </form>

            <div className="mt-4 pt-3 border-t border-black text-center">
              <p className="font-['Times_New_Roman',Times,serif] text-sm text-black">
                Already have an account?{" "}
                <a href="/login" className="text-[#0000ee] underline">
                  Sign in
                </a>
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
