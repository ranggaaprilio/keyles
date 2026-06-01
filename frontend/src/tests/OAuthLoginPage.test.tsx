import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { OAuthLoginPage } from '../pages/OAuthLoginPage';
import { submitLogin } from '../services/oauthInteractionService';

vi.mock('../services/oauthInteractionService', () => ({ submitLogin: vi.fn() }));

describe('OAuthLoginPage', () => {
  beforeEach(() => {
    vi.mocked(submitLogin).mockReset();
  });

  it('rejects a missing transaction id', () => {
    render(<MemoryRouter><OAuthLoginPage /></MemoryRouter>);
    expect(screen.getByText('Invalid or expired login session.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Sign In' })).toBeDisabled();
  });

  it('submits email and password for the trusted transaction', async () => {
    vi.mocked(submitLogin).mockResolvedValue({ redirect_url: '/oauth2/consent?transaction_id=txn-1' });
    render(<MemoryRouter initialEntries={['/oauth2/login?transaction_id=txn-1']}><OAuthLoginPage /></MemoryRouter>);

    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'alice@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'secret' } });
    fireEvent.click(screen.getByRole('button', { name: 'Sign In' }));

    await waitFor(() => expect(submitLogin).toHaveBeenCalledWith({
      transaction_id: 'txn-1',
      email: 'alice@example.com',
      password: 'secret',
    }));
  });

  it('renders generic and throttled errors', async () => {
    vi.mocked(submitLogin).mockRejectedValueOnce({ error: 'invalid_credentials' });
    render(<MemoryRouter initialEntries={['/oauth2/login?transaction_id=txn-1']}><OAuthLoginPage /></MemoryRouter>);
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'alice@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'bad' } });
    fireEvent.click(screen.getByRole('button', { name: 'Sign In' }));
    expect(await screen.findByText('Invalid email or password.')).toBeInTheDocument();
  });

  it('renders a throttle error', async () => {
    vi.mocked(submitLogin).mockRejectedValueOnce({ error: 'throttled' });
    render(<MemoryRouter initialEntries={['/oauth2/login?transaction_id=txn-1']}><OAuthLoginPage /></MemoryRouter>);
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'alice@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'bad' } });
    fireEvent.click(screen.getByRole('button', { name: 'Sign In' }));
    expect(await screen.findByText('Too many login attempts. Please try again later.')).toBeInTheDocument();
  });

  it('renders a loading state while submitting', async () => {
    let resolveLogin!: (value: { redirect_url: string }) => void;
    vi.mocked(submitLogin).mockReturnValue(new Promise(resolve => {
      resolveLogin = resolve;
    }));
    render(<MemoryRouter initialEntries={['/oauth2/login?transaction_id=txn-1']}><OAuthLoginPage /></MemoryRouter>);
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'alice@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'secret' } });
    fireEvent.click(screen.getByRole('button', { name: 'Sign In' }));
    expect(screen.getByRole('button', { name: 'Signing in...' })).toBeDisabled();
    resolveLogin({ redirect_url: '/oauth2/consent?transaction_id=txn-1' });
    await waitFor(() => expect(screen.getByRole('button', { name: 'Sign In' })).toBeEnabled());
  });
});
