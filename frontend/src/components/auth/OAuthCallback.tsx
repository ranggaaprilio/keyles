/**
 * OAuthCallback Component
 * Handles OAuth 2.0 authorization callback
 * Extracts code and state from URL, exchanges for tokens
 */

import { useEffect, useState, useCallback } from "react";
import type { TokenResponse, AuthorizationErrorCode } from "../../types/oauth";

/**
 * Props for OAuthCallback component
 */
interface OAuthCallbackProps {
  /** Called on successful token exchange */
  onSuccess: (tokens: TokenResponse) => void;
  /** Called on error */
  onError: (error: string, errorCode?: AuthorizationErrorCode) => void;
  /** Custom loading component */
  loadingComponent?: React.ReactNode;
  /** Custom error component */
  errorComponent?: (error: string) => React.ReactNode;
  /** OAuth service completeCallback function */
  completeCallback: (url?: string) => Promise<TokenResponse>;
  /** Whether to send result via postMessage (for popup flows) */
  usePostMessage?: boolean;
  /** Target origin for postMessage */
  postMessageOrigin?: string;
}

/**
 * OAuth Callback handler component
 *
 * @example
 * ```tsx
 * // In your callback route component:
 * function OAuthCallbackPage() {
 *   const navigate = useNavigate();
 *   const { service, handleCallback } = useOAuth({ config });
 *
 *   return (
 *     <OAuthCallback
 *       completeCallback={handleCallback}
 *       onSuccess={(tokens) => {
 *         navigate('/dashboard');
 *       }}
 *       onError={(error) => {
 *         navigate('/login', { state: { error } });
 *       }}
 *     />
 *   );
 * }
 * ```
 */
export function OAuthCallback({
  onSuccess,
  onError,
  loadingComponent,
  errorComponent,
  completeCallback,
  usePostMessage = false,
  postMessageOrigin = window.location.origin,
}: OAuthCallbackProps): JSX.Element {
  const [error, setError] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(true);

  const handlePostMessage = useCallback(
    (success: boolean, data: TokenResponse | { error: string }) => {
      if (usePostMessage && window.opener) {
        window.opener.postMessage(
          {
            type: "oauth_callback",
            success,
            data,
          },
          postMessageOrigin
        );
        window.close();
      }
    },
    [usePostMessage, postMessageOrigin]
  );

  useEffect(() => {
    let isMounted = true;

    const processCallback = async () => {
      try {
        // Check for error in URL first
        const url = new URL(window.location.href);
        const errorParam = url.searchParams.get("error");
        const errorDescription = url.searchParams.get("error_description");

        if (errorParam) {
          const errorMessage = errorDescription || errorParam;

          if (isMounted) {
            setError(errorMessage);
            setIsProcessing(false);
          }

          if (usePostMessage) {
            handlePostMessage(false, { error: errorMessage });
          } else {
            onError(errorMessage, errorParam as AuthorizationErrorCode);
          }
          return;
        }

        // Process the callback
        const tokens = await completeCallback();

        if (isMounted) {
          setIsProcessing(false);
        }

        if (usePostMessage) {
          handlePostMessage(true, tokens);
        } else {
          onSuccess(tokens);
        }
      } catch (err: unknown) {
        const errorMessage =
          err instanceof Error ? err.message : "OAuth callback failed";

        if (isMounted) {
          setError(errorMessage);
          setIsProcessing(false);
        }

        if (usePostMessage) {
          handlePostMessage(false, { error: errorMessage });
        } else {
          onError(errorMessage);
        }
      }
    };

    processCallback();

    return () => {
      isMounted = false;
    };
  }, [completeCallback, onSuccess, onError, usePostMessage, handlePostMessage]);

  if (isProcessing) {
    return (
      <>
        {loadingComponent || (
          <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 font-['Times_New_Roman',Times,serif]">
            <div className="border-4 border-black w-12 h-12 border-t-transparent animate-spin"></div>
            <p className="mt-4 text-sm text-gray-600">Completing sign in...</p>
          </div>
        )}
      </>
    );
  }

  if (error) {
    return (
      <>
        {errorComponent ? (
          errorComponent(error)
        ) : (
          <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 font-['Times_New_Roman',Times,serif]">
            <div className="bg-white border border-black shadow-[2px_2px_0_#000] p-6 max-w-md">
              <h2 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-lg text-red-800 mb-2">
                Authentication Failed
              </h2>
              <div className="bg-red-50 border border-red-700 p-3 mb-4">
                <p className="text-sm text-red-800">{error}</p>
              </div>
              <button
                onClick={() => (window.location.href = "/")}
                className="w-full px-4 py-2 border border-black bg-red-700 text-white text-[12px] font-bold uppercase tracking-[1.5px] font-[Helvetica,Arial,system-ui,sans-serif] hover:bg-red-800 transition-colors"
              >
                Return Home
              </button>
            </div>
          </div>
        )}
      </>
    );
  }

  // Success - component should have navigated away
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 font-['Times_New_Roman',Times,serif]">
      <div className="text-green-700">
        <svg
          className="h-12 w-12"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M5 13l4 4L19 7"
          />
        </svg>
      </div>
      <p className="mt-4 text-sm text-gray-600">Sign in successful! Redirecting...</p>
    </div>
  );
}

export default OAuthCallback;