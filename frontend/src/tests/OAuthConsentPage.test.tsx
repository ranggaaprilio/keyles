import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { OAuthConsentPage } from '../pages/OAuthConsentPage';
import { getConsentDetails, submitConsentDecision } from '../services/oauthInteractionService';

vi.mock('../services/oauthInteractionService', () => ({
  getConsentDetails: vi.fn(),
  submitConsentDecision: vi.fn(),
}));

describe('OAuthConsentPage', () => {
  beforeEach(() => {
    vi.mocked(getConsentDetails).mockReset();
    vi.mocked(submitConsentDecision).mockReset();
  });

  it('loads trusted details and submits approval', async () => {
    vi.mocked(getConsentDetails).mockResolvedValue({
      transaction_id: 'txn-1',
      interaction_csrf_token: 'csrf-1',
      client_id: 'client-1',
      client_name: 'Example App',
      scopes: ['openid', 'email'],
      user_display: 'Alice',
    });
    vi.mocked(submitConsentDecision).mockResolvedValue({ redirect_url: 'https://app.example/callback' });

    render(<MemoryRouter initialEntries={['/oauth2/consent?transaction_id=txn-1']}><OAuthConsentPage /></MemoryRouter>);
    expect(await screen.findByText('Example App')).toBeInTheDocument();
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Verify your identity')).toBeInTheDocument();
    expect(screen.getByText('Access your email address')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Allow' }));

    await waitFor(() => expect(submitConsentDecision).toHaveBeenCalledWith({
      transaction_id: 'txn-1',
      interaction_csrf_token: 'csrf-1',
      approved: true,
    }));
  });

  it('submits denial', async () => {
    vi.mocked(getConsentDetails).mockResolvedValue({
      transaction_id: 'txn-1',
      interaction_csrf_token: 'csrf-1',
      client_id: 'client-1',
      client_name: 'Example App',
      scopes: ['openid'],
      user_display: 'Alice',
    });
    vi.mocked(submitConsentDecision).mockResolvedValue({ redirect_url: 'https://app.example/callback?error=access_denied' });
    render(<MemoryRouter initialEntries={['/oauth2/consent?transaction_id=txn-1']}><OAuthConsentPage /></MemoryRouter>);
    fireEvent.click(await screen.findByRole('button', { name: 'Deny' }));
    await waitFor(() => expect(submitConsentDecision).toHaveBeenCalledWith({
      transaction_id: 'txn-1',
      interaction_csrf_token: 'csrf-1',
      approved: false,
    }));
  });

  it('renders an expired-interaction message', async () => {
    vi.mocked(getConsentDetails).mockRejectedValue({ error: 'invalid_request' });
    render(<MemoryRouter initialEntries={['/oauth2/consent?transaction_id=expired']}><OAuthConsentPage /></MemoryRouter>);
    expect(await screen.findByText('This sign-in request is invalid or has expired.')).toBeInTheDocument();
  });
});
