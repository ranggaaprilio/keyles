import { afterEach, describe, expect, it, vi } from 'vitest';

describe('VITE_API_URL is treated as an origin', () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.resetModules();
  });

  it('tenant service appends /api/v1 to VITE_API_URL', async () => {
    vi.stubEnv('VITE_API_URL', 'http://example.test');
    vi.resetModules();
    const { apiClient } = await import('../../services/api/tenant');
    expect(apiClient.defaults.baseURL).toBe('http://example.test/api/v1');
  });

  it('auth service appends /api/v1 to VITE_API_URL', async () => {
    vi.stubEnv('VITE_API_URL', 'http://example.test');
    vi.resetModules();
    const { apiClient } = await import('../../services/api/auth');
    expect(apiClient.defaults.baseURL).toBe('http://example.test/api/v1');
  });
});
