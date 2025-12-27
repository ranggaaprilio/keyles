/**
 * usePKCE Hook
 * React hook for managing PKCE parameters in OAuth flows
 */

import { useState, useCallback, useEffect } from 'react';
import type { PKCEParams } from '../types/oauth';
import { generatePKCE, isValidCodeVerifier, verifyPKCE } from '../utils/pkce';

/**
 * Storage key for PKCE parameters
 */
const PKCE_STORAGE_KEY = 'keyles_pkce_params';

/**
 * PKCE hook state
 */
interface UsePKCEState {
  /** PKCE parameters (verifier and challenge) */
  params: PKCEParams | null;
  /** Whether PKCE is being generated */
  isGenerating: boolean;
  /** Error message if generation failed */
  error: string | null;
}

/**
 * PKCE hook return value
 */
interface UsePKCEReturn extends UsePKCEState {
  /** Generate new PKCE parameters */
  generate: () => Promise<PKCEParams>;
  /** Clear stored PKCE parameters */
  clear: () => void;
  /** Get stored code verifier */
  getStoredVerifier: () => string | null;
  /** Verify a code challenge matches a verifier */
  verify: (codeVerifier: string, codeChallenge: string) => Promise<boolean>;
  /** Check if verifier format is valid */
  isValid: (verifier: string) => boolean;
}

/**
 * Custom hook for managing PKCE parameters
 * 
 * @param autoGenerate - Whether to automatically generate PKCE on mount
 * @param storageKey - Optional custom storage key
 * @returns PKCE management functions and state
 * 
 * @example
 * ```tsx
 * function LoginButton() {
 *   const { params, generate, isGenerating } = usePKCE();
 * 
 *   const handleLogin = async () => {
 *     const pkce = await generate();
 *     // Store verifier and redirect with challenge
 *   };
 * 
 *   return (
 *     <button onClick={handleLogin} disabled={isGenerating}>
 *       {isGenerating ? 'Preparing...' : 'Login'}
 *     </button>
 *   );
 * }
 * ```
 */
export function usePKCE(
  autoGenerate: boolean = false,
  storageKey: string = PKCE_STORAGE_KEY
): UsePKCEReturn {
  const [state, setState] = useState<UsePKCEState>({
    params: null,
    isGenerating: false,
    error: null,
  });

  /**
   * Generate new PKCE parameters
   */
  const generate = useCallback(async (): Promise<PKCEParams> => {
    setState(prev => ({ ...prev, isGenerating: true, error: null }));

    try {
      const params = await generatePKCE();

      // Store verifier in sessionStorage
      sessionStorage.setItem(storageKey, params.code_verifier);

      setState({
        params,
        isGenerating: false,
        error: null,
      });

      return params;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to generate PKCE';
      setState({
        params: null,
        isGenerating: false,
        error: errorMessage,
      });
      throw err;
    }
  }, [storageKey]);

  /**
   * Clear stored PKCE parameters
   */
  const clear = useCallback(() => {
    sessionStorage.removeItem(storageKey);
    setState({
      params: null,
      isGenerating: false,
      error: null,
    });
  }, [storageKey]);

  /**
   * Get stored code verifier
   */
  const getStoredVerifier = useCallback((): string | null => {
    return sessionStorage.getItem(storageKey);
  }, [storageKey]);

  /**
   * Verify a code challenge matches a verifier
   */
  const verify = useCallback(
    async (codeVerifier: string, codeChallenge: string): Promise<boolean> => {
      return verifyPKCE(codeVerifier, codeChallenge);
    },
    []
  );

  /**
   * Check if verifier format is valid
   */
  const isValid = useCallback((verifier: string): boolean => {
    return isValidCodeVerifier(verifier);
  }, []);

  /**
   * Auto-generate PKCE on mount if requested
   */
  useEffect(() => {
    if (autoGenerate) {
      generate().catch(() => {
        // Error is already captured in state
      });
    }
  }, [autoGenerate, generate]);

  return {
    params: state.params,
    isGenerating: state.isGenerating,
    error: state.error,
    generate,
    clear,
    getStoredVerifier,
    verify,
    isValid,
  };
}

export default usePKCE;
