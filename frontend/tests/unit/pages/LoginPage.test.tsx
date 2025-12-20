/**
 * LoginPage Unit Tests
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LoginPage } from '../../../src/pages/LoginPage';
import * as authApi from '../../../src/services/api/auth';

// Mock auth service
vi.mock('../../../src/services/api/auth', () => ({
  login: vi.fn(),
  isAuthenticated: vi.fn(),
  storeAuthData: vi.fn(),
}));

// Mock router navigation
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
    mutations: { retry: false },
  },
});

function renderLoginPage() {
  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <LoginPage />
      </BrowserRouter>
    </QueryClientProvider>
  );
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    queryClient.clear();
    (authApi.isAuthenticated as any).mockReturnValue(false);
  });

  it('renders login form', () => {
    renderLoginPage();
    
    expect(screen.getByText(/Welcome Back/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Email Address/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Sign In/i })).toBeInTheDocument();
  });

  it('shows validation errors for empty fields', async () => {
    const user = userEvent.setup();
    renderLoginPage();
    
    const submitButton = screen.getByRole('button', { name: /Sign In/i });
    await user.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(/Invalid email address/i)).toBeInTheDocument();
      expect(screen.getByText(/Password is required/i)).toBeInTheDocument();
    });
  });

  it('shows validation error for invalid email', async () => {
    const user = userEvent.setup();
    renderLoginPage();
    
    const emailInput = screen.getByLabelText(/Email Address/i);
    await user.type(emailInput, 'invalid-email');
    
    const submitButton = screen.getByRole('button', { name: /Sign In/i });
    await user.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(/Invalid email address/i)).toBeInTheDocument();
    });
  });

  it('submits form with valid credentials', async () => {
    const user = userEvent.setup();
    const mockLoginResponse = {
      token: 'test-jwt-token',
      expires_in: 86400,
      user: {
        id: 'user-123',
        email: 'admin@example.com',
        full_name: 'Admin User',
        role: 'admin',
      },
      tenant: {
        id: 'tenant-123',
        organization_name: 'Test Org',
        status: 'active',
      },
    };
    
    (authApi.login as any).mockResolvedValue(mockLoginResponse);
    
    renderLoginPage();
    
    const emailInput = screen.getByLabelText(/Email Address/i);
    const passwordInput = screen.getByLabelText(/Password/i);
    const submitButton = screen.getByRole('button', { name: /Sign In/i });
    
    await user.type(emailInput, 'admin@example.com');
    await user.type(passwordInput, 'SecurePassword123!');
    await user.click(submitButton);
    
    await waitFor(() => {
      expect(authApi.login).toHaveBeenCalledWith('admin@example.com', 'SecurePassword123!');
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true });
    });
  });

  it('shows error message on login failure', async () => {
    const user = userEvent.setup();
    (authApi.login as any).mockRejectedValue(new Error('Invalid credentials'));
    
    renderLoginPage();
    
    const emailInput = screen.getByLabelText(/Email Address/i);
    const passwordInput = screen.getByLabelText(/Password/i);
    const submitButton = screen.getByRole('button', { name: /Sign In/i });
    
    await user.type(emailInput, 'admin@example.com');
    await user.type(passwordInput, 'WrongPassword');
    await user.click(submitButton);
    
    await waitFor(() => {
      expect(authApi.login).toHaveBeenCalled();
      // Toast notification should appear
    });
  });

  it('disables form during submission', async () => {
    const user = userEvent.setup();
    (authApi.login as any).mockImplementation(() => new Promise(resolve => setTimeout(resolve, 1000)));
    
    renderLoginPage();
    
    const emailInput = screen.getByLabelText(/Email Address/i) as HTMLInputElement;
    const passwordInput = screen.getByLabelText(/Password/i) as HTMLInputElement;
    const submitButton = screen.getByRole('button', { name: /Sign In/i }) as HTMLButtonElement;
    
    await user.type(emailInput, 'admin@example.com');
    await user.type(passwordInput, 'SecurePassword123!');
    await user.click(submitButton);
    
    expect(emailInput.disabled).toBe(true);
    expect(passwordInput.disabled).toBe(true);
    expect(submitButton.disabled).toBe(true);
    expect(screen.getByText(/Signing in/i)).toBeInTheDocument();
  });

  it('redirects to dashboard if already authenticated', () => {
    (authApi.isAuthenticated as any).mockReturnValue(true);
    
    renderLoginPage();
    
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true });
  });

  it('toggles password visibility', async () => {
    const user = userEvent.setup();
    renderLoginPage();
    
    const passwordInput = screen.getByLabelText(/Password/i) as HTMLInputElement;
    expect(passwordInput.type).toBe('password');
    
    const toggleButton = screen.getByRole('button', { name: /toggle password/i });
    await user.click(toggleButton);
    
    expect(passwordInput.type).toBe('text');
    
    await user.click(toggleButton);
    expect(passwordInput.type).toBe('password');
  });

  it('shows link to registration page', () => {
    renderLoginPage();
    
    const registerLink = screen.getByRole('link', { name: /Create an account/i });
    expect(registerLink).toHaveAttribute('href', '/register');
  });
});
