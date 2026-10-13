import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { LoginPage } from '../pages/LoginPage';

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    login: vi.fn(),
    isLoggingIn: false,
    isAuthenticated: false,
    loginMutation: { isSuccess: false, isError: false, error: null },
  }),
}));

describe('LoginPage', () => {
  it('shows the redirect message toast exactly once', async () => {
    const message = 'Account created successfully. Please log in.';
    render(
      <MemoryRouter initialEntries={[{ pathname: '/login', state: { message } }]}>
        <LoginPage />
      </MemoryRouter>,
    );

    await screen.findByText(message);
    // Let any follow-up renders settle, then count toasts.
    await new Promise((r) => setTimeout(r, 50));
    expect(screen.getAllByText(message)).toHaveLength(1);
  });
});
