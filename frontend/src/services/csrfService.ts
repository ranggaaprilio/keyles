/**
 * CSRF Service
 * Extracts CSRF token from cookie and provides it for API requests
 */

export const CSRF_COOKIE_NAME = 'keyles_csrf';
export const CSRF_HEADER_NAME = 'X-CSRF-Token';

/**
 * Get the CSRF token from the browser cookies
 */
export function getCsrfToken(): string | null {
  const cookies = document.cookie.split(';');
  for (const cookie of cookies) {
    const [name, value] = cookie.trim().split('=');
    if (name === CSRF_COOKIE_NAME) {
      return decodeURIComponent(value);
    }
  }
  return null;
}

/**
 * Check if CSRF token exists in cookies
 */
export function hasCsrfToken(): boolean {
  return getCsrfToken() !== null;
}
