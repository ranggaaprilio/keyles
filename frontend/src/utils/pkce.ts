/**
 * PKCE (Proof Key for Code Exchange) Utilities
 * Implements RFC 7636 for OAuth 2.0 public clients
 * 
 * This implementation uses the S256 method which:
 * 1. Generates a cryptographically random code_verifier (43-128 chars)
 * 2. Creates code_challenge = BASE64URL(SHA256(code_verifier))
 */

import type { PKCEParams } from '../types/oauth';

/**
 * Characters allowed in code_verifier per RFC 7636
 * unreserved characters: [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
 */
const ALLOWED_CHARS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';

/**
 * Default length for code_verifier (must be 43-128 characters)
 */
const CODE_VERIFIER_LENGTH = 64;

/**
 * Generates a cryptographically random code_verifier
 * Uses Web Crypto API for secure random number generation
 * 
 * @param length - Length of the verifier (43-128, default 64)
 * @returns Random code_verifier string
 * @throws Error if Web Crypto API is not available
 */
export function generateCodeVerifier(length: number = CODE_VERIFIER_LENGTH): string {
  if (length < 43 || length > 128) {
    throw new Error('Code verifier length must be between 43 and 128 characters');
  }

  if (typeof crypto === 'undefined' || !crypto.getRandomValues) {
    throw new Error('Web Crypto API is required for secure PKCE generation');
  }

  const randomValues = new Uint8Array(length);
  crypto.getRandomValues(randomValues);

  let verifier = '';
  for (let i = 0; i < length; i++) {
    const randomValue = randomValues[i]!;
    verifier += ALLOWED_CHARS[randomValue % ALLOWED_CHARS.length];
  }

  return verifier;
}

/**
 * Generates a code_challenge from a code_verifier using S256 method
 * code_challenge = BASE64URL(SHA256(ASCII(code_verifier)))
 * 
 * @param codeVerifier - The code_verifier to hash
 * @returns Promise resolving to base64url-encoded SHA256 hash
 * @throws Error if Web Crypto API is not available
 */
export async function generateCodeChallenge(codeVerifier: string): Promise<string> {
  if (typeof crypto === 'undefined' || !crypto.subtle) {
    throw new Error('Web Crypto API (SubtleCrypto) is required for PKCE S256 method');
  }

  // Convert verifier to UTF-8 bytes
  const encoder = new TextEncoder();
  const data = encoder.encode(codeVerifier);

  // Generate SHA-256 hash
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);

  // Convert to base64url
  return base64URLEncode(hashBuffer);
}

/**
 * Encodes an ArrayBuffer to base64url format
 * Base64url encoding differs from standard base64:
 * - Uses '-' instead of '+'
 * - Uses '_' instead of '/'
 * - No padding '='
 * 
 * @param buffer - ArrayBuffer to encode
 * @returns base64url-encoded string
 */
export function base64URLEncode(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]!);
  }

  // Convert to base64
  const base64 = btoa(binary);

  // Convert to base64url (RFC 4648 Section 5)
  return base64
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

/**
 * Decodes a base64url string to ArrayBuffer
 * 
 * @param base64url - base64url-encoded string
 * @returns ArrayBuffer containing decoded bytes
 */
export function base64URLDecode(base64url: string): ArrayBuffer {
  // Convert from base64url to base64
  let base64 = base64url
    .replace(/-/g, '+')
    .replace(/_/g, '/');

  // Add padding if necessary
  const padding = base64.length % 4;
  if (padding) {
    base64 += '='.repeat(4 - padding);
  }

  // Decode base64
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }

  return bytes.buffer;
}

/**
 * Generates complete PKCE parameters for OAuth authorization
 * 
 * @returns Promise resolving to PKCEParams object
 * @example
 * const pkce = await generatePKCE();
 * // Send pkce.code_challenge to authorization endpoint
 * // Store pkce.code_verifier in sessionStorage for token exchange
 */
export async function generatePKCE(): Promise<PKCEParams> {
  const code_verifier = generateCodeVerifier();
  const code_challenge = await generateCodeChallenge(code_verifier);

  return {
    code_verifier,
    code_challenge,
    code_challenge_method: 'S256',
  };
}

/**
 * Validates a code_verifier format
 * 
 * @param verifier - String to validate
 * @returns true if valid code_verifier format
 */
export function isValidCodeVerifier(verifier: string): boolean {
  if (typeof verifier !== 'string') {
    return false;
  }

  if (verifier.length < 43 || verifier.length > 128) {
    return false;
  }

  // Check all characters are in allowed set
  for (const char of verifier) {
    if (!ALLOWED_CHARS.includes(char)) {
      return false;
    }
  }

  return true;
}

/**
 * Validates that a code_challenge matches a code_verifier
 * Useful for testing/debugging PKCE implementation
 * 
 * @param codeVerifier - The original code_verifier
 * @param codeChallenge - The code_challenge to verify
 * @returns Promise resolving to true if challenge matches verifier
 */
export async function verifyPKCE(
  codeVerifier: string,
  codeChallenge: string
): Promise<boolean> {
  if (!isValidCodeVerifier(codeVerifier)) {
    return false;
  }

  const expectedChallenge = await generateCodeChallenge(codeVerifier);
  return expectedChallenge === codeChallenge;
}

/**
 * Generates a random state parameter for CSRF protection
 * 
 * @param length - Length of the state string (default 32)
 * @returns Random state string
 */
export function generateState(length: number = 32): string {
  const randomValues = new Uint8Array(length);
  crypto.getRandomValues(randomValues);

  let state = '';
  for (let i = 0; i < length; i++) {
    const randomValue = randomValues[i]!;
    state += ALLOWED_CHARS[randomValue % ALLOWED_CHARS.length];
  }

  return state;
}

/**
 * Generates a nonce for ID token replay protection
 * 
 * @param length - Length of the nonce string (default 32)
 * @returns Random nonce string
 */
export function generateNonce(length: number = 32): string {
  return generateState(length);
}
