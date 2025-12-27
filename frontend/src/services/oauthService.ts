/**
 * OAuth 2.0 Service
 * Handles OAuth authorization flow with PKCE support
 */

import type {
  AuthorizationRequest,
  AuthorizationResponse,
  TokenRequest,
  TokenResponse,
  TokenErrorResponse,
  OAuthState,
  OIDCDiscovery,
  JWKS,
} from '../types/oauth';
import { generatePKCE, generateState, generateNonce } from '../utils/pkce';

/**
 * Storage key for OAuth state
 */
const OAUTH_STATE_KEY = 'keyles_oauth_state';

/**
 * OAuth service configuration
 */
export interface OAuthConfig {
  /** Authorization server base URL */
  issuer: string;
  /** OAuth client identifier */
  clientId: string;
  /** Default redirect URI */
  redirectUri: string;
  /** Default scopes to request */
  defaultScopes?: string[];
}

/**
 * OAuthService class for handling OAuth 2.0 Authorization Code flow with PKCE
 */
export class OAuthService {
  private config: OAuthConfig;
  private discoveryDoc: OIDCDiscovery | null = null;

  constructor(config: OAuthConfig) {
    this.config = {
      ...config,
      defaultScopes: config.defaultScopes || ['openid', 'profile', 'email'],
    };
  }

  /**
   * Fetches OIDC discovery document
   */
  async fetchDiscovery(): Promise<OIDCDiscovery> {
    if (this.discoveryDoc) {
      return this.discoveryDoc;
    }

    const response = await fetch(
      `${this.config.issuer}/.well-known/openid-configuration`
    );

    if (!response.ok) {
      throw new Error('Failed to fetch OIDC discovery document');
    }

    this.discoveryDoc = await response.json();
    return this.discoveryDoc!;
  }

  /**
   * Fetches JWKS (JSON Web Key Set) for token validation
   */
  async fetchJWKS(): Promise<JWKS> {
    const discovery = await this.fetchDiscovery();
    const response = await fetch(discovery.jwks_uri);

    if (!response.ok) {
      throw new Error('Failed to fetch JWKS');
    }

    return response.json();
  }

  /**
   * Builds the authorization URL and returns it along with state to store
   * 
   * @param options - Optional overrides for authorization request
   * @returns Authorization URL and OAuth state to store
   */
  async buildAuthURL(options?: Partial<AuthorizationRequest>): Promise<{
    url: string;
    state: OAuthState;
  }> {
    const pkce = await generatePKCE();
    const state = generateState();
    const nonce = generateNonce();

    const redirectUri = options?.redirect_uri || this.config.redirectUri;
    const scopes = options?.scope || this.config.defaultScopes!.join(' ');

    const params: AuthorizationRequest = {
      client_id: options?.client_id || this.config.clientId,
      redirect_uri: redirectUri,
      response_type: 'code',
      scope: scopes,
      state,
      code_challenge: pkce.code_challenge,
      code_challenge_method: 'S256',
      nonce,
      ...options,
    };

    // Build URL
    const url = new URL(`${this.config.issuer}/oauth2/auth`);
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        url.searchParams.set(key, String(value));
      }
    });

    // Return state to be stored by caller
    const oauthState: OAuthState = {
      state,
      code_verifier: pkce.code_verifier,
      redirect_uri: redirectUri,
      nonce,
      started_at: Date.now(),
    };

    return {
      url: url.toString(),
      state: oauthState,
    };
  }

  /**
   * Initiates the authorization flow by redirecting to the authorization endpoint
   * Stores PKCE verifier and state in sessionStorage
   * 
   * @param options - Optional overrides for authorization request
   */
  async authorize(options?: Partial<AuthorizationRequest>): Promise<void> {
    const { url, state } = await this.buildAuthURL(options);

    // Store state in sessionStorage
    sessionStorage.setItem(OAUTH_STATE_KEY, JSON.stringify(state));

    // Redirect to authorization endpoint
    window.location.href = url;
  }

  /**
   * Handles the OAuth callback by parsing URL parameters
   * Validates state and returns authorization response
   * 
   * @param callbackUrl - The callback URL with query parameters (default: current URL)
   * @returns Authorization response with code or error
   * @throws Error if state mismatch or no stored state
   */
  handleCallback(callbackUrl?: string): AuthorizationResponse {
    const url = new URL(callbackUrl || window.location.href);
    const params = url.searchParams;

    const code = params.get('code');
    const error = params.get('error') as AuthorizationResponse['error'];
    const errorDescription = params.get('error_description');
    const errorUri = params.get('error_uri');

    const response: AuthorizationResponse = {
      state: params.get('state') || '',
      ...(code && { code }),
      ...(error && { error }),
      ...(errorDescription && { error_description: errorDescription }),
      ...(errorUri && { error_uri: errorUri }),
    };

    return response;
  }

  /**
   * Validates the callback state against stored state
   * 
   * @param receivedState - State from callback URL
   * @returns Stored OAuth state if valid
   * @throws Error if state mismatch or expired
   */
  validateState(receivedState: string): OAuthState {
    const storedStateJson = sessionStorage.getItem(OAUTH_STATE_KEY);
    
    if (!storedStateJson) {
      throw new Error('No stored OAuth state found');
    }

    const storedState: OAuthState = JSON.parse(storedStateJson);

    if (storedState.state !== receivedState) {
      throw new Error('State mismatch - possible CSRF attack');
    }

    // Check if flow has expired (15 minutes max)
    const maxAge = 15 * 60 * 1000; // 15 minutes
    if (Date.now() - storedState.started_at > maxAge) {
      sessionStorage.removeItem(OAUTH_STATE_KEY);
      throw new Error('OAuth flow has expired');
    }

    return storedState;
  }

  /**
   * Exchanges authorization code for tokens
   * 
   * @param code - Authorization code from callback
   * @param codeVerifier - PKCE code verifier (from stored state)
   * @returns Token response with access_token, refresh_token, id_token
   * @throws Error if token exchange fails
   */
  async exchangeCodeForTokens(
    code: string,
    codeVerifier: string
  ): Promise<TokenResponse> {
    const tokenRequest: TokenRequest = {
      grant_type: 'authorization_code',
      code,
      redirect_uri: this.config.redirectUri,
      client_id: this.config.clientId,
      code_verifier: codeVerifier,
    };

    // Filter out undefined values and convert to URLSearchParams
    const params = new URLSearchParams();
    Object.entries(tokenRequest).forEach(([key, value]) => {
      if (value !== undefined) {
        params.set(key, String(value));
      }
    });

    const response = await fetch(`${this.config.issuer}/oauth2/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params,
    });

    if (!response.ok) {
      const errorResponse: TokenErrorResponse = await response.json();
      throw new Error(
        errorResponse.error_description || errorResponse.error || 'Token exchange failed'
      );
    }

    return response.json();
  }

  /**
   * Complete OAuth callback handling - validates state, exchanges code for tokens
   * 
   * @param callbackUrl - The callback URL (default: current URL)
   * @returns Token response
   * @throws Error on validation failure or token exchange error
   */
  async completeCallback(callbackUrl?: string): Promise<TokenResponse> {
    const authResponse = this.handleCallback(callbackUrl);

    // Check for errors
    if (authResponse.error) {
      throw new Error(
        authResponse.error_description || authResponse.error
      );
    }

    if (!authResponse.code) {
      throw new Error('No authorization code in callback');
    }

    // Validate state
    const storedState = this.validateState(authResponse.state);

    // Exchange code for tokens
    const tokens = await this.exchangeCodeForTokens(
      authResponse.code,
      storedState.code_verifier
    );

    // Clear stored state
    sessionStorage.removeItem(OAUTH_STATE_KEY);

    return tokens;
  }

  /**
   * Refreshes tokens using a refresh token
   * 
   * @param refreshToken - The refresh token
   * @returns New token response
   */
  async refreshTokens(refreshToken: string): Promise<TokenResponse> {
    const tokenRequest: TokenRequest = {
      grant_type: 'refresh_token',
      refresh_token: refreshToken,
      client_id: this.config.clientId,
    };

    // Filter out undefined values and convert to URLSearchParams
    const params = new URLSearchParams();
    Object.entries(tokenRequest).forEach(([key, value]) => {
      if (value !== undefined) {
        params.set(key, String(value));
      }
    });

    const response = await fetch(`${this.config.issuer}/oauth2/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params,
    });

    if (!response.ok) {
      const errorResponse: TokenErrorResponse = await response.json();
      throw new Error(
        errorResponse.error_description || errorResponse.error || 'Token refresh failed'
      );
    }

    return response.json();
  }

  /**
   * Clears stored OAuth state
   */
  clearState(): void {
    sessionStorage.removeItem(OAUTH_STATE_KEY);
  }

  /**
   * Checks if there's a pending OAuth flow
   */
  hasPendingFlow(): boolean {
    return sessionStorage.getItem(OAUTH_STATE_KEY) !== null;
  }
}

/**
 * Create a default OAuth service instance
 */
export function createOAuthService(config: OAuthConfig): OAuthService {
  return new OAuthService(config);
}

/**
 * Singleton instance for the default OAuth service
 */
let defaultOAuthService: OAuthService | null = null;

/**
 * Initialize the default OAuth service
 */
export function initOAuthService(config: OAuthConfig): OAuthService {
  defaultOAuthService = new OAuthService(config);
  return defaultOAuthService;
}

/**
 * Get the default OAuth service instance
 */
export function getOAuthService(): OAuthService {
  if (!defaultOAuthService) {
    throw new Error('OAuth service not initialized. Call initOAuthService first.');
  }
  return defaultOAuthService;
}
