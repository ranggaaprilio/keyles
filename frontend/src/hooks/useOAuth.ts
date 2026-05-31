/**
 * useOAuth Hook
 * React hook for managing OAuth 2.0 Authorization Code flow with PKCE
 */

import { useState, useCallback, useEffect, useMemo } from 'react';
import type {
  AuthorizationRequest,
  TokenResponse,
} from '../types/oauth';
import { OAuthService, OAuthConfig } from '../services/oauthService';

/**
 * Storage key for OAuth tokens
 */
const TOKENS_STORAGE_KEY = 'keyles_oauth_tokens';

/**
 * OAuth hook state
 */
interface UseOAuthState {
  /** Whether authentication is in progress */
  isLoading: boolean;
  /** Whether user is authenticated */
  isAuthenticated: boolean;
  /** Error message if any */
  error: string | null;
  /** Current tokens */
  tokens: TokenResponse | null;
}

/**
 * OAuth hook return value
 */
interface UseOAuthReturn extends UseOAuthState {
  /** Start the authorization flow */
  authorize: (options?: Partial<AuthorizationRequest>) => Promise<void>;
  /** Handle OAuth callback (call on callback page) */
  handleCallback: (callbackUrl?: string) => Promise<TokenResponse>;
  /** Refresh tokens */
  refreshTokens: () => Promise<TokenResponse>;
  /** Logout and clear tokens */
  logout: () => void;
  /** Get access token (returns null if expired) */
  getAccessToken: () => string | null;
  /** Get ID token */
  getIdToken: () => string | null;
  /** Get refresh token */
  getRefreshToken: () => string | null;
  /** OAuth service instance */
  service: OAuthService;
}

/**
 * Options for useOAuth hook
 */
interface UseOAuthOptions {
  /** OAuth configuration */
  config: OAuthConfig;
  /** Whether to persist tokens in localStorage */
  persistTokens?: boolean;
  /** Custom token storage key */
  storageKey?: string;
  /** Auto-refresh tokens before expiration */
  autoRefresh?: boolean;
  /** Seconds before expiration to trigger auto-refresh */
  refreshThreshold?: number;
}

/**
 * Custom hook for managing OAuth 2.0 Authorization Code flow with PKCE
 * 
 * @param options - OAuth configuration and options
 * @returns OAuth management functions and state
 * 
 * @example
 * ```tsx
 * function App() {
 *   const {
 *     isAuthenticated,
 *     isLoading,
 *     authorize,
 *     logout,
 *     getAccessToken,
 *   } = useOAuth({
 *     config: {
 *       issuer: 'https://sso.example.com',
 *       clientId: 'my-app',
 *       redirectUri: 'https://myapp.com/callback',
 *     },
 *   });
 * 
 *   if (isLoading) return <div>Loading...</div>;
 * 
 *   if (!isAuthenticated) {
 *     return <button onClick={() => authorize()}>Login</button>;
 *   }
 * 
 *   return (
 *     <div>
 *       <p>Welcome!</p>
 *       <button onClick={logout}>Logout</button>
 *     </div>
 *   );
 * }
 * ```
 */
export function useOAuth(options: UseOAuthOptions): UseOAuthReturn {
  const {
    config,
    persistTokens = true,
    storageKey = TOKENS_STORAGE_KEY,
    autoRefresh = true,
    refreshThreshold = 60, // Refresh 60 seconds before expiration
  } = options;

  // Create OAuth service instance
  const service = useMemo(() => new OAuthService(config), [config]);

  // Initialize state
  const [state, setState] = useState<UseOAuthState>(() => {
    // Try to load tokens from storage
    const storedTokens = persistTokens
      ? localStorage.getItem(storageKey)
      : null;

    let tokens: TokenResponse | null = null;
    if (storedTokens) {
      try {
        tokens = JSON.parse(storedTokens);
      } catch {
        localStorage.removeItem(storageKey);
      }
    }

    return {
      isLoading: false,
      isAuthenticated: tokens !== null,
      error: null,
      tokens,
    };
  });

  /**
   * Save tokens to storage
   */
  const saveTokens = useCallback(
    (tokens: TokenResponse) => {
      if (persistTokens) {
        localStorage.setItem(storageKey, JSON.stringify(tokens));
      }
      setState(prev => ({
        ...prev,
        tokens,
        isAuthenticated: true,
        error: null,
      }));
    },
    [persistTokens, storageKey]
  );

  /**
   * Clear tokens from storage
   */
  const clearTokens = useCallback(() => {
    localStorage.removeItem(storageKey);
    service.clearState();
    setState({
      isLoading: false,
      isAuthenticated: false,
      error: null,
      tokens: null,
    });
  }, [service, storageKey]);

  /**
   * Start the authorization flow
   */
  const authorize = useCallback(
    async (authOptions?: Partial<AuthorizationRequest>) => {
      setState(prev => ({ ...prev, isLoading: true, error: null }));

      try {
        await service.authorize(authOptions);
        // Note: This will redirect, so we won't reach here normally
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Authorization failed';
        setState(prev => ({
          ...prev,
          isLoading: false,
          error: errorMessage,
        }));
        throw err;
      }
    },
    [service]
  );

  /**
   * Handle OAuth callback
   */
  const handleCallback = useCallback(
    async (callbackUrl?: string): Promise<TokenResponse> => {
      setState(prev => ({ ...prev, isLoading: true, error: null }));

      try {
        const tokens = await service.completeCallback(callbackUrl);
        saveTokens(tokens);
        return tokens;
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Callback handling failed';
        setState(prev => ({
          ...prev,
          isLoading: false,
          error: errorMessage,
        }));
        throw err;
      }
    },
    [service, saveTokens]
  );

  /**
   * Refresh tokens
   */
  const refreshTokensFunc = useCallback(async (): Promise<TokenResponse> => {
    if (!state.tokens?.refresh_token) {
      throw new Error('No refresh token available');
    }

    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      const refreshedTokens = await service.refreshTokens(state.tokens.refresh_token);
      const tokens: TokenResponse = {
        ...state.tokens,
        ...refreshedTokens,
        refresh_token: refreshedTokens.refresh_token ?? state.tokens.refresh_token,
      };
      saveTokens(tokens);
      return tokens;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Token refresh failed';
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: errorMessage,
      }));
      // If refresh fails, clear tokens and require re-authentication
      clearTokens();
      throw err;
    }
  }, [service, state.tokens, saveTokens, clearTokens]);

  /**
   * Logout
   */
  const logout = useCallback(() => {
    clearTokens();
  }, [clearTokens]);

  /**
   * Get access token (checks expiration)
   */
  const getAccessToken = useCallback((): string | null => {
    if (!state.tokens?.access_token) {
      return null;
    }

    // Try to decode token and check expiration
    try {
      const parts = state.tokens.access_token.split('.');
      if (parts.length !== 3) {
        return state.tokens.access_token;
      }

      const payload = JSON.parse(atob(parts[1]!));
      const exp = payload.exp;

      if (exp && Date.now() / 1000 > exp) {
        // Token is expired
        return null;
      }

      return state.tokens.access_token;
    } catch {
      // If we can't decode, just return the token
      return state.tokens.access_token;
    }
  }, [state.tokens]);

  /**
   * Get ID token
   */
  const getIdToken = useCallback((): string | null => {
    return state.tokens?.id_token || null;
  }, [state.tokens]);

  /**
   * Get refresh token
   */
  const getRefreshToken = useCallback((): string | null => {
    return state.tokens?.refresh_token || null;
  }, [state.tokens]);

  /**
   * Auto-refresh tokens before expiration
   */
  useEffect(() => {
    if (!autoRefresh || !state.tokens?.access_token) {
      return;
    }

    let timeoutId: NodeJS.Timeout | null = null;

    try {
      const parts = state.tokens.access_token.split('.');
      if (parts.length !== 3) {
        return;
      }

      const payload = JSON.parse(atob(parts[1]!));
      const exp = payload.exp;

      if (!exp) {
        return;
      }

      // Calculate time until refresh needed
      const expiresAt = exp * 1000;
      const refreshAt = expiresAt - refreshThreshold * 1000;
      const timeUntilRefresh = refreshAt - Date.now();

      if (timeUntilRefresh > 0) {
        timeoutId = setTimeout(() => {
          refreshTokensFunc().catch(() => {
            // Error handling is done in the refresh function
          });
        }, timeUntilRefresh);
      } else if (state.tokens.refresh_token) {
        // Token is already expired or about to expire, refresh now
        refreshTokensFunc().catch(() => {
          // Error handling is done in the refresh function
        });
      }
    } catch {
      // Can't decode token, skip auto-refresh
    }

    return () => {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
    };
  }, [autoRefresh, state.tokens, refreshThreshold, refreshTokensFunc]);

  return {
    isLoading: state.isLoading,
    isAuthenticated: state.isAuthenticated,
    error: state.error,
    tokens: state.tokens,
    authorize,
    handleCallback,
    refreshTokens: refreshTokensFunc,
    logout,
    getAccessToken,
    getIdToken,
    getRefreshToken,
    service,
  };
}

export default useOAuth;
