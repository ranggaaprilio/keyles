import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OAuthService } from '@/services/oauthService';

const storageKey = 'keyles_oauth_state';
const config = {
  issuer: 'https://sso.example.com',
  clientId: 'browser-client',
  redirectUri: 'https://app.example.com/callback',
};

describe('OAuth callback compatibility', () => {
  beforeEach(() => {
    sessionStorage.clear();
    vi.restoreAllMocks();
  });

  it('rejects a callback with a mismatched state before token exchange', async () => {
    sessionStorage.setItem(storageKey, JSON.stringify({
      state: 'expected-state',
      code_verifier: 'stored-verifier',
      redirect_uri: config.redirectUri,
      nonce: 'nonce',
      started_at: Date.now(),
    }));
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);

    await expect(new OAuthService(config).completeCallback(
      `${config.redirectUri}?code=approved-code&state=attacker-state`,
    )).rejects.toThrow('State mismatch');
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('exchanges an approved code with the stored PKCE verifier', async () => {
    sessionStorage.setItem(storageKey, JSON.stringify({
      state: 'expected-state',
      code_verifier: 'stored-verifier',
      redirect_uri: config.redirectUri,
      nonce: 'nonce',
      started_at: Date.now(),
    }));
    const tokens = {
      access_token: 'access-token',
      token_type: 'Bearer',
      expires_in: 900,
      refresh_token: 'refresh-token',
      id_token: 'id-token',
    };
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue(tokens),
    });
    vi.stubGlobal('fetch', fetchMock);

    await expect(new OAuthService(config).completeCallback(
      `${config.redirectUri}?code=approved-code&state=expected-state`,
    )).resolves.toEqual(tokens);

    expect(fetchMock).toHaveBeenCalledWith(`${config.issuer}/oauth2/token`, expect.objectContaining({
      method: 'POST',
    }));
    const request = fetchMock.mock.calls[0][1] as RequestInit;
    expect(String(request.body)).toContain('grant_type=authorization_code');
    expect(String(request.body)).toContain('code=approved-code');
    expect(String(request.body)).toContain('client_id=browser-client');
    expect(String(request.body)).toContain('code_verifier=stored-verifier');
    expect(sessionStorage.getItem(storageKey)).toBeNull();
  });
});
