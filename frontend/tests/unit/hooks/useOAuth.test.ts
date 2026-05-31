/**
 * useOAuth Hook Tests
 * Tests for OAuth authorization hook functionality
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useOAuth } from '../../../src/hooks/useOAuth';
import type { OAuthConfig } from '../../../src/services/oauthService';

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock crypto
const mockCrypto = {
  getRandomValues: (arr: Uint8Array) => {
    for (let i = 0; i < arr.length; i++) {
      arr[i] = Math.floor(Math.random() * 256);
    }
    return arr;
  },
  subtle: {
    digest: vi.fn().mockResolvedValue(new ArrayBuffer(32)),
  },
};
Object.defineProperty(global, 'crypto', { value: mockCrypto });

// Mock window.location
const mockLocation = {
  href: 'http://localhost:3000',
  origin: 'http://localhost:3000',
};
Object.defineProperty(window, 'location', {
  value: mockLocation,
  writable: true,
});

describe('useOAuth', () => {
  const mockConfig: OAuthConfig = {
    issuer: 'https://sso.example.com',
    clientId: 'test-client-id',
    redirectUri: 'http://localhost:3000/callback',
    defaultScopes: ['openid', 'profile', 'email'],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    sessionStorage.clear();
    mockLocation.href = 'http://localhost:3000';
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('initialization', () => {
    it('initializes with unauthenticated state', () => {
      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.tokens).toBeNull();
    });

    it('loads tokens from localStorage on init', () => {
      const tokens = {
        access_token: 'test-access-token',
        token_type: 'Bearer' as const,
        expires_in: 3600,
        refresh_token: 'test-refresh-token',
      };
      localStorage.setItem('keyles_oauth_tokens', JSON.stringify(tokens));

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.isAuthenticated).toBe(true);
      expect(result.current.tokens).toEqual(tokens);
    });

    it('handles corrupted localStorage tokens', () => {
      localStorage.setItem('keyles_oauth_tokens', 'invalid-json');

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.tokens).toBeNull();
    });
  });

  describe('authorize', () => {
    it('redirects to authorization endpoint', async () => {
      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await act(async () => {
        await result.current.authorize();
      });

      // Should have redirected
      expect(mockLocation.href).toContain('https://sso.example.com/oauth2/auth');
      expect(mockLocation.href).toContain('client_id=test-client-id');
      expect(mockLocation.href).toContain('response_type=code');
      expect(mockLocation.href).toContain('code_challenge_method=S256');
    });

    it('includes PKCE parameters', async () => {
      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await act(async () => {
        await result.current.authorize();
      });

      expect(mockLocation.href).toContain('code_challenge=');
      expect(mockLocation.href).toContain('state=');
    });

    it('stores OAuth state in sessionStorage', async () => {
      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await act(async () => {
        await result.current.authorize();
      });

      const storedState = sessionStorage.getItem('keyles_oauth_state');
      expect(storedState).not.toBeNull();

      const state = JSON.parse(storedState!);
      expect(state.code_verifier).toBeDefined();
      expect(state.state).toBeDefined();
      expect(state.redirect_uri).toBe(mockConfig.redirectUri);
    });

    it('accepts custom scopes', async () => {
      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await act(async () => {
        await result.current.authorize({ scope: 'openid custom_scope' });
      });

      expect(mockLocation.href).toContain('scope=openid+custom_scope');
    });
  });

  describe('handleCallback', () => {
    const mockTokenResponse = {
      access_token: 'new-access-token',
      token_type: 'Bearer' as const,
      expires_in: 3600,
      refresh_token: 'new-refresh-token',
      id_token: 'new-id-token',
    };

    beforeEach(() => {
      // Setup stored state
      const state = {
        state: 'test-state',
        code_verifier: 'test-verifier',
        redirect_uri: mockConfig.redirectUri,
        nonce: 'test-nonce',
        started_at: Date.now(),
      };
      sessionStorage.setItem('keyles_oauth_state', JSON.stringify(state));

      // Mock token endpoint
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockTokenResponse),
      });
    });

    it('exchanges code for tokens', async () => {
      mockLocation.href = 'http://localhost:3000/callback?code=test-code&state=test-state';

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      let tokens;
      await act(async () => {
        tokens = await result.current.handleCallback();
      });

      expect(tokens).toEqual(mockTokenResponse);
      expect(result.current.isAuthenticated).toBe(true);
    });

    it('saves tokens to localStorage', async () => {
      mockLocation.href = 'http://localhost:3000/callback?code=test-code&state=test-state';

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await act(async () => {
        await result.current.handleCallback();
      });

      const storedTokens = localStorage.getItem('keyles_oauth_tokens');
      expect(storedTokens).not.toBeNull();
      expect(JSON.parse(storedTokens!)).toEqual(mockTokenResponse);
    });

    it('throws on state mismatch', async () => {
      mockLocation.href = 'http://localhost:3000/callback?code=test-code&state=wrong-state';

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await expect(
        act(async () => {
          await result.current.handleCallback();
        })
      ).rejects.toThrow('State mismatch');
    });

    it('throws on authorization error', async () => {
      mockLocation.href =
        'http://localhost:3000/callback?error=access_denied&error_description=User+denied+access';

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await expect(
        act(async () => {
          await result.current.handleCallback();
        })
      ).rejects.toThrow('User denied access');
    });

    it('throws on token exchange failure', async () => {
      mockLocation.href = 'http://localhost:3000/callback?code=test-code&state=test-state';
      
      mockFetch.mockResolvedValue({
        ok: false,
        json: () => Promise.resolve({ error: 'invalid_grant', error_description: 'Invalid code' }),
      });

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await expect(
        act(async () => {
          await result.current.handleCallback();
        })
      ).rejects.toThrow('Invalid code');
    });

    it('clears OAuth state after successful callback', async () => {
      mockLocation.href = 'http://localhost:3000/callback?code=test-code&state=test-state';

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await act(async () => {
        await result.current.handleCallback();
      });

      expect(sessionStorage.getItem('keyles_oauth_state')).toBeNull();
    });
  });

  describe('refreshTokens', () => {
    const initialTokens = {
      access_token: 'old-access-token',
      token_type: 'Bearer' as const,
      expires_in: 3600,
      refresh_token: 'old-refresh-token',
    };

    const newTokens = {
      access_token: 'new-access-token',
      token_type: 'Bearer' as const,
      expires_in: 3600,
      refresh_token: 'new-refresh-token',
    };

    beforeEach(() => {
      localStorage.setItem('keyles_oauth_tokens', JSON.stringify(initialTokens));
    });

    it('refreshes tokens using refresh_token', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(newTokens),
      });

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await act(async () => {
        await result.current.refreshTokens();
      });

      expect(result.current.tokens).toEqual(newTokens);
      expect(mockFetch).toHaveBeenCalledWith(
        'https://sso.example.com/oauth2/token',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });

    it('persists a replacement refresh token while preserving the prior ID token', async () => {
      localStorage.setItem(
        'keyles_oauth_tokens',
        JSON.stringify({ ...initialTokens, id_token: 'existing-id-token' })
      );
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(newTokens),
      });

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await act(async () => {
        await result.current.refreshTokens();
      });

      const expectedTokens = { ...newTokens, id_token: 'existing-id-token' };
      expect(result.current.tokens).toEqual(expectedTokens);
      expect(JSON.parse(localStorage.getItem('keyles_oauth_tokens')!)).toEqual(expectedTokens);
    });

    it('throws when no refresh token available', async () => {
      localStorage.setItem(
        'keyles_oauth_tokens',
        JSON.stringify({ access_token: 'token', token_type: 'Bearer', expires_in: 3600 })
      );

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      await expect(
        act(async () => {
          await result.current.refreshTokens();
        })
      ).rejects.toThrow('No refresh token available');
    });

    it('clears tokens on refresh failure', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        json: () => Promise.resolve({ error: 'invalid_grant' }),
      });

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      // Verify initially authenticated
      expect(result.current.isAuthenticated).toBe(true);

      try {
        await act(async () => {
          await result.current.refreshTokens();
        });
      } catch {
        // Expected to throw
      }

      // After refresh failure, tokens should be cleared from localStorage
      expect(localStorage.getItem('keyles_oauth_tokens')).toBeNull();
    });
  });

  describe('logout', () => {
    it('clears tokens and state', () => {
      localStorage.setItem(
        'keyles_oauth_tokens',
        JSON.stringify({ access_token: 'token', token_type: 'Bearer', expires_in: 3600 })
      );

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));
      expect(result.current.isAuthenticated).toBe(true);

      act(() => {
        result.current.logout();
      });

      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.tokens).toBeNull();
      expect(localStorage.getItem('keyles_oauth_tokens')).toBeNull();
    });
  });

  describe('getAccessToken', () => {
    it('returns access token when valid', () => {
      // Create a valid JWT that expires in the future
      const payload = { exp: Math.floor(Date.now() / 1000) + 3600 };
      const token = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      localStorage.setItem(
        'keyles_oauth_tokens',
        JSON.stringify({ access_token: token, token_type: 'Bearer', expires_in: 3600 })
      );

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.getAccessToken()).toBe(token);
    });

    it('returns null when token is expired', () => {
      // Create an expired JWT
      const payload = { exp: Math.floor(Date.now() / 1000) - 100 };
      const token = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      localStorage.setItem(
        'keyles_oauth_tokens',
        JSON.stringify({ access_token: token, token_type: 'Bearer', expires_in: 3600 })
      );

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.getAccessToken()).toBeNull();
    });

    it('returns null when no tokens', () => {
      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.getAccessToken()).toBeNull();
    });
  });

  describe('getIdToken', () => {
    it('returns ID token when available', () => {
      localStorage.setItem(
        'keyles_oauth_tokens',
        JSON.stringify({
          access_token: 'access',
          token_type: 'Bearer',
          expires_in: 3600,
          id_token: 'id-token-value',
        })
      );

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.getIdToken()).toBe('id-token-value');
    });

    it('returns null when no ID token', () => {
      localStorage.setItem(
        'keyles_oauth_tokens',
        JSON.stringify({ access_token: 'access', token_type: 'Bearer', expires_in: 3600 })
      );

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.getIdToken()).toBeNull();
    });
  });

  describe('getRefreshToken', () => {
    it('returns refresh token when available', () => {
      localStorage.setItem(
        'keyles_oauth_tokens',
        JSON.stringify({
          access_token: 'access',
          token_type: 'Bearer',
          expires_in: 3600,
          refresh_token: 'refresh-token-value',
        })
      );

      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.getRefreshToken()).toBe('refresh-token-value');
    });
  });

  describe('service', () => {
    it('provides access to OAuth service instance', () => {
      const { result } = renderHook(() => useOAuth({ config: mockConfig }));

      expect(result.current.service).toBeDefined();
      expect(typeof result.current.service.authorize).toBe('function');
    });
  });
});
