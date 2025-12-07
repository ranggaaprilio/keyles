/**
 * Hook for OTP verification and resend operations
 */

import { useMutation, UseMutationResult } from '@tanstack/react-query';
import { verifyOTP, resendOTP } from '../services/api/tenant';
import {
  VerifyOTPRequest,
  VerifyOTPResponse,
  ResendOTPRequest,
  ResendOTPResponse,
} from '../types/tenant';
import { ApiException } from '../types/api';

interface UseOTPVerificationOptions {
  onVerifySuccess?: (data: VerifyOTPResponse) => void;
  onVerifyError?: (error: ApiException) => void;
  onResendSuccess?: (data: ResendOTPResponse) => void;
  onResendError?: (error: ApiException) => void;
}

interface UseOTPVerificationReturn {
  verifyMutation: UseMutationResult<
    VerifyOTPResponse,
    ApiException,
    VerifyOTPRequest
  >;
  resendMutation: UseMutationResult<
    ResendOTPResponse,
    ApiException,
    ResendOTPRequest
  >;
  isVerifying: boolean;
  isResending: boolean;
}

export function useOTPVerification(
  options: UseOTPVerificationOptions = {}
): UseOTPVerificationReturn {
  const {
    onVerifySuccess,
    onVerifyError,
    onResendSuccess,
    onResendError,
  } = options;

  const verifyMutation = useMutation<
    VerifyOTPResponse,
    ApiException,
    VerifyOTPRequest
  >({
    mutationFn: verifyOTP,
    ...(onVerifySuccess && { onSuccess: onVerifySuccess }),
    ...(onVerifyError && { onError: onVerifyError }),
  });

  const resendMutation = useMutation<
    ResendOTPResponse,
    ApiException,
    ResendOTPRequest
  >({
    mutationFn: resendOTP,
    ...(onResendSuccess && { onSuccess: onResendSuccess }),
    ...(onResendError && { onError: onResendError }),
  });

  return {
    verifyMutation,
    resendMutation,
    isVerifying: verifyMutation.isPending,
    isResending: resendMutation.isPending,
  };
}
