/**
 * OTP Verification Page
 * Integrates OTPVerificationForm and ResendOTPButton
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
  const { toasts, closeToast, success, error } = useToast();

  const state = location.state as LocationState;
  const tenantId = state?.tenantId;
  const organizationName = state?.organizationName;
  const email = state?.email;

  // Redirect to register if no tenant ID
  if (!tenantId) {
    navigate('/register', { replace: true });
    return null;
  }

  const handleVerifySuccess = () => {
    success('Email verified successfully! Redirecting to login...');
    setTimeout(() => {
      navigate('/login', { replace: true });
    }, 2000);
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
      <ToastContainer toasts={toasts} onClose={closeToast} />
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4 py-12">
        <div className="max-w-md w-full">
          <div className="bg-white rounded-lg shadow-lg p-8">
            {/* Header */}
            <div className="text-center mb-8">
              <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
                <Mail className="w-8 h-8 text-blue-600" />
              </div>
              <h1 className="text-2xl font-bold text-gray-900 mb-2">
                Verify Your Email
              </h1>
              <p className="text-sm text-gray-600">
                We've sent a 6-digit verification code to
              </p>
              <p className="text-sm font-medium text-gray-900 mt-1">
                {email || 'your email address'}
              </p>
              {organizationName && (
                <p className="text-xs text-gray-500 mt-2">
                  for <span className="font-medium">{organizationName}</span>
                </p>
              )}
            </div>

            {/* OTP Form */}
            <OTPVerificationForm
              tenantId={tenantId}
              onSuccess={handleVerifySuccess}
              onError={handleVerifyError}
            />

            {/* Resend Button */}
            <div className="mt-6">
              <ResendOTPButton
                tenantId={tenantId}
                onSuccess={handleResendSuccess}
                onError={handleResendError}
              />
            </div>

            {/* Help Section */}
            <div className="mt-8 pt-6 border-t border-gray-200">
              <p className="text-xs text-gray-500 text-center">
                Having trouble? Check your spam folder or{' '}
                <a
                  href="mailto:support@keyles.io"
                  className="text-blue-600 hover:underline"
                >
                  contact support
                </a>
              </p>
            </div>
          </div>

          {/* Security Note */}
          <div className="mt-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            <p className="text-xs text-yellow-800">
              <strong>Security Note:</strong> This code expires in 10 minutes. Never share
              your verification code with anyone.
            </p>
          </div>
        </div>
      </div>
    </>
  );
}
