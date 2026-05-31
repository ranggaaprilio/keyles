/**
 * Resend OTP Button — Dell 1996 retro style
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
        className={`font-['Times_New_Roman',Times,serif] text-sm inline-flex items-center gap-2
          ${isDisabled
            ? 'text-gray-400 cursor-not-allowed'
            : 'text-[#0000ee] underline hover:text-[#551a8b]'
          } transition-colors`}
      >
        {isResending && <Loader2 className="w-4 h-4 animate-spin" />}
        {isResending ? 'Sending...' : isActive ? `Resend code in ${countdown}s` : 'Resend code'}
      </button>
      {!isActive && !isResending && (
        <p className="font-['Times_New_Roman',Times,serif] text-[11px] text-gray-500 mt-1">
          Didn&apos;t receive the code?
        </p>
      )}
    </div>
  );
}
