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
    <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-md bg-white rounded-xl shadow-lg overflow-hidden">
        {/* Header */}
        <div className="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-8 text-white text-center">
          <div className="flex justify-center mb-4">
            {client.logo_uri ? (
              <img
                src={client.logo_uri}
                alt={`${client.client_name} logo`}
                className="h-16 w-16 rounded-full bg-white p-1"
              />
            ) : (
              <div className="h-16 w-16 rounded-full bg-white/20 flex items-center justify-center text-2xl">
                🔑
              </div>
            )}
          </div>
          <h1 className="text-xl font-bold mb-2">{client.client_name}</h1>
          <p className="text-blue-100 text-sm">wants to access your account</p>
        </div>

        {/* User info */}
        <div className="px-6 py-4 bg-gray-50 border-b">
          <p className="text-sm text-gray-600 text-center">
            Signing in as{" "}
            <span className="font-medium text-gray-900">{user}</span>
          </p>
        </div>

        {/* Permissions */}
        <div className="px-6 py-6">
          <h2 className="text-sm font-semibold text-gray-700 uppercase tracking-wide mb-4">
            This will allow {client.client_name} to:
          </h2>

          <ul className="space-y-3">
            {scopes.map((scope) => {
              const info = SCOPE_INFO[scope] || {
                name: scope,
                description: `Access to ${scope}`,
                icon: "🔷",
              };

              return (
                <li
                  key={scope}
                  className="flex items-start gap-3 p-3 rounded-lg bg-gray-50"
                >
                  <span className="text-xl flex-shrink-0">{info.icon}</span>
                  <div>
                    <p className="font-medium text-gray-900">{info.name}</p>
                    <p className="text-sm text-gray-600">{info.description}</p>
                  </div>
                </li>
              );
            })}
          </ul>
        </div>

        {/* Policy links */}
        {(client.policy_uri || client.tos_uri) && (
          <div className="px-6 py-3 bg-gray-50 border-t border-b text-center text-sm text-gray-500">
            By clicking Allow, you agree to {client.client_name}'s{" "}
            {client.tos_uri && (
              <a
                href={client.tos_uri}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:underline"
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
                className="text-blue-600 hover:underline"
              >
                Privacy Policy
              </a>
            )}
          </div>
        )}

        {/* Actions */}
        <div className="px-6 py-4 flex gap-3">
          <button
            type="button"
            onClick={handleDeny}
            disabled={buttonDisabled}
            className="flex-1 px-4 py-3 text-gray-700 bg-gray-100 rounded-lg font-medium hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-300 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
            className="flex-1 px-4 py-3 text-white bg-blue-600 rounded-lg font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
        <div className="px-6 py-3 bg-amber-50 border-t">
          <p className="text-xs text-amber-800 text-center">
            ⚠️ Only grant access to applications you trust. You can revoke
            access at any time.
          </p>
        </div>
      </div>

      {/* Footer */}
      <p className="mt-6 text-sm text-gray-500">
        Protected by <span className="font-medium">Keyles SSO</span>
      </p>
    </div>
  );
}

export default ConsentScreen;
