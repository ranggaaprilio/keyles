/**
 * OTP Verification Page — Dell 1996 retro style
 */

import { useNavigate, useLocation } from 'react-router-dom';
import { OTPVerificationForm } from '../components/verification/OTPVerificationForm';
import { ResendOTPButton } from '../components/verification/ResendOTPButton';
import { ToastContainer, useToast } from '../components/ui/toast';
import { Mail } from 'lucide-react';

interface LocationState {
  tenantId?: string;
  organizationName?: string;
  email?: string;
}

export function VerifyOTPPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { toasts, removeToast, success, error } = useToast();

  const state = location.state as LocationState;
  const tenantId = state?.tenantId;
  const organizationName = state?.organizationName;
  const email = state?.email;

  if (!tenantId) {
    navigate('/register', { replace: true });
    return null;
  }

  const handleVerifySuccess = () => {
    success('Email verified successfully! Redirecting to login...');
    setTimeout(() => navigate('/login', {
      state: { message: 'Account verified. Please sign in.' }
    }), 1500);
  };

  const handleVerifyError = (message: string) => {
    error(message);
  };

  const handleResendSuccess = (message: string) => {
    success(message);
  };

  const handleResendError = (message: string) => {
    error(message);
  };

  return (
    <>
      <ToastContainer toasts={toasts} onClose={removeToast} />
      <div className="min-h-screen bg-white flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          {/* Section eyebrow — sky */}
          <div className="bg-[#9ab6c8] px-4 py-4 mb-0">
            <div className="flex items-center gap-2">
              <Mail className="w-5 h-5 text-black" />
              <h1 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[28px] font-black uppercase leading-[1.0] text-black">
                VERIFY EMAIL
              </h1>
            </div>
          </div>

          {/* Form card — ribbon card style */}
          <div className="border-x border-b border-black">
            <div className="border-b border-black bg-white px-3 py-1.5">
              <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                ENTER VERIFICATION CODE
              </h3>
            </div>
            <div className="bg-[#b3bd95] px-4 py-4">
              <p className="mb-3 font-['Times_New_Roman',Times,serif] text-sm text-black">
                We&apos;ve sent a 6-digit code to{' '}
                <strong>{email || 'your email'}</strong>
              </p>
              {organizationName && (
                <p className="mb-3 font-['Times_New_Roman',Times,serif] text-[11px] text-gray-700">
                  for <strong>{organizationName}</strong>
                </p>
              )}

              <OTPVerificationForm
                tenantId={tenantId}
                onSuccess={handleVerifySuccess}
                onError={handleVerifyError}
              />

              <div className="mt-4">
                <ResendOTPButton
                  tenantId={tenantId}
                  onSuccess={handleResendSuccess}
                  onError={handleResendError}
                />
              </div>
            </div>
          </div>

          {/* Help text */}
          <div className="mt-4 pt-3 border-x border-b border-black bg-white px-4">
            <p className="font-['Times_New_Roman',Times,serif] text-[11px] text-gray-500 text-center">
              Having trouble? Check your spam folder or{' '}
              <a href="mailto:support@keyles.io" className="text-[#0000ee] underline">
                contact support
              </a>
            </p>
          </div>

          {/* Security note — yellow sticker */}
          <div className="mt-4 border border-black bg-[#fcc20f] px-3 py-2">
            <p className="font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold text-black">
              SECURITY NOTE: This code expires in 10 minutes. Never share
              your verification code with anyone.
            </p>
          </div>
        </div>
      </div>
    </>
  );
}
