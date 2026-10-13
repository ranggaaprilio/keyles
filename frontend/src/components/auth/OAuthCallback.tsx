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
          <div className="flex flex-col items-center justify-center min-h-screen bg-muted">
            <div className="h-10 w-10 rounded-full border-4 border-border border-t-foreground animate-spin"></div>
            <p className="mt-4 text-sm text-muted-foreground">Completing sign in...</p>
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
          <div className="flex flex-col items-center justify-center min-h-screen bg-muted px-4">
            <div className="w-full max-w-md rounded-xl border border-border bg-background p-6 shadow-card sm:p-8">
              <h2 className="text-xl font-medium tracking-tight text-foreground mb-4">
                Authentication Failed
              </h2>
              <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive mb-6">
                <p>{error}</p>
              </div>
              <button
                onClick={() => (window.location.href = "/")}
                className="inline-flex h-10 w-full items-center justify-center rounded-full bg-primary px-5 text-sm font-medium text-primary-foreground hover:bg-primary/85 transition-colors duration-fast"
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
    <div className="flex flex-col items-center justify-center min-h-screen bg-muted">
      <div className="text-success">
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
      <p className="mt-4 text-sm text-muted-foreground">Sign in successful! Redirecting...</p>
    </div>
  );
}

export default OAuthCallback;