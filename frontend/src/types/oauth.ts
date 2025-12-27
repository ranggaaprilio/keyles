/**
 * OAuth 2.0 / OIDC Types
 * Types for OAuth authorization flow with PKCE support
 */

/**
 * Authorization request parameters for OAuth 2.0 Authorization Code flow
 */
export interface AuthorizationRequest {
  /** The OAuth client identifier */
  client_id: string;
  /** URI to redirect after authorization */
  redirect_uri: string;
  /** Must be "code" for Authorization Code flow */
  response_type: 'code';
  /** Space-separated list of scopes (must include "openid") */
  scope: string;
  /** CSRF protection token, will be returned in response */
  state: string;
  /** PKCE code challenge (base64url-encoded SHA256 hash of code_verifier) */
  code_challenge: string;
  /** Must be "S256" for our implementation (PKCE requirement) */
  code_challenge_method: 'S256';
  /** Optional nonce for ID token replay protection */
  nonce?: string;
  /** Optional prompt behavior: none, login, consent, select_account */
  prompt?: 'none' | 'login' | 'consent' | 'select_account';
  /** Optional max age of authentication in seconds */
  max_age?: number;
  /** Optional UI locales preference */
  ui_locales?: string;
  /** Optional login hint (email or username) */
  login_hint?: string;
}

/**
 * Authorization response from OAuth server
 * Returned via redirect URL query parameters
 */
export interface AuthorizationResponse {
  /** Authorization code (present on success) */
  code?: string;
  /** CSRF state from request (always present) */
  state: string;
  /** Error code (present on failure) */
  error?: AuthorizationErrorCode;
  /** Human-readable error description */
  error_description?: string;
  /** URI with more error information */
  error_uri?: string;
}

/**
 * OAuth 2.0 error codes for authorization endpoint
 */
export type AuthorizationErrorCode =
  | 'invalid_request'
  | 'unauthorized_client'
  | 'access_denied'
  | 'unsupported_response_type'
  | 'invalid_scope'
  | 'server_error'
  | 'temporarily_unavailable';

/**
 * Token request for exchanging authorization code
 */
export interface TokenRequest {
  /** Must be "authorization_code" */
  grant_type: 'authorization_code' | 'refresh_token';
  /** Authorization code from authorization response */
  code?: string;
  /** Must match the redirect_uri from authorization request */
  redirect_uri?: string;
  /** The OAuth client identifier */
  client_id: string;
  /** Client secret (for confidential clients) */
  client_secret?: string;
  /** PKCE code verifier (required) */
  code_verifier?: string;
  /** Refresh token (for refresh_token grant) */
  refresh_token?: string;
}

/**
 * Token response from OAuth server
 */
export interface TokenResponse {
  /** JWT access token */
  access_token: string;
  /** Token type (always "Bearer") */
  token_type: 'Bearer';
  /** Access token expiration in seconds */
  expires_in: number;
  /** JWT refresh token */
  refresh_token?: string;
  /** JWT ID token (OIDC) */
  id_token?: string;
  /** Granted scopes (if different from requested) */
  scope?: string;
}

/**
 * Token error response
 */
export interface TokenErrorResponse {
  /** Error code */
  error: TokenErrorCode;
  /** Human-readable error description */
  error_description?: string;
  /** URI with more error information */
  error_uri?: string;
}

/**
 * OAuth 2.0 error codes for token endpoint
 */
export type TokenErrorCode =
  | 'invalid_request'
  | 'invalid_client'
  | 'invalid_grant'
  | 'unauthorized_client'
  | 'unsupported_grant_type'
  | 'invalid_scope';

/**
 * Decoded ID Token claims (OIDC)
 */
export interface IDTokenClaims {
  /** Issuer identifier */
  iss: string;
  /** Subject identifier (user ID) */
  sub: string;
  /** Audience (client_id) */
  aud: string | string[];
  /** Expiration time (Unix timestamp) */
  exp: number;
  /** Issued at time (Unix timestamp) */
  iat: number;
  /** Authentication time (Unix timestamp) */
  auth_time?: number;
  /** Nonce value from authorization request */
  nonce?: string;
  /** Access Control (authentication context class reference) */
  acr?: string;
  /** Authentication Methods References */
  amr?: string[];
  /** Authorized party (client that requested the token) */
  azp?: string;
  /** Tenant identifier (custom claim) */
  tenant_id?: string;
  /** User's email address */
  email?: string;
  /** Whether email is verified */
  email_verified?: boolean;
  /** User's display name */
  name?: string;
  /** User's given name (first name) */
  given_name?: string;
  /** User's family name (last name) */
  family_name?: string;
  /** User's preferred username */
  preferred_username?: string;
  /** URL of user's profile picture */
  picture?: string;
  /** User's locale preference */
  locale?: string;
  /** User's timezone */
  zoneinfo?: string;
  /** Time when user info was last updated */
  updated_at?: number;
}

/**
 * Decoded Access Token claims
 */
export interface AccessTokenClaims {
  /** Issuer identifier */
  iss: string;
  /** Subject identifier (user ID) */
  sub: string;
  /** Audience */
  aud: string | string[];
  /** Expiration time (Unix timestamp) */
  exp: number;
  /** Issued at time (Unix timestamp) */
  iat: number;
  /** Not before time (Unix timestamp) */
  nbf?: number;
  /** JWT ID */
  jti?: string;
  /** Authorized scopes */
  scope?: string;
  /** Client ID */
  client_id?: string;
  /** Tenant ID (custom claim) */
  tenant_id?: string;
  /** User roles (custom claim) */
  roles?: string[];
}

/**
 * PKCE parameters stored during authorization flow
 */
export interface PKCEParams {
  /** Random verifier (43-128 characters) */
  code_verifier: string;
  /** SHA256 hash of verifier, base64url encoded */
  code_challenge: string;
  /** Challenge method (always S256) */
  code_challenge_method: 'S256';
}

/**
 * OAuth flow state stored in sessionStorage
 */
export interface OAuthState {
  /** CSRF state token */
  state: string;
  /** PKCE code verifier */
  code_verifier: string;
  /** Original redirect URI */
  redirect_uri: string;
  /** Nonce for ID token validation */
  nonce?: string;
  /** Timestamp when flow started */
  started_at: number;
}

/**
 * Client information displayed on consent screen
 */
export interface ClientInfo {
  /** Client ID */
  client_id: string;
  /** Client display name */
  client_name: string;
  /** Client logo URL */
  logo_uri?: string;
  /** Client policy URL */
  policy_uri?: string;
  /** Client terms of service URL */
  tos_uri?: string;
}

/**
 * Consent screen props
 */
export interface ConsentScreenProps {
  /** Client requesting access */
  client: ClientInfo;
  /** Requested scopes */
  scopes: string[];
  /** User's email or name */
  user: string;
  /** Called when user approves */
  onApprove: () => void;
  /** Called when user denies */
  onDeny: () => void;
  /** Loading state */
  isLoading?: boolean;
}

/**
 * OIDC Discovery document
 */
export interface OIDCDiscovery {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint?: string;
  jwks_uri: string;
  registration_endpoint?: string;
  scopes_supported?: string[];
  response_types_supported: string[];
  response_modes_supported?: string[];
  grant_types_supported?: string[];
  acr_values_supported?: string[];
  subject_types_supported: string[];
  id_token_signing_alg_values_supported: string[];
  token_endpoint_auth_methods_supported?: string[];
  claims_supported?: string[];
  code_challenge_methods_supported?: string[];
}

/**
 * JWKS (JSON Web Key Set) document
 */
export interface JWKS {
  keys: JWK[];
}

/**
 * JSON Web Key for RS256
 */
export interface JWK {
  /** Key type (RSA) */
  kty: 'RSA';
  /** Key ID */
  kid: string;
  /** Public key use (signature) */
  use: 'sig';
  /** Algorithm (RS256) */
  alg: 'RS256';
  /** RSA modulus (base64url) */
  n: string;
  /** RSA exponent (base64url) */
  e: string;
}
