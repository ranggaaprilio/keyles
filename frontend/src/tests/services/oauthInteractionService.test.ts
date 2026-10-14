import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  getConsentDetails,
  submitConsentDecision,
  submitLogin,
  submitLogout,
} from '../../services/oauthInteractionService';

describe('oauthInteractionService', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('submits login credentials with cookies', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ redirect_url: '/oauth2/consent' }), { status: 200 }));

    await submitLogin({ transaction_id: 'txn-1', email: 'alice@example.com', password: 'secret' });

    expect(fetch).toHaveBeenCalledWith('http://localhost:8080/oauth2/login', expect.objectContaining({
      method: 'POST',
      credentials: 'include',
    }));
  });

  it('loads and submits consent with cookies', async () => {
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify({
        transaction_id: 'txn-1',
        interaction_csrf_token: 'csrf-1',
        client_id: 'client-1',
        client_name: 'Example',
        scopes: ['openid'],
        user_display: 'Alice',
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ redirect_url: 'https://app.example/callback' }), { status: 200 }));

    await getConsentDetails('txn-1');
    await submitConsentDecision({ transaction_id: 'txn-1', interaction_csrf_token: 'csrf-1', approved: true });

    expect(fetch).toHaveBeenNthCalledWith(1, 'http://localhost:8080/oauth2/consent/txn-1', { credentials: 'include' });
    expect(fetch).toHaveBeenNthCalledWith(2, 'http://localhost:8080/oauth2/consent', expect.objectContaining({
      method: 'POST',
      credentials: 'include',
    }));
  });

  it('submits provider-local logout with cookies', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(null, { status: 204 }));

    await submitLogout();

    expect(fetch).toHaveBeenCalledWith('http://localhost:8080/oauth2/logout', expect.objectContaining({
      method: 'POST',
      credentials: 'include',
    }));
  });

  it('sends the CSRF header on login, consent and logout', async () => {
    document.cookie = 'keyles_csrf=csrf-cookie-token; path=/';
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify({ redirect_url: '/oauth2/consent' }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ redirect_url: 'https://app.example/callback' }), { status: 200 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    await submitLogin({ transaction_id: 'txn-1', email: 'alice@example.com', password: 'secret' });
    await submitConsentDecision({ transaction_id: 'txn-1', interaction_csrf_token: 'csrf-1', approved: true });
    await submitLogout();

    ['/oauth2/login', '/oauth2/consent', '/oauth2/logout'].forEach((path, i) => {
      expect(fetch).toHaveBeenNthCalledWith(i + 1, `http://localhost:8080${path}`, expect.objectContaining({
        headers: expect.objectContaining({ 'X-CSRF-Token': 'csrf-cookie-token' }),
      }));
    });
  });
});
