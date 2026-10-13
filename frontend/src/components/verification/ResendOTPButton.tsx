/**
 * Resend OTP Button
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
      startCountdown(60);
    },
    onResendError: (error) => {
      onError(error.message);
    },
  });

  useEffect(() => {
    startCountdown(60);
  }, [startCountdown]);

  const handleResend = () => {
    if (isActive || isResending) return;
    resendMutation.mutate({ tenant_id: tenantId });
  };

  const isDisabled = isActive || isResending;

  return (
    <div className="text-center">
      <button
        type="button"
        onClick={handleResend}
        disabled={isDisabled}
        className={`text-sm inline-flex items-center gap-2 transition-colors duration-fast
          ${isDisabled
            ? 'text-muted-foreground cursor-not-allowed'
            : 'font-medium text-foreground underline underline-offset-4 hover:text-muted-foreground'
          }`}
      >
        {isResending && <Loader2 className="w-4 h-4 animate-spin" />}
        {isResending ? 'Sending...' : isActive ? `Resend code in ${countdown}s` : 'Resend code'}
      </button>
      {!isActive && !isResending && (
        <p className="text-xs text-muted-foreground mt-1">
          Didn&apos;t receive the code?
        </p>
      )}
    </div>
  );
}
