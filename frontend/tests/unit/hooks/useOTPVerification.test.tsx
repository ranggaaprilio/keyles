/**
 * Tests for useOTPVerification Hook
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useOTPVerification } from "../../../src/hooks/useOTPVerification";
import { ApiException } from "../../../src/types/api";

// Mock functions
const mockVerifyOTP = vi.fn();
const mockResendOTP = vi.fn();

// Mock the API
vi.mock("../../../src/services/api/tenant", () => ({
  verifyOTP: (...args: any[]) => mockVerifyOTP(...args),
  resendOTP: (...args: any[]) => mockResendOTP(...args),
}));

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

describe("useOTPVerification", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("verifyMutation", () => {
    it("calls verifyOTP API with correct data", async () => {
      mockVerifyOTP.mockResolvedValue({
        tenant_id: "test-id",
        status: "active",
        message: "Success",
      });

      const { result } = renderHook(() => useOTPVerification(), {
        wrapper: createWrapper(),
      });

      result.current.verifyMutation.mutate({
        tenant_id: "test-id",
        otp_code: "123456",
      });

      await waitFor(() => {
        expect(mockVerifyOTP).toHaveBeenCalled();
      });

      // Check the first argument is the request data
      expect(mockVerifyOTP.mock.calls[0][0]).toEqual({
        tenant_id: "test-id",
        otp_code: "123456",
      });
    });

    it("calls onVerifySuccess callback on success", async () => {
      const responseData = {
        tenant_id: "test-id",
        status: "active",
        message: "Success",
      };
      mockVerifyOTP.mockResolvedValue(responseData);

      const onVerifySuccess = vi.fn();

      const { result } = renderHook(
        () => useOTPVerification({ onVerifySuccess }),
        {
          wrapper: createWrapper(),
        }
      );

      result.current.verifyMutation.mutate({
        tenant_id: "test-id",
        otp_code: "123456",
      });

      await waitFor(() => {
        expect(onVerifySuccess).toHaveBeenCalled();
      });

      // React Query calls onSuccess with (data, variables, context)
      expect(onVerifySuccess.mock.calls[0][0]).toEqual(responseData);
    });

    it("calls onVerifyError callback on error", async () => {
      const error = new ApiException(401, "Invalid OTP");
      mockVerifyOTP.mockRejectedValue(error);

      const onVerifyError = vi.fn();

      const { result } = renderHook(
        () => useOTPVerification({ onVerifyError }),
        {
          wrapper: createWrapper(),
        }
      );

      result.current.verifyMutation.mutate({
        tenant_id: "test-id",
        otp_code: "123456",
      });

      await waitFor(() => {
        expect(onVerifyError).toHaveBeenCalled();
      });

      // React Query calls onError with (error, variables, context)
      expect(onVerifyError.mock.calls[0][0]).toBe(error);
    });

    it("sets isVerifying to true during mutation", async () => {
      let resolvePromise: () => void;
      mockVerifyOTP.mockImplementation(
        () =>
          new Promise((resolve) => {
            resolvePromise = () =>
              resolve({
                tenant_id: "test-id",
                status: "active",
                message: "Success",
              });
          })
      );

      const { result } = renderHook(() => useOTPVerification(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isVerifying).toBe(false);

      result.current.verifyMutation.mutate({
        tenant_id: "test-id",
        otp_code: "123456",
      });

      await waitFor(() => {
        expect(result.current.isVerifying).toBe(true);
      });

      resolvePromise!();
    });
  });

  describe("resendMutation", () => {
    it("calls resendOTP API with correct data", async () => {
      mockResendOTP.mockResolvedValue({
        tenant_id: "test-id",
        message: "OTP resent successfully",
      });

      const { result } = renderHook(() => useOTPVerification(), {
        wrapper: createWrapper(),
      });

      result.current.resendMutation.mutate({
        tenant_id: "test-id",
      });

      await waitFor(() => {
        expect(mockResendOTP).toHaveBeenCalled();
      });

      // Check the first argument is the request data
      expect(mockResendOTP.mock.calls[0][0]).toEqual({
        tenant_id: "test-id",
      });
    });

    it("calls onResendSuccess callback on success", async () => {
      const responseData = {
        tenant_id: "test-id",
        message: "OTP resent successfully",
      };
      mockResendOTP.mockResolvedValue(responseData);

      const onResendSuccess = vi.fn();

      const { result } = renderHook(
        () => useOTPVerification({ onResendSuccess }),
        {
          wrapper: createWrapper(),
        }
      );

      result.current.resendMutation.mutate({
        tenant_id: "test-id",
      });

      await waitFor(() => {
        expect(onResendSuccess).toHaveBeenCalled();
      });

      // React Query calls onSuccess with (data, variables, context)
      expect(onResendSuccess.mock.calls[0][0]).toEqual(responseData);
    });

    it("calls onResendError callback on error", async () => {
      const error = new ApiException(429, "Rate limit exceeded");
      mockResendOTP.mockRejectedValue(error);

      const onResendError = vi.fn();

      const { result } = renderHook(
        () => useOTPVerification({ onResendError }),
        {
          wrapper: createWrapper(),
        }
      );

      result.current.resendMutation.mutate({
        tenant_id: "test-id",
      });

      await waitFor(() => {
        expect(onResendError).toHaveBeenCalled();
      });

      // React Query calls onError with (error, variables, context)
      expect(onResendError.mock.calls[0][0]).toBe(error);
    });

    it("sets isResending to true during mutation", async () => {
      let resolvePromise: () => void;
      mockResendOTP.mockImplementation(
        () =>
          new Promise((resolve) => {
            resolvePromise = () =>
              resolve({ tenant_id: "test-id", message: "OTP resent" });
          })
      );

      const { result } = renderHook(() => useOTPVerification(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isResending).toBe(false);

      result.current.resendMutation.mutate({
        tenant_id: "test-id",
      });

      await waitFor(() => {
        expect(result.current.isResending).toBe(true);
      });

      resolvePromise!();
    });
  });

  it("works without callbacks", async () => {
    mockVerifyOTP.mockResolvedValue({
      tenant_id: "test-id",
      status: "active",
      message: "Success",
    });

    const { result } = renderHook(() => useOTPVerification(), {
      wrapper: createWrapper(),
    });

    result.current.verifyMutation.mutate({
      tenant_id: "test-id",
      otp_code: "123456",
    });

    await waitFor(() => {
      expect(result.current.verifyMutation.isSuccess).toBe(true);
    });
  });
});
