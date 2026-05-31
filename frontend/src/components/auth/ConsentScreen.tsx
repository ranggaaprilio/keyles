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
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center p-4 font-['Times_New_Roman',Times,serif]">
      <div className="w-full max-w-md bg-white border border-black shadow-[2px_2px_0_#000] overflow-hidden">
        {/* Periwinkle Eyebrow */}
        <div className="bg-[#8c9ae0] px-6 py-4">
          <div className="flex justify-center mb-3">
            {client.logo_uri ? (
              <img
                src={client.logo_uri}
                alt={`${client.client_name} logo`}
                className="h-16 w-16 border-2 border-black bg-white p-1"
              />
            ) : (
              <div className="h-16 w-16 border-2 border-black bg-white flex items-center justify-center text-2xl">
                🔑
              </div>
            )}
          </div>
          <h1 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-white text-center text-lg">
            {client.client_name}
          </h1>
          <p className="text-white/80 text-sm text-center font-['Times_New_Roman',Times,serif]">
            wants to access your account
          </p>
        </div>

        {/* User info */}
        <div className="px-6 py-3 bg-gray-100 border-b border-black">
          <p className="text-sm text-gray-600 text-center font-['Times_New_Roman',Times,serif]">
            Signing in as{" "}
            <span className="font-bold text-black">{user}</span>
          </p>
        </div>

        {/* Permissions - Ribbon Cards */}
        <div className="px-6 py-6">
          <h2 className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-gray-700 mb-4">
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
                  className="border border-black shadow-[2px_2px_0_#000]"
                >
                  {/* Ribbon card title bar */}
                  <div className="bg-white border-b border-black px-3 py-1.5">
                    <span className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px]">
                      {info.icon} {info.name}
                    </span>
                  </div>
                  {/* Tinted body */}
                  <div className="bg-[#8c9ae0]/15 px-3 py-2">
                    <p className="font-['Times_New_Roman',Times,serif] text-sm text-gray-700">
                      {info.description}
                    </p>
                  </div>
                </li>
              );
            })}
          </ul>
        </div>

        {/* Policy links */}
        {(client.policy_uri || client.tos_uri) && (
          <div className="px-6 py-3 bg-gray-100 border-y border-black text-center text-sm text-gray-500 font-['Times_New_Roman',Times,serif]">
            By clicking Allow, you agree to {client.client_name}&apos;s{" "}
            {client.tos_uri && (
              <a
                href={client.tos_uri}
                target="_blank"
                rel="noopener noreferrer"
                className="text-[#0000ee] underline"
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
                className="text-[#0000ee] underline"
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
            className="flex-1 px-4 py-3 text-[12px] font-bold uppercase tracking-[1.5px] font-[Helvetica,Arial,system-ui,sans-serif] border border-black bg-white text-black hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-dell-red disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
            className="flex-1 px-4 py-3 text-[12px] font-bold uppercase tracking-[1.5px] font-[Helvetica,Arial,system-ui,sans-serif] border border-black bg-black text-white hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-dell-red focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
        <div className="px-6 py-3 bg-amber-100 border-t border-black">
          <p className="text-xs text-amber-900 text-center font-['Times_New_Roman',Times,serif]">
            ⚠️ Only grant access to applications you trust. You can revoke
            access at any time.
          </p>
        </div>
      </div>

      {/* Footer */}
      <p className="mt-6 text-sm text-gray-500 font-['Times_New_Roman',Times,serif]">
        Protected by <span className="font-bold">Keyles SSO</span>
      </p>
    </div>
  );
}

export default ConsentScreen;