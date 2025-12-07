/**
 * Tests for ResendOTPButton Component
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ResendOTPButton } from '../../../src/components/verification/ResendOTPButton';
import * as tenantApi from '../../../src/services/api/tenant';

// Mock the API
vi.mock('../../../src/services/api/tenant');

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('ResendOTPButton', () => {
  const mockOnSuccess = vi.fn();
  const mockOnError = vi.fn();
  const tenantId = 'test-tenant-id';

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders resend button', () => {
    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('starts with countdown active on mount', () => {
    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText(/resend code in \d+s/i)).toBeInTheDocument();
  });

  it('disables button during countdown', () => {
    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const button = screen.getByRole('button');
    expect(button).toBeDisabled();
  });

  it('enables button after countdown completes', async () => {
    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    // Fast-forward 60 seconds
    vi.advanceTimersByTime(61000);

    await waitFor(() => {
      const button = screen.getByRole('button');
      expect(button).not.toBeDisabled();
    });
  });

  it('calls resendOTP API when button clicked', async () => {
    const mockResendOTP = vi.spyOn(tenantApi, 'resendOTP').mockResolvedValue({
      tenant_id: tenantId,
      message: 'OTP resent successfully',
    });

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    // Wait for countdown to finish
    vi.advanceTimersByTime(61000);

    await waitFor(() => {
      const button = screen.getByRole('button');
      expect(button).not.toBeDisabled();
    });

    const button = screen.getByRole('button');
    fireEvent.click(button);

    await waitFor(() => {
      expect(mockResendOTP).toHaveBeenCalledWith({
        tenant_id: tenantId,
      });
    });
  });

  it('calls onSuccess callback when resend succeeds', async () => {
    vi.spyOn(tenantApi, 'resendOTP').mockResolvedValue({
      tenant_id: tenantId,
      message: 'OTP resent successfully',
    });

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    vi.advanceTimersByTime(61000);

    await waitFor(() => {
      const button = screen.getByRole('button');
      expect(button).not.toBeDisabled();
    });

    const button = screen.getByRole('button');
    fireEvent.click(button);

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith('OTP resent successfully');
    });
  });

  it('calls onError callback when resend fails', async () => {
    const error = { message: 'Failed to resend OTP' };
    vi.spyOn(tenantApi, 'resendOTP').mockRejectedValue(error);

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    vi.advanceTimersByTime(61000);

    await waitFor(() => {
      const button = screen.getByRole('button');
      expect(button).not.toBeDisabled();
    });

    const button = screen.getByRole('button');
    fireEvent.click(button);

    await waitFor(() => {
      expect(mockOnError).toHaveBeenCalled();
    });
  });

  it('shows loading state during resend', async () => {
    vi.spyOn(tenantApi, 'resendOTP').mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    vi.advanceTimersByTime(61000);

    await waitFor(() => {
      const button = screen.getByRole('button');
      expect(button).not.toBeDisabled();
    });

    const button = screen.getByRole('button');
    fireEvent.click(button);

    expect(screen.getByText(/sending/i)).toBeInTheDocument();
  });

  it('restarts countdown after successful resend', async () => {
    vi.spyOn(tenantApi, 'resendOTP').mockResolvedValue({
      tenant_id: tenantId,
      message: 'OTP resent successfully',
    });

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    // Wait for first countdown
    vi.advanceTimersByTime(61000);

    await waitFor(() => {
      const button = screen.getByRole('button');
      expect(button).not.toBeDisabled();
    });

    const button = screen.getByRole('button');
    fireEvent.click(button);

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalled();
    });

    // Check countdown restarted
    expect(screen.getByText(/resend code in \d+s/i)).toBeInTheDocument();
  });
});
