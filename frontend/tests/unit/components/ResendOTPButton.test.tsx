/**
 * Tests for ResendOTPButton Component
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ResendOTPButton } from "../../../src/components/verification/ResendOTPButton";

// Mock functions
const mockVerifyOTP = vi.fn();
const mockResendOTP = vi.fn();

// Mock the API
vi.mock("../../../src/services/api/tenant", () => ({
  verifyOTP: (...args: any[]) => mockVerifyOTP(...args),
  resendOTP: (...args: any[]) => mockResendOTP(...args),
}));

// Mock the OTP store to avoid real timers
vi.mock("../../../src/stores/otpStore", () => ({
  useOTPStore: vi.fn(() => ({
    countdown: 0,
    isActive: false,
    startCountdown: vi.fn(),
    resetCountdown: vi.fn(),
    tick: vi.fn(),
  })),
}));

import { useOTPStore } from "../../../src/stores/otpStore";

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

describe("ResendOTPButton", () => {
  const mockOnSuccess = vi.fn();
  const mockOnError = vi.fn();
  const tenantId = "test-tenant-id";
  const mockStartCountdown = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    // Default: countdown finished (button enabled)
    (useOTPStore as any).mockReturnValue({
      countdown: 0,
      isActive: false,
      startCountdown: mockStartCountdown,
      resetCountdown: vi.fn(),
      tick: vi.fn(),
    });
  });

  it("renders resend button", () => {
    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByRole("button")).toBeInTheDocument();
  });

  it("starts with countdown active on mount", () => {
    (useOTPStore as any).mockReturnValue({
      countdown: 30,
      isActive: true,
      startCountdown: mockStartCountdown,
      resetCountdown: vi.fn(),
      tick: vi.fn(),
    });

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText(/resend code in 30s/i)).toBeInTheDocument();
  });

  it("disables button during countdown", () => {
    (useOTPStore as any).mockReturnValue({
      countdown: 30,
      isActive: true,
      startCountdown: mockStartCountdown,
      resetCountdown: vi.fn(),
      tick: vi.fn(),
    });

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const button = screen.getByRole("button");
    expect(button).toBeDisabled();
  });

  it("enables button after countdown completes", async () => {
    // countdown = 0, isActive = false
    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const button = screen.getByRole("button");
    expect(button).not.toBeDisabled();
  });

  it("calls resendOTP API when button clicked", async () => {
    mockResendOTP.mockResolvedValue({
      tenant_id: tenantId,
      message: "OTP resent successfully",
    });

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const button = screen.getByRole("button");
    fireEvent.click(button);

    await waitFor(() => {
      expect(mockResendOTP).toHaveBeenCalled();
    });

    // Check the first argument is the request data
    expect(mockResendOTP.mock.calls[0][0]).toEqual({
      tenant_id: tenantId,
    });
  });

  it("calls onSuccess callback when resend succeeds", async () => {
    mockResendOTP.mockResolvedValue({
      tenant_id: tenantId,
      message: "OTP resent successfully",
    });

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const button = screen.getByRole("button");
    fireEvent.click(button);

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith("OTP resent successfully");
    });
  });

  it("calls onError callback when resend fails", async () => {
    const error = { message: "Failed to resend OTP" };
    mockResendOTP.mockRejectedValue(error);

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const button = screen.getByRole("button");
    fireEvent.click(button);

    await waitFor(() => {
      expect(mockOnError).toHaveBeenCalled();
    });
  });

  it("shows loading state during resend", async () => {
    let resolvePromise: (val: any) => void;
    mockResendOTP.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolvePromise = resolve;
        })
    );

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const button = screen.getByRole("button");
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText(/sending/i)).toBeInTheDocument();
    });

    resolvePromise!({ tenant_id: tenantId, message: "Success" });
  });

  it("restarts countdown after successful resend", async () => {
    mockResendOTP.mockResolvedValue({
      tenant_id: tenantId,
      message: "OTP resent successfully",
    });

    render(
      <ResendOTPButton
        tenantId={tenantId}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />,
      { wrapper: createWrapper() }
    );

    const button = screen.getByRole("button");
    fireEvent.click(button);

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalled();
    });

    // After success, startCountdown should be called with 60
    expect(mockStartCountdown).toHaveBeenCalledWith(60);
  });
});
