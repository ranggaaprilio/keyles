import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { OAuthSignupPage } from '../pages/OAuthSignupPage';
import {
  getSignupContext,
  resendSignupCode,
  submitSignup,
  verifySignup,
} from '../services/oauthInteractionService';

vi.mock('../services/oauthInteractionService', () => ({
  getSignupContext: vi.fn(),
  submitSignup: vi.fn(),
  resendSignupCode: vi.fn(),
  verifySignup: vi.fn(),
}));

const accepted = { status: 'verification_sent' as const, resend_after_seconds: 60 };

function renderPage(path = '/oauth2/signup?transaction_id=txn-1') {
  return render(<MemoryRouter initialEntries={[path]}><OAuthSignupPage /></MemoryRouter>);
}

async function fillDetails(confirm = 'Str0ng!Pass') {
  await screen.findByText('Create an account to continue to Acme.');
  fireEvent.change(screen.getByLabelText('Full name'), { target: { value: 'Alice Doe' } });
  fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'alice@example.com' } });
  fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'Str0ng!Pass' } });
  fireEvent.change(screen.getByLabelText('Confirm password'), { target: { value: confirm } });
  fireEvent.click(screen.getByRole('button', { name: 'Create account' }));
}

describe('OAuthSignupPage', () => {
  beforeEach(() => {
    vi.mocked(getSignupContext).mockReset().mockResolvedValue({ client_name: 'Acme', signup_enabled: true });
    vi.mocked(submitSignup).mockReset().mockResolvedValue(accepted);
    vi.mocked(resendSignupCode).mockReset().mockResolvedValue(accepted);
    vi.mocked(verifySignup).mockReset();
  });

  it('rejects a missing transaction id', () => {
    renderPage('/oauth2/signup');
    expect(screen.getByText('Invalid or expired sign-up session.')).toBeInTheDocument();
    expect(getSignupContext).not.toHaveBeenCalled();
  });

  it('explains when the application does not allow sign-up', async () => {
    vi.mocked(getSignupContext).mockResolvedValue({ client_name: 'Acme', signup_enabled: false });
    renderPage();
    expect(await screen.findByText('Account creation is not available for this application.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Back to sign in' })).toHaveAttribute('href', '/oauth2/login?transaction_id=txn-1');
  });

  it('blocks mismatched passwords before calling the API', async () => {
    renderPage();
    await fillDetails('Different!1');
    expect(await screen.findByText('Passwords do not match.')).toBeInTheDocument();
    expect(submitSignup).not.toHaveBeenCalled();
  });

  it('shows the backend validation message', async () => {
    vi.mocked(submitSignup).mockRejectedValue({ error: 'invalid_request', error_description: 'password must contain at least one special character' });
    renderPage();
    await fillDetails();
    expect(await screen.findByText('password must contain at least one special character')).toBeInTheDocument();
  });

  it('submits details, then verifies the code and redirects to consent', async () => {
    const redirect = '/oauth2/consent?transaction_id=txn-1';
    vi.mocked(verifySignup).mockResolvedValue({ redirect_url: redirect });
    const originalLocation = window.location;
    Object.defineProperty(window, 'location', { configurable: true, value: { ...originalLocation, href: '' } });

    renderPage();
    await fillDetails();

    await waitFor(() => expect(submitSignup).toHaveBeenCalledWith({
      transaction_id: 'txn-1',
      email: 'alice@example.com',
      password: 'Str0ng!Pass',
      display_name: 'Alice Doe',
    }));
    expect(await screen.findByText('Enter the 6-digit code we sent to alice@example.com.')).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText('Verification code'), { target: { value: '12a3456' } });
    expect(screen.getByLabelText('Verification code')).toHaveValue('123456');
    fireEvent.click(screen.getByRole('button', { name: 'Verify and continue' }));

    await waitFor(() => expect(verifySignup).toHaveBeenCalledWith({ transaction_id: 'txn-1', email: 'alice@example.com', code: '123456' }));
    await waitFor(() => expect(window.location.href).toBe(redirect));
    Object.defineProperty(window, 'location', { configurable: true, value: originalLocation });
  });

  it('shows a generic message for a wrong code', async () => {
    vi.mocked(verifySignup).mockRejectedValue({ error: 'invalid_code', error_description: '' });
    renderPage();
    await fillDetails();
    await screen.findByLabelText('Verification code');

    fireEvent.change(screen.getByLabelText('Verification code'), { target: { value: '000000' } });
    fireEvent.click(screen.getByRole('button', { name: 'Verify and continue' }));

    expect(await screen.findByText(/Invalid or expired verification code\./)).toBeInTheDocument();
  });

  it('enables resend after the cooldown', async () => {
    vi.mocked(submitSignup).mockResolvedValue({ ...accepted, resend_after_seconds: 1 });
    renderPage();
    await fillDetails();

    expect(await screen.findByRole('button', { name: 'Resend code in 1s' })).toBeDisabled();
    fireEvent.click(await screen.findByRole('button', { name: 'Resend code' }, { timeout: 2500 }));
    await waitFor(() => expect(resendSignupCode).toHaveBeenCalledWith({ transaction_id: 'txn-1', email: 'alice@example.com' }));
    expect(await screen.findByText('If a sign-up is pending for this email, a new code has been sent.')).toBeInTheDocument();
  });

  it('returns to the details step to change the email', async () => {
    renderPage();
    await fillDetails();
    fireEvent.click(await screen.findByRole('button', { name: 'Back to details' }));
    expect(screen.getByRole('button', { name: 'Create account' })).toBeInTheDocument();
  });
});
