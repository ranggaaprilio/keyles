// Credentialed OAuth login API client for browser-flow interactions
import type {
  OAuthConsentDecisionRequest,
  OAuthConsentDetails,
  OAuthInteractionError,
  OAuthInteractionRedirect,
  OAuthLoginRequest,
  OAuthLoginResponse,
} from '../types/oauth';
import { CSRF_HEADER_NAME, getCsrfToken } from './csrfService';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

/**
 * Send login credentials to the backend.
 * Includes credentials (cookies) for CORS origins in development.
 */
export async function submitLogin(req: OAuthLoginRequest): Promise<OAuthLoginResponse> {
  const response = await postWithCsrf('/oauth2/login', req);

  if (!response.ok) {
    const errorBody: OAuthInteractionError = await response.json().catch(() => ({
      error: 'server_error',
      error_description: 'An unexpected error occurred.',
    }));
    throw errorBody;
  }

  return response.json();
}

async function parseResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorBody: OAuthInteractionError = await response.json().catch(() => ({
      error: 'server_error',
      error_description: 'An unexpected error occurred.',
    }));
    throw errorBody;
  }
  return response.json() as Promise<T>;
}

/** Load trusted consent details for a pending browser interaction. */
export async function getConsentDetails(transactionId: string): Promise<OAuthConsentDetails> {
  const response = await fetch(`${API_BASE_URL}/oauth2/consent/${encodeURIComponent(transactionId)}`, {
    credentials: 'include',
  });
  return parseResponse<OAuthConsentDetails>(response);
}

/** Submit an allow or deny decision for a pending browser interaction. */
export async function submitConsentDecision(req: OAuthConsentDecisionRequest): Promise<OAuthInteractionRedirect> {
  const response = await postWithCsrf('/oauth2/consent', req);
  return parseResponse<OAuthInteractionRedirect>(response);
}

/** End the provider-local browser session. */
export async function submitLogout(): Promise<void> {
  const response = await postWithCsrf('/oauth2/logout');
  if (!response.ok) {
    throw new Error('Logout failed');
  }
}

/**
 * POST with credentials and the double-submit CSRF header required by the
 * backend's global CSRF middleware. Sends a JSON body when one is given.
 */
function postWithCsrf(path: string, body?: unknown): Promise<Response> {
  const headers: Record<string, string> = {};
  const csrfToken = getCsrfToken();
  if (csrfToken) {
    headers[CSRF_HEADER_NAME] = csrfToken;
  }
  const init: RequestInit = { method: 'POST', headers, credentials: 'include' };
  if (body !== undefined) {
    headers['Content-Type'] = 'application/json';
    init.body = JSON.stringify(body);
  }
  return fetch(`${API_BASE_URL}${path}`, init);
}

/** Parse interaction error from URL search params */
export function parseInteractionError(searchParams: URLSearchParams): OAuthInteractionError | null {
  const error = searchParams.get('error');
  if (!error) return null;
  return {
    error,
    error_description: searchParams.get('error_description') || '',
  };
}
