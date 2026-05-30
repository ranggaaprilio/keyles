import { useState, useEffect } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAcceptInvitation } from '@/hooks/useInvitation';
import { ToastContainer, useToast } from '@/components/ui/toast';
import { Loader2, Mail, Lock, CheckCircle, AlertTriangle } from 'lucide-react';
import { isAxiosError } from 'axios';

const acceptSchema = z
  .object({
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

type AcceptFormData = z.infer<typeof acceptSchema>;

function InvitationExpired() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
        <div className="inline-flex items-center justify-center w-16 h-16 bg-red-100 rounded-full mb-4">
          <AlertTriangle className="w-8 h-8 text-red-600" />
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Invitation Expired</h1>
        <p className="text-gray-600 mb-6">
          This invitation link has expired. Please request a new invitation from your administrator.
        </p>
        <a
          href="/login"
          className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Go to Login
        </a>
      </div>
    </div>
  );
}

export function AcceptInvitationPage() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { toasts, closeToast, success, error: showError } = useToast();
  const [expired, setExpired] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const email = searchParams.get('email') ?? '';

  const acceptMutation = useAcceptInvitation();

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<AcceptFormData>({
    resolver: zodResolver(acceptSchema),
  });

  const password = watch('password', '');
  const confirmPassword = watch('confirmPassword', '');
  const passwordsMatch = password && confirmPassword && password === confirmPassword;

  useEffect(() => {
    if (acceptMutation.isSuccess) {
      success('Account activated! Redirecting to login...');
      setTimeout(() => {
        navigate('/login?activated=true', { replace: true });
      }, 1500);
    }
  }, [acceptMutation.isSuccess, navigate, success]);

  useEffect(() => {
    if (acceptMutation.isError) {
      const err = acceptMutation.error;
      if (isAxiosError(err) && err.response) {
        const status = err.response.status;
        if (status === 410) {
          setExpired(true);
          return;
        }
        if (status === 400 && err.response.data?.errors) {
          const apiErrors: Record<string, string> = {};
          for (const [key, value] of Object.entries(err.response.data.errors)) {
            apiErrors[key] = String(value);
          }
          setFieldErrors(apiErrors);
          return;
        }
      }
      showError(err instanceof Error ? err.message : 'Failed to accept invitation');
    }
  }, [acceptMutation.isError, acceptMutation.error, showError]);

  const onSubmit = (data: AcceptFormData) => {
    if (!token) return;
    setFieldErrors({});
    acceptMutation.mutate({ token, password: data.password });
  };

  if (expired) {
    return <InvitationExpired />;
  }

  if (acceptMutation.isSuccess) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-green-100 rounded-full mb-4">
            <CheckCircle className="w-8 h-8 text-green-600" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Account Activated!</h1>
          <p className="text-gray-600">Redirecting to login...</p>
        </div>
      </div>
    );
  }

  return (
    <>
      <ToastContainer toasts={toasts} onClose={closeToast} />
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4 py-12">
        <div className="max-w-md w-full">
          <div className="bg-white rounded-lg shadow-lg p-8">
            <div className="text-center mb-8">
              <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
                <Mail className="w-8 h-8 text-blue-600" />
              </div>
              <h1 className="text-2xl font-bold text-gray-900 mb-2">Accept Invitation</h1>
              <p className="text-sm text-gray-600">
                Set your password to activate your account
              </p>
            </div>

            {email && (
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email
                </label>
                <div className="flex items-center gap-2 px-3 py-2 bg-gray-100 border border-gray-200 rounded-lg text-gray-600">
                  <Mail className="w-4 h-4" />
                  <span className="text-sm">{email}</span>
                </div>
              </div>
            )}

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              <div>
                <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                  Password
                </label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                  <input
                    id="password"
                    type="password"
                    {...register('password')}
                    className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Enter your password"
                  />
                </div>
                {errors.password && (
                  <p className="mt-1 text-sm text-red-600">{errors.password.message}</p>
                )}
                {fieldErrors['password'] && (
                  <p className="mt-1 text-sm text-red-600">{fieldErrors['password']}</p>
                )}
              </div>

              <div>
                <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
                  Confirm Password
                </label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                  <input
                    id="confirmPassword"
                    type="password"
                    {...register('confirmPassword')}
                    className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Confirm your password"
                  />
                </div>
                {errors.confirmPassword && (
                  <p className="mt-1 text-sm text-red-600">{errors.confirmPassword.message}</p>
                )}
                {fieldErrors['confirmPassword'] && (
                  <p className="mt-1 text-sm text-red-600">{fieldErrors['confirmPassword']}</p>
                )}
              </div>

              <button
                type="submit"
                disabled={!passwordsMatch || acceptMutation.isPending}
                className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {acceptMutation.isPending ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Activating...
                  </>
                ) : (
                  'Activate Account'
                )}
              </button>
            </form>
          </div>
        </div>
      </div>
    </>
  );
}