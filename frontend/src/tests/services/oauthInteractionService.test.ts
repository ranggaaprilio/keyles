import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  getConsentDetails,
  getSignupContext,
  resendSignupCode,
  submitConsentDecision,
  submitLogin,
  submitLogout,
  submitSignup,
  verifySignup,
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

  describe('self sign-up', () => {
    beforeEach(() => {
      document.cookie = 'keyles_csrf=csrf-cookie-token; path=/';
    });

    it('loads the sign-up context with an encoded transaction id', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ client_name: 'Acme', signup_enabled: true }), { status: 200 }));

      const ctx = await getSignupContext('txn 1');

      expect(ctx.signup_enabled).toBe(true);
      expect(fetch).toHaveBeenCalledWith('http://localhost:8080/oauth2/signup/context?transaction_id=txn%201', { credentials: 'include' });
    });

    it('sends sign-up, resend and verify requests with cookies and the CSRF header', async () => {
      const accepted = { status: 'verification_sent', resend_after_seconds: 60 };
      vi.spyOn(globalThis, 'fetch')
        .mockResolvedValueOnce(new Response(JSON.stringify(accepted), { status: 202 }))
        .mockResolvedValueOnce(new Response(JSON.stringify(accepted), { status: 202 }))
        .mockResolvedValueOnce(new Response(JSON.stringify({ redirect_url: '/oauth2/consent?transaction_id=txn-1' }), { status: 200 }));

      await submitSignup({ transaction_id: 'txn-1', email: 'a@example.com', password: 'Str0ng!Pass', display_name: 'Alice' });
      await resendSignupCode({ transaction_id: 'txn-1', email: 'a@example.com' });
      const result = await verifySignup({ transaction_id: 'txn-1', email: 'a@example.com', code: '123456' });

      expect(result.redirect_url).toBe('/oauth2/consent?transaction_id=txn-1');
      const paths = ['/oauth2/signup', '/oauth2/signup/resend', '/oauth2/signup/verify'];
      paths.forEach((path, i) => {
        expect(fetch).toHaveBeenNthCalledWith(i + 1, `http://localhost:8080${path}`, expect.objectContaining({
          method: 'POST',
          credentials: 'include',
          headers: expect.objectContaining({ 'X-CSRF-Token': 'csrf-cookie-token' }),
        }));
      });
    });

    it('surfaces backend error codes', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ error: 'invalid_code', error_description: 'bad' }), { status: 401 }));

      await expect(verifySignup({ transaction_id: 'txn-1', email: 'a@example.com', code: '000000' }))
        .rejects.toEqual({ error: 'invalid_code', error_description: 'bad' });
    });
  });
});
