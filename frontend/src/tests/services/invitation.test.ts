import axios from 'axios';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { acceptInvitation } from '../../services/api/invitation';

describe('invitation service', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('sends the CSRF header when accepting an invitation', async () => {
    document.cookie = 'keyles_csrf=csrf-cookie-token; path=/';
    const post = vi.spyOn(axios, 'post').mockResolvedValue({ data: {} });

    await acceptInvitation('tok/en=', { password: 'Str0ng!Pass' });

    expect(post).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/invitations/tok%2Fen%3D/accept',
      { password: 'Str0ng!Pass' },
      { headers: { 'X-CSRF-Token': 'csrf-cookie-token' } },
    );
  });
});
