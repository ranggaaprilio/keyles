/**
 * ConsentScreen Component
 * Displays OAuth consent screen for user approval/denial
 */

import { useState } from "react";
import type { ConsentScreenProps } from "../../types/oauth";

/**
 * Scope display names and descriptions
 */
const SCOPE_INFO: Record<
  string,
  { name: string; description: string; icon: string }
> = {
  openid: {
    name: "OpenID",
    description: "Verify your identity",
    icon: "🔐",
  },
  profile: {
    name: "Profile",
    description: "Access your name and profile picture",
    icon: "👤",
  },
  email: {
    name: "Email",
    description: "Access your email address",
    icon: "✉️",
  },
  offline_access: {
    name: "Offline Access",
    description: "Access your data when you're not present",
    icon: "🔄",
  },
};

/**
 * ConsentScreen component for OAuth authorization
 *
 * @example
 * ```tsx
 * <ConsentScreen
 *   client={{
 *     client_id: 'my-app',
 *     client_name: 'My Application',
 *     logo_uri: 'https://example.com/logo.png',
 *   }}
 *   scopes={['openid', 'profile', 'email']}
 *   user="user@example.com"
 *   onApprove={() => handleApprove()}
 *   onDeny={() => handleDeny()}
 * />
 * ```
 */
export function ConsentScreen({
  client,
  scopes,
  user,
  onApprove,
  onDeny,
  isLoading = false,
}: ConsentScreenProps): JSX.Element {
  const [isApproving, setIsApproving] = useState(false);
  const [isDenying, setIsDenying] = useState(false);

  const handleApprove = async () => {
    setIsApproving(true);
    try {
      await onApprove();
    } finally {
      setIsApproving(false);
    }
  };

  const handleDeny = async () => {
    setIsDenying(true);
    try {
      await onDeny();
    } finally {
      setIsDenying(false);
    }
  };

  const buttonDisabled = isLoading || isApproving || isDenying;

  return (
    <div className="min-h-screen bg-muted flex flex-col items-center justify-center px-4 py-12">
      <div className="w-full max-w-md rounded-xl border border-border bg-background shadow-card overflow-hidden">
        {/* Client identity */}
        <div className="px-6 pt-8 pb-6 text-center">
          <div className="flex justify-center mb-4">
            {client.logo_uri ? (
              <img
                src={client.logo_uri}
                alt={`${client.client_name} logo`}
                className="h-16 w-16 rounded-lg border border-border bg-background object-contain p-1"
              />
            ) : (
              <div className="h-16 w-16 rounded-lg bg-muted flex items-center justify-center text-2xl">
                🔑
              </div>
            )}
          </div>
          <h1 className="text-xl font-medium tracking-tight text-foreground">
            {client.client_name}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            wants to access your account
          </p>
        </div>

        {/* User info */}
        <div className="px-6 py-3 bg-muted border-y border-border">
          <p className="text-sm text-muted-foreground text-center">
            Signing in as{" "}
            <span className="font-medium text-foreground">{user}</span>
          </p>
        </div>

        {/* Permissions */}
        <div className="px-6 py-6">
          <h2 className="text-sm font-medium text-foreground mb-3">
            This will allow {client.client_name} to:
          </h2>

          <ul className="space-y-2">
            {scopes.map((scope) => {
              const info = SCOPE_INFO[scope] || {
                name: scope,
                description: `Access to ${scope}`,
                icon: "🔷",
              };

              return (
                <li
                  key={scope}
                  className="flex items-start gap-3 rounded-md border border-border px-3 py-2.5"
                >
                  <span className="text-base leading-6">{info.icon}</span>
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-foreground">{info.name}</p>
                    <p className="text-sm text-muted-foreground">{info.description}</p>
                  </div>
                </li>
              );
            })}
          </ul>
        </div>

        {/* Policy links */}
        {(client.policy_uri || client.tos_uri) && (
          <div className="px-6 py-3 bg-muted border-y border-border text-center text-xs text-muted-foreground">
            By clicking Allow, you agree to {client.client_name}&apos;s{" "}
            {client.tos_uri && (
              <a
                href={client.tos_uri}
                target="_blank"
                rel="noopener noreferrer"
                className="text-foreground underline underline-offset-4 hover:text-muted-foreground"
              >
                Terms of Service
              </a>
            )}
            {client.tos_uri && client.policy_uri && " and "}
            {client.policy_uri && (
              <a
                href={client.policy_uri}
                target="_blank"
                rel="noopener noreferrer"
                className="text-foreground underline underline-offset-4 hover:text-muted-foreground"
              >
                Privacy Policy
              </a>
            )}
          </div>
        )}

        {/* Actions */}
        <div className="px-6 py-5 flex gap-3">
          <button
            type="button"
            onClick={handleDeny}
            disabled={buttonDisabled}
            className="flex-1 inline-flex h-11 items-center justify-center rounded-md border border-input bg-background px-4 text-sm font-medium text-foreground hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-fast"
          >
            {isDenying ? (
              <span className="flex items-center justify-center gap-2">
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                    fill="none"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                Denying...
              </span>
            ) : (
              "Deny"
            )}
          </button>

          <button
            type="button"
            onClick={handleApprove}
            disabled={buttonDisabled}
            className="flex-1 inline-flex h-11 items-center justify-center rounded-full bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/85 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-fast"
          >
            {isApproving ? (
              <span className="flex items-center justify-center gap-2">
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                    fill="none"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                Allowing...
              </span>
            ) : (
              "Allow"
            )}
          </button>
        </div>

        {/* Security notice */}
        <div className="px-6 py-3 bg-warning/10 border-t border-border">
          <p className="text-xs text-warning text-center">
            ⚠️ Only grant access to applications you trust. You can revoke
            access at any time.
          </p>
        </div>
      </div>

      {/* Footer */}
      <p className="mt-6 text-sm text-muted-foreground">
        Protected by <span className="font-medium text-foreground">Keyles SSO</span>
      </p>
    </div>
  );
}

export default ConsentScreen;