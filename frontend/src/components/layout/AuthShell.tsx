/**
 * Centered card layout shared by the public auth pages
 * (login, register, OTP verification, invitation, OAuth consent).
 */

import { ReactNode } from "react";
import { KeyRound } from "lucide-react";

import { cn } from "@/lib/utils";

interface AuthShellProps {
  title: string;
  description?: ReactNode;
  children: ReactNode;
  /** Rendered below the card, outside it (help text, security notes). */
  footer?: ReactNode;
  /** Card width; defaults to max-w-md. */
  size?: "md" | "lg";
}

export function AuthShell({ title, description, children, footer, size = "md" }: AuthShellProps) {
  return (
    <div className="min-h-screen bg-muted flex items-center justify-center px-4 py-12">
      <div className={cn("w-full", size === "lg" ? "max-w-2xl" : "max-w-md")}>
        <div className="mb-6 flex items-center justify-center gap-2.5">
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary text-primary-foreground">
            <KeyRound className="h-4 w-4" />
          </span>
          <span className="text-lg font-medium tracking-tight text-foreground">Keyles</span>
        </div>

        <div className="rounded-xl border border-border bg-background p-6 shadow-card sm:p-8">
          <h1 className="text-2xl font-medium leading-tight tracking-tight text-foreground">{title}</h1>
          {description && (
            <p className="mt-1.5 text-sm text-muted-foreground">{description}</p>
          )}
          <div className="mt-6">{children}</div>
        </div>

        {footer && <div className="mt-6 text-center text-xs text-muted-foreground">{footer}</div>}
      </div>
    </div>
  );
}
