/**
 * OTP Verification Page
 */

import { useNavigate, useLocation } from 'react-router-dom';
import { OTPVerificationForm } from '../components/verification/OTPVerificationForm';
import { ResendOTPButton } from '../components/verification/ResendOTPButton';
import { ToastContainer, useToast } from '../components/ui/toast';
import { AuthShell } from '../components/layout/AuthShell';

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
      <AuthShell
        title="Verify email"
        description={
          <>
            We&apos;ve sent a 6-digit code to{' '}
            <strong className="font-medium text-foreground">{email || 'your email'}</strong>
            {organizationName && (
              <>
                {' '}for <strong className="font-medium text-foreground">{organizationName}</strong>
              </>
            )}
            .
          </>
        }
        footer={
          <div className="space-y-3">
            <p>
              Having trouble? Check your spam folder or{' '}
              <a href="mailto:support@keyles.io" className="text-foreground underline underline-offset-4 hover:text-muted-foreground">
                contact support
              </a>
              .
            </p>
            <p className="rounded-md border border-border bg-background px-3 py-2 text-left">
              <span className="font-medium text-foreground">Security note:</span> This code expires
              in 10 minutes. Never share your verification code with anyone.
            </p>
          </div>
        }
      >
        <OTPVerificationForm
          tenantId={tenantId}
          onSuccess={handleVerifySuccess}
          onError={handleVerifyError}
        />

        <div className="mt-5">
          <ResendOTPButton
            tenantId={tenantId}
            onSuccess={handleResendSuccess}
            onError={handleResendError}
          />
        </div>
      </AuthShell>
    </>
  );
}
