/**
 * Tests for useOTPVerification Hook
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useOTPVerification } from '../../../src/hooks/useOTPVerification';
import * as tenantApi from '../../../src/services/api/tenant';
import { ApiException } from '../../../src/types/api';

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

describe('useOTPVerification', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('verifyMutation', () => {
    it('calls verifyOTP API with correct data', async () => {
      const mockVerifyOTP = vi.spyOn(tenantApi, 'verifyOTP').mockResolvedValue({
        tenant_id: 'test-id',
        status: 'active',
        message: 'Success',
      });

      const { result } = renderHook(() => useOTPVerification(), {
        wrapper: createWrapper(),
      });

      result.current.verifyMutation.mutate({
        tenant_id: 'test-id',
        otp_code: '123456',
      });

      await waitFor(() => {
        expect(mockVerifyOTP).toHaveBeenCalledWith({
          tenant_id: 'test-id',
          otp_code: '123456',
        });
      });
    });

    it('calls onVerifySuccess callback on success', async () => {
      vi.spyOn(tenantApi, 'verifyOTP').mockResolvedValue({
        tenant_id: 'test-id',
        status: 'active',
        message: 'Success',
      });

      const onVerifySuccess = vi.fn();

      const { result } = renderHook(() => useOTPVerification({ onVerifySuccess }), {
        wrapper: createWrapper(),
      });

      result.current.verifyMutation.mutate({
        tenant_id: 'test-id',
        otp_code: '123456',
      });

      await waitFor(() => {
        expect(onVerifySuccess).toHaveBeenCalledWith({
          tenant_id: 'test-id',
          status: 'active',
          message: 'Success',
        });
      });
    });

    it('calls onVerifyError callback on error', async () => {
      const error = new ApiException(401, 'Invalid OTP');
      vi.spyOn(tenantApi, 'verifyOTP').mockRejectedValue(error);

      const onVerifyError = vi.fn();

      const { result } = renderHook(() => useOTPVerification({ onVerifyError }), {
        wrapper: createWrapper(),
      });

      result.current.verifyMutation.mutate({
        tenant_id: 'test-id',
        otp_code: '123456',
      });

      await waitFor(() => {
        expect(onVerifyError).toHaveBeenCalledWith(error);
      });
    });

    it('sets isVerifying to true during mutation', async () => {
      vi.spyOn(tenantApi, 'verifyOTP').mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );

      const { result } = renderHook(() => useOTPVerification(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isVerifying).toBe(false);

      result.current.verifyMutation.mutate({
        tenant_id: 'test-id',
        otp_code: '123456',
      });

      expect(result.current.isVerifying).toBe(true);
    });
  });

  describe('resendMutation', () => {
    it('calls resendOTP API with correct data', async () => {
      const mockResendOTP = vi.spyOn(tenantApi, 'resendOTP').mockResolvedValue({
        tenant_id: 'test-id',
        message: 'OTP resent successfully',
      });

      const { result } = renderHook(() => useOTPVerification(), {
        wrapper: createWrapper(),
      });

      result.current.resendMutation.mutate({
        tenant_id: 'test-id',
      });

      await waitFor(() => {
        expect(mockResendOTP).toHaveBeenCalledWith({
          tenant_id: 'test-id',
        });
      });
    });

    it('calls onResendSuccess callback on success', async () => {
      vi.spyOn(tenantApi, 'resendOTP').mockResolvedValue({
        tenant_id: 'test-id',
        message: 'OTP resent successfully',
      });

      const onResendSuccess = vi.fn();

      const { result } = renderHook(() => useOTPVerification({ onResendSuccess }), {
        wrapper: createWrapper(),
      });

      result.current.resendMutation.mutate({
        tenant_id: 'test-id',
      });

      await waitFor(() => {
        expect(onResendSuccess).toHaveBeenCalledWith({
          tenant_id: 'test-id',
          message: 'OTP resent successfully',
        });
      });
    });

    it('calls onResendError callback on error', async () => {
      const error = new ApiException(429, 'Rate limit exceeded');
      vi.spyOn(tenantApi, 'resendOTP').mockRejectedValue(error);

      const onResendError = vi.fn();

      const { result } = renderHook(() => useOTPVerification({ onResendError }), {
        wrapper: createWrapper(),
      });

      result.current.resendMutation.mutate({
        tenant_id: 'test-id',
      });

      await waitFor(() => {
        expect(onResendError).toHaveBeenCalledWith(error);
      });
    });

    it('sets isResending to true during mutation', async () => {
      vi.spyOn(tenantApi, 'resendOTP').mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );

      const { result } = renderHook(() => useOTPVerification(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isResending).toBe(false);

      result.current.resendMutation.mutate({
        tenant_id: 'test-id',
      });

      expect(result.current.isResending).toBe(true);
    });
  });

  it('works without callbacks', async () => {
    vi.spyOn(tenantApi, 'verifyOTP').mockResolvedValue({
      tenant_id: 'test-id',
      status: 'active',
      message: 'Success',
    });

    const { result } = renderHook(() => useOTPVerification(), {
      wrapper: createWrapper(),
    });

    result.current.verifyMutation.mutate({
      tenant_id: 'test-id',
      otp_code: '123456',
    });

    await waitFor(() => {
      expect(result.current.verifyMutation.isSuccess).toBe(true);
    });
  });
});
