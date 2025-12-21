/**
 * Resend OTP Button Component
 * With countdown timer using Zustand
 */

import { useEffect } from 'react';
import { useOTPStore } from '../../stores/otpStore';
import { useOTPVerification } from '../../hooks/useOTPVerification';
import { Loader2 } from 'lucide-react';

interface ResendOTPButtonProps {
  tenantId: string;
  onSuccess: (message: string) => void;
  onError: (message: string) => void;
}

export function ResendOTPButton({
  tenantId,
  onSuccess,
  onError,
}: ResendOTPButtonProps) {
  const { countdown, isActive, startCountdown } = useOTPStore();

  const { resendMutation, isResending } = useOTPVerification({
    onResendSuccess: (data) => {
      onSuccess(data.message);
      startCountdown(60); // Start 60 second countdown
    },
    onResendError: (error) => {
      onError(error.message);
    },
  });

  useEffect(() => {
    // Start initial countdown on mount (assuming OTP was just sent during registration)
    startCountdown(60);
  }, [startCountdown]);

  const handleResend = () => {
    if (isActive || isResending) return;

    resendMutation.mutate({
      tenant_id: tenantId,
    });
  };

  const isDisabled = isActive || isResending;

  return (
    <div className="text-center">
      <button
        type="button"
        onClick={handleResend}
        disabled={isDisabled}
        className={`text-sm font-medium ${
          isDisabled
            ? 'text-gray-400 cursor-not-allowed'
            : 'text-blue-600 hover:text-blue-700 hover:underline'
        } transition-colors inline-flex items-center gap-2`}
      >
        {isResending && <Loader2 className="w-4 h-4 animate-spin" />}
        {isResending ? 'Sending...' : isActive ? `Resend code in ${countdown}s` : 'Resend code'}
      </button>
      {!isActive && !isResending && (
        <p className="text-xs text-gray-500 mt-1">
          Didn't receive the code?
        </p>
      )}
    </div>
  );
}
