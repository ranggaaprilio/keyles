import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { OAuthLogoutPage } from '../pages/OAuthLogoutPage';
import { submitLogout } from '../services/oauthInteractionService';

vi.mock('../services/oauthInteractionService', () => ({ submitLogout: vi.fn() }));

describe('OAuthLogoutPage', () => {
  beforeEach(() => vi.mocked(submitLogout).mockReset());

  it('submits logout once and renders completion', async () => {
    vi.mocked(submitLogout).mockResolvedValue();
    render(<OAuthLogoutPage />);
    expect(await screen.findByText('Signed Out')).toBeInTheDocument();
    expect(submitLogout).toHaveBeenCalledTimes(1);
  });
});
