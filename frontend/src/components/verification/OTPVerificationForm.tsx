/**
 * OTP Verification Form
 */

import { useState, useRef, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useOTPVerification } from '../../hooks/useOTPVerification';
import { Loader2 } from 'lucide-react';
import { Button } from '../ui/button';

const otpSchema = z.object({
  otp: z.string().length(6, 'OTP must be exactly 6 digits').regex(/^\d{6}$/, 'OTP must contain only digits'),
});

type OTPFormData = z.infer<typeof otpSchema>;

interface OTPVerificationFormProps {
  tenantId: string;
  onSuccess: () => void;
  onError: (message: string) => void;
}

export function OTPVerificationForm({
  tenantId,
  onSuccess,
  onError,
}: OTPVerificationFormProps) {
  const [otpDigits, setOtpDigits] = useState<string[]>(['', '', '', '', '', '']);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const { handleSubmit, formState: { errors }, setValue } = useForm<OTPFormData>({
    resolver: zodResolver(otpSchema),
  });

  const { verifyMutation, isVerifying } = useOTPVerification({
    onVerifySuccess: onSuccess,
    onVerifyError: (err) => onError(err.message),
  });

  useEffect(() => {
    inputRefs.current[0]?.focus();
  }, []);

  const handleDigitChange = (index: number, value: string) => {
    if (!/^\d*$/.test(value)) return;

    const newDigits = [...otpDigits];
    newDigits[index] = value.slice(-1);
    setOtpDigits(newDigits);

    // Auto-focus next
    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }

    setValue('otp', newDigits.join(''));
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !otpDigits[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const pasted = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
    if (!pasted) return;

    const newDigits = [...otpDigits];
    for (let i = 0; i < pasted.length; i++) {
      newDigits[i] = pasted[i] ?? '';
    }
    setOtpDigits(newDigits);
    setValue('otp', newDigits.join(''));

    const focusIndex = Math.min(pasted.length, 5);
    inputRefs.current[focusIndex]?.focus();
  };

  const onSubmit = () => {
    const otp = otpDigits.join('');
    verifyMutation.mutate({ tenant_id: tenantId, otp_code: otp });
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-foreground mb-3 text-center">
          Enter Verification Code
        </label>
        <div className="flex gap-2 justify-center">
          {otpDigits.map((digit, index) => (
            <input
              key={index}
              ref={(el) => (inputRefs.current[index] = el)}
              type="text"
              inputMode="numeric"
              maxLength={1}
              value={digit}
              onChange={(e) => handleDigitChange(index, e.target.value)}
              onKeyDown={(e) => handleKeyDown(index, e)}
              onPaste={handlePaste}
              disabled={isVerifying}
              className={`h-12 w-10 rounded-md border bg-background text-center text-xl font-semibold text-foreground transition-colors duration-fast
                ${errors.otp ? 'border-destructive' : 'border-input'}
                focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-0 focus:border-foreground
                disabled:bg-muted disabled:cursor-not-allowed`}
              aria-label={`Digit ${index + 1}`}
            />
          ))}
        </div>
        {errors.otp && (
          <p className="mt-2 text-sm text-destructive text-center">
            {errors.otp.message}
          </p>
        )}
      </div>

      <Button
        type="submit"
        className="w-full"
        disabled={isVerifying || otpDigits.some((d) => !d)}
      >
        {isVerifying && <Loader2 className="w-4 h-4 animate-spin" />}
        {isVerifying ? 'Verifying...' : 'Verify email'}
      </Button>
    </form>
  );
}
