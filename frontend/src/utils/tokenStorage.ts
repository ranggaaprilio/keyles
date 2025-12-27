/**
 * Secure token storage utility for OAuth tokens
 * Implements memory-based storage with optional persistence to sessionStorage
 * 
 * Security considerations:
 * - Tokens are stored in memory by default (cleared on page refresh)
 * - SessionStorage used for persistence (cleared on tab close)
 * - Never use localStorage for tokens (persists across sessions)
 * - Access tokens expire in 15 minutes, refresh tokens in 7 days
 */

export interface TokenSet {
  accessToken: string;
  idToken?: string;
  refreshToken?: string;
  tokenType: string;
  expiresIn: number;
  scope?: string;
  issuedAt: number; // Unix timestamp when tokens were stored
}

// Memory storage (most secure - cleared on page refresh)
let inMemoryTokens: TokenSet | null = null;

// Storage keys for sessionStorage (optional persistence)
const STORAGE_PREFIX = 'keyles_oauth_';
const ACCESS_TOKEN_KEY = `${STORAGE_PREFIX}access_token`;
const ID_TOKEN_KEY = `${STORAGE_PREFIX}id_token`;
const REFRESH_TOKEN_KEY = `${STORAGE_PREFIX}refresh_token`;
const TOKEN_METADATA_KEY = `${STORAGE_PREFIX}metadata`;

export interface TokenMetadata {
  tokenType: string;
  expiresIn: number;
  scope?: string;
  issuedAt: number;
}

/**
 * Store OAuth tokens
 * @param tokens - The token set from the OAuth token response
 * @param persist - Whether to persist to sessionStorage (default: false - memory only)
 */
export function storeTokens(tokens: TokenSet, persist: boolean = false): void {
  // Always store in memory
  inMemoryTokens = {
    ...tokens,
    issuedAt: tokens.issuedAt || Date.now(),
  };

  if (persist) {
    try {
      sessionStorage.setItem(ACCESS_TOKEN_KEY, tokens.accessToken);
      if (tokens.idToken) {
        sessionStorage.setItem(ID_TOKEN_KEY, tokens.idToken);
      }
      if (tokens.refreshToken) {
        sessionStorage.setItem(REFRESH_TOKEN_KEY, tokens.refreshToken);
      }
      sessionStorage.setItem(TOKEN_METADATA_KEY, JSON.stringify({
        tokenType: tokens.tokenType,
        expiresIn: tokens.expiresIn,
        scope: tokens.scope,
        issuedAt: inMemoryTokens.issuedAt,
      } as TokenMetadata));
    } catch (error) {
      console.warn('Failed to persist tokens to sessionStorage:', error);
    }
  }
}

/**
 * Get the current access token
 * @returns The access token or null if not available or expired
 */
export function getAccessToken(): string | null {
  // First check memory
  if (inMemoryTokens) {
    if (isAccessTokenExpired()) {
      return null;
    }
    return inMemoryTokens.accessToken;
  }

  // Fall back to sessionStorage
  try {
    const accessToken = sessionStorage.getItem(ACCESS_TOKEN_KEY);
    if (accessToken) {
      // Restore to memory cache
      restoreFromSessionStorage();
      if (isAccessTokenExpired()) {
        return null;
      }
      return accessToken;
    }
  } catch (error) {
    console.warn('Failed to read access token from sessionStorage:', error);
  }

  return null;
}

/**
 * Get the current ID token
 * @returns The ID token or null if not available
 */
export function getIdToken(): string | null {
  if (inMemoryTokens?.idToken) {
    return inMemoryTokens.idToken;
  }

  try {
    return sessionStorage.getItem(ID_TOKEN_KEY);
  } catch (error) {
    console.warn('Failed to read ID token from sessionStorage:', error);
  }

  return null;
}

/**
 * Get the current refresh token
 * @returns The refresh token or null if not available
 */
export function getRefreshToken(): string | null {
  if (inMemoryTokens?.refreshToken) {
    return inMemoryTokens.refreshToken;
  }

  try {
    return sessionStorage.getItem(REFRESH_TOKEN_KEY);
  } catch (error) {
    console.warn('Failed to read refresh token from sessionStorage:', error);
  }

  return null;
}

/**
 * Get all current tokens
 * @returns The token set or null if not available
 */
export function getTokens(): TokenSet | null {
  if (inMemoryTokens) {
    return structuredClone(inMemoryTokens);
  }

  restoreFromSessionStorage();
  return inMemoryTokens ? structuredClone(inMemoryTokens) : null;
}

/**
 * Clear all stored tokens (used for logout)
 */
export function clearTokens(): void {
  // Clear memory
  inMemoryTokens = null;

  // Clear sessionStorage
  try {
    sessionStorage.removeItem(ACCESS_TOKEN_KEY);
    sessionStorage.removeItem(ID_TOKEN_KEY);
    sessionStorage.removeItem(REFRESH_TOKEN_KEY);
    sessionStorage.removeItem(TOKEN_METADATA_KEY);
  } catch (error) {
    console.warn('Failed to clear tokens from sessionStorage:', error);
  }
}

/**
 * Check if the access token is expired
 * @param bufferSeconds - Number of seconds before actual expiry to consider it expired (default: 60)
 * @returns true if expired or about to expire
 */
export function isAccessTokenExpired(bufferSeconds: number = 60): boolean {
  const tokens = inMemoryTokens || restoreFromSessionStorage();
  if (!tokens) {
    return true;
  }

  const expiryTime = tokens.issuedAt + (tokens.expiresIn * 1000);
  const bufferMs = bufferSeconds * 1000;
  return Date.now() >= (expiryTime - bufferMs);
}

/**
 * Check if we have valid tokens (not expired)
 * @returns true if we have valid, non-expired tokens
 */
export function hasValidTokens(): boolean {
  const accessToken = getAccessToken();
  return accessToken !== null;
}

/**
 * Check if we have a refresh token available
 * @returns true if refresh token is available
 */
export function canRefreshToken(): boolean {
  return getRefreshToken() !== null;
}

/**
 * Get token expiry information
 * @returns Object with expiry details or null if no tokens
 */
export function getTokenExpiry(): { expiresAt: Date; expiresIn: number; isExpired: boolean } | null {
  const tokens = getTokens();
  if (!tokens) {
    return null;
  }

  const expiresAt = new Date(tokens.issuedAt + (tokens.expiresIn * 1000));
  const expiresIn = Math.max(0, Math.floor((expiresAt.getTime() - Date.now()) / 1000));
  const isExpired = expiresIn <= 0;

  return { expiresAt, expiresIn, isExpired };
}

/**
 * Parse JWT payload without verification (for extracting claims)
 * Note: This does NOT validate the token - only use for display purposes
 * @param token - JWT token string
 * @returns Parsed payload or null if invalid format
 */
export function parseJwtPayload<T = Record<string, unknown>>(token: string): T | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      return null;
    }
    const payload = parts[1]!;
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(decoded) as T;
  } catch {
    return null;
  }
}

/**
 * Get user info from ID token
 * @returns User claims from ID token or null
 */
export function getUserFromIdToken(): Record<string, unknown> | null {
  const idToken = getIdToken();
  if (!idToken) {
    return null;
  }
  return parseJwtPayload(idToken);
}

// Internal helper to restore tokens from sessionStorage to memory cache
function restoreFromSessionStorage(): TokenSet | null {
  try {
    const accessToken = sessionStorage.getItem(ACCESS_TOKEN_KEY);
    const metadataStr = sessionStorage.getItem(TOKEN_METADATA_KEY);
    
    if (!accessToken || !metadataStr) {
      return null;
    }

    const metadata = JSON.parse(metadataStr) as TokenMetadata;
    const idToken = sessionStorage.getItem(ID_TOKEN_KEY);
    const refreshToken = sessionStorage.getItem(REFRESH_TOKEN_KEY);
    
    const tokens: TokenSet = {
      accessToken,
      tokenType: metadata.tokenType,
      expiresIn: metadata.expiresIn,
      issuedAt: metadata.issuedAt,
    };
    
    if (idToken) {
      tokens.idToken = idToken;
    }
    if (refreshToken) {
      tokens.refreshToken = refreshToken;
    }
    if (metadata.scope) {
      tokens.scope = metadata.scope;
    }
    
    inMemoryTokens = tokens;

    return inMemoryTokens;
  } catch (error) {
    console.warn('Failed to restore tokens from sessionStorage:', error);
    return null;
  }
}
