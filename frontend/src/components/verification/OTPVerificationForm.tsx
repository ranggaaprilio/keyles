/**
 * OTP Verification Form Component
 * 6-digit OTP input with auto-focus and validation
 */

import { useState, useRef, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useOTPVerification } from '../../hooks/useOTPVerification';
import { Loader2 } from 'lucide-react';

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
    onVerifySuccess: () => {
      onSuccess();
    },
    onVerifyError: (error) => {
      onError(error.message);
    },
  });

  useEffect(() => {
    // Focus first input on mount
    inputRefs.current[0]?.focus();
  }, []);

  const handleDigitChange = (index: number, value: string) => {
    // Only allow single digit
    const digit = value.slice(-1);
    if (digit && !/^\d$/.test(digit)) return;

    const newDigits = [...otpDigits];
    newDigits[index] = digit;
    setOtpDigits(newDigits);

    // Update form value
    const otpValue = newDigits.join('');
    setValue('otp', otpValue);

    // Auto-focus next input
    if (digit && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace') {
      if (!otpDigits[index] && index > 0) {
        // Move to previous input if current is empty
        inputRefs.current[index - 1]?.focus();
      } else {
        // Clear current input
        const newDigits = [...otpDigits];
        newDigits[index] = '';
        setOtpDigits(newDigits);
        setValue('otp', newDigits.join(''));
      }
    } else if (e.key === 'ArrowLeft' && index > 0) {
      inputRefs.current[index - 1]?.focus();
    } else if (e.key === 'ArrowRight' && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const pastedData = e.clipboardData.getData('text').trim();
    
    if (/^\d{6}$/.test(pastedData)) {
      const digits = pastedData.split('');
      setOtpDigits(digits);
      setValue('otp', pastedData);
      inputRefs.current[5]?.focus();
    }
  };

  const onSubmit = (data: OTPFormData) => {
    verifyMutation.mutate({
      tenant_id: tenantId,
      otp_code: data.otp,
    });
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div className="space-y-2">
        <label className="block text-sm font-medium text-gray-700">
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
              className={`w-12 h-14 text-center text-2xl font-bold border-2 rounded-lg
                ${errors.otp ? 'border-red-500' : 'border-gray-300'}
                focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent
                disabled:bg-gray-100 disabled:cursor-not-allowed
                transition-colors`}
              aria-label={`Digit ${index + 1}`}
            />
          ))}
        </div>
        {errors.otp && (
          <p className="text-sm text-red-600 text-center">{errors.otp.message}</p>
        )}
      </div>

      <button
        type="submit"
        disabled={isVerifying || otpDigits.some((d) => !d)}
        className={`w-full flex items-center justify-center gap-2 px-4 py-3 rounded-lg font-medium
          ${isVerifying || otpDigits.some((d) => !d)
            ? 'bg-gray-300 cursor-not-allowed text-gray-500'
            : 'bg-blue-600 hover:bg-blue-700 text-white'
          }
          transition-colors`}
      >
        {isVerifying && <Loader2 className="w-5 h-5 animate-spin" />}
        {isVerifying ? 'Verifying...' : 'Verify Email'}
      </button>
    </form>
  );
}
