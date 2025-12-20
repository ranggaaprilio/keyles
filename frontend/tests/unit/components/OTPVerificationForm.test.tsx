/**
 * Tests for OTPVerificationForm Component
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { OTPVerificationForm } from '../../../src/components/verification/OTPVerificationForm';
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

describe('OTPVerificationForm', () => {
  const mockOnSuccess = vi.fn();
  const mockOnError = vi.fn();
  const tenantId = 'test-tenant-id';

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders 6 input fields', () => {
    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox');
    expect(inputs).toHaveLength(6);
  });

  it('allows entering single digits only', () => {
    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    fireEvent.change(inputs[0]!, { target: { value: '1' } });
    expect(inputs[0]!.value).toBe('1');

    fireEvent.change(inputs[0]!, { target: { value: '12' } });
    expect(inputs[0]!.value).toBe('2'); // Only last digit
  });

  it('auto-focuses next input after entering digit', async () => {
    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    fireEvent.change(inputs[0]!, { target: { value: '1' } });
    
    await waitFor(() => {
      expect(inputs[1]).toHaveFocus();
    });
  });

  it('handles backspace to move to previous input', () => {
    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    // Enter digit in first input
    fireEvent.change(inputs[0]!, { target: { value: '1' } });
    
    // Focus second input and press backspace (empty)
    fireEvent.keyDown(inputs[1]!, { key: 'Backspace' });
    
    expect(inputs[0]).toHaveFocus();
  });

  it('handles paste of 6-digit code', () => {
    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    const clipboardData = {
      getData: vi.fn(() => '123456'),
    };

    fireEvent.paste(inputs[0]!, { clipboardData } as any);

    expect(inputs[0]!.value).toBe('1');
    expect(inputs[1]!.value).toBe('2');
    expect(inputs[2]!.value).toBe('3');
    expect(inputs[3]!.value).toBe('4');
    expect(inputs[4]!.value).toBe('5');
    expect(inputs[5]!.value).toBe('6');
  });

  it('disables submit button when OTP is incomplete', () => {
    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const submitButton = screen.getByRole('button', { name: /verify email/i });
    expect(submitButton).toBeDisabled();
  });

  it('enables submit button when OTP is complete', () => {
    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    // Enter complete OTP
    '123456'.split('').forEach((digit, index) => {
      fireEvent.change(inputs[index]!, { target: { value: digit } });
    });

    const submitButton = screen.getByRole('button', { name: /verify email/i });
    expect(submitButton).not.toBeDisabled();
  });

  it('calls verifyOTP API on submit with correct data', async () => {
    const mockVerifyOTP = vi.spyOn(tenantApi, 'verifyOTP').mockResolvedValue({
      tenant_id: tenantId,
      status: 'active',
      message: 'Verified successfully',
    });

    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    // Enter complete OTP
    '123456'.split('').forEach((digit, index) => {
      fireEvent.change(inputs[index]!, { target: { value: digit } });
    });

    const submitButton = screen.getByRole('button', { name: /verify email/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockVerifyOTP).toHaveBeenCalledWith({
        tenant_id: tenantId,
        otp_code: '123456',
      });
    });
  });

  it('calls onSuccess callback when verification succeeds', async () => {
    vi.spyOn(tenantApi, 'verifyOTP').mockResolvedValue({
      tenant_id: tenantId,
      status: 'active',
      message: 'Verified successfully',
    });

    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    '123456'.split('').forEach((digit, index) => {
      fireEvent.change(inputs[index]!, { target: { value: digit } });
    });

    const submitButton = screen.getByRole('button', { name: /verify email/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalled();
    });
  });

  it('calls onError callback when verification fails', async () => {
    const error = new Error('Invalid OTP');
    vi.spyOn(tenantApi, 'verifyOTP').mockRejectedValue(error);

    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    '123456'.split('').forEach((digit, index) => {
      fireEvent.change(inputs[index]!, { target: { value: digit } });
    });

    const submitButton = screen.getByRole('button', { name: /verify email/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockOnError).toHaveBeenCalled();
    });
  });

  it('shows loading state during verification', async () => {
    vi.spyOn(tenantApi, 'verifyOTP').mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    '123456'.split('').forEach((digit, index) => {
      fireEvent.change(inputs[index]!, { target: { value: digit } });
    });

    const submitButton = screen.getByRole('button', { name: /verify email/i });
    fireEvent.click(submitButton);

    expect(screen.getByText(/verifying/i)).toBeInTheDocument();
  });

  it('rejects non-numeric input', () => {
    render(
      <OTPVerificationForm
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    
    fireEvent.change(inputs[0]!, { target: { value: 'a' } });
    expect(inputs[0]!.value).toBe('');
  });
});
