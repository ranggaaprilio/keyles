/**
 * AcceptInvitationForm — Dell 1996 retro style
 */

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Button } from "../ui/button";

const schema = z
  .object({
    password: z
      .string()
      .min(8, "Password must be at least 8 characters")
      .regex(/[a-z]/, "Must contain a lowercase letter")
      .regex(/[A-Z]/, "Must contain an uppercase letter")
      .regex(/[0-9]/, "Must contain a digit"),
    confirmPassword: z.string(),
  })
  .refine((d) => d.password === d.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

type FormData = z.infer<typeof schema>;

function strengthLevel(pw: string): {
  label: string;
  color: string;
  width: string;
} {
  let score = 0;
  if (pw.length >= 8) score++;
  if (pw.length >= 12) score++;
  if (/[a-z]/.test(pw) && /[A-Z]/.test(pw)) score++;
  if (/[0-9]/.test(pw)) score++;
  if (/[^a-zA-Z0-9]/.test(pw)) score++;

  if (score <= 1) return { label: "Weak", color: "bg-red-700", width: "w-1/4" };
  if (score <= 2) return { label: "Fair", color: "bg-yellow-600", width: "w-1/2" };
  if (score <= 3) return { label: "Good", color: "bg-blue-700", width: "w-3/4" };
  return { label: "Strong", color: "bg-green-700", width: "w-full" };
}

interface AcceptInvitationFormProps {
  email: string;
  displayName?: string;
  onSubmit: (password: string) => Promise<void>;
  isSubmitting: boolean;
  error?: string;
}

export function AcceptInvitationForm({
  email,
  displayName,
  onSubmit,
  isSubmitting,
  error,
}: AcceptInvitationFormProps) {
  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
  });

  const password = watch("password", "");
  const strength = strengthLevel(password);

  const onFormSubmit = async (data: FormData) => {
    await onSubmit(data.password);
  };

  return (
    <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
      {/* Read-only email */}
      <div>
        <Label>Email</Label>
        <Input value={email} disabled className="mt-1 bg-gray-100" />
      </div>

      {displayName && (
        <div>
          <Label>Name</Label>
          <Input value={displayName} disabled className="mt-1 bg-gray-100" />
        </div>
      )}

      {/* Password */}
      <div>
        <Label htmlFor="password">Password *</Label>
        <Input id="password" type="password" {...register("password")} className="mt-1" />
        {password.length > 0 && (
          <div className="mt-1 space-y-1">
            <div className="h-1.5 w-full bg-gray-200 overflow-hidden border border-black">
              <div
                className={`h-full ${strength.color} ${strength.width} transition-all`}
              />
            </div>
            <p className="font-['Times_New_Roman',Times,serif] text-[11px] text-gray-600">
              {strength.label}
            </p>
          </div>
        )}
        {errors.password && (
          <p className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-700">
            {errors.password.message}
          </p>
        )}
      </div>

      {/* Confirm password */}
      <div>
        <Label htmlFor="confirmPassword">Confirm Password *</Label>
        <Input
          id="confirmPassword"
          type="password"
          {...register("confirmPassword")}
          className="mt-1"
        />
        {errors.confirmPassword && (
          <p className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-700">
            {errors.confirmPassword.message}
          </p>
        )}
      </div>

      {error && (
        <p className="font-['Times_New_Roman',Times,serif] text-sm text-red-700">
          {error}
        </p>
      )}

      <Button type="submit" className="w-full" disabled={isSubmitting}>
        {isSubmitting ? "CREATING ACCOUNT..." : "CREATE ACCOUNT"}
      </Button>
    </form>
  );
}
