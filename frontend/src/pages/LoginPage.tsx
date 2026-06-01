/**
 * Login Page Component — Dell 1996 retro style
 */

import { useState, useEffect } from 'react';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '../hooks/useAuth';
import { ToastContainer, useToast } from '../components/ui/toast';
import { Loader2 } from 'lucide-react';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(1, 'Password is required'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, isLoggingIn, loginMutation, isAuthenticated } = useAuth();
  const { toasts, removeToast, success, error } = useToast();
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    const state = location.state as { message?: string } | null;
    if (state?.message) {
      success(state.message);
      window.history.replaceState({}, '');
    }
  }, [location.state, success]);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    if (loginMutation.isSuccess) {
      success('Login successful! Redirecting...');
      setTimeout(() => navigate('/dashboard', { replace: true }), 500);
    }
    if (loginMutation.isError) {
      const err = loginMutation.error as { message?: string };
      error(err?.message || 'Login failed. Please check your credentials.');
    }
  }, [loginMutation.isSuccess, loginMutation.isError, loginMutation.error, navigate, success, error]);

  const onSubmit = (data: LoginFormData) => {
    login(data);
  };

  return (
    <>
      <ToastContainer toasts={toasts} onClose={removeToast} />
      <div className="min-h-screen bg-white flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          {/* Section eyebrow */}
          <div className="bg-[#8c9ae0] px-4 py-4 mb-0">
            <h1 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[28px] font-black uppercase leading-[1.0] text-black">
              WELCOME BACK
            </h1>
          </div>

          {/* Form card — ribbon card style */}
          <div className="border-x border-b border-black">
            <div className="border-b border-black bg-white px-3 py-1.5">
              <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
                SIGN IN TO YOUR ACCOUNT
              </h3>
            </div>
            <div className="bg-[#9ab6c8] px-4 py-4">
              <form
                onSubmit={handleSubmit(onSubmit)}
                className="space-y-4"
                aria-label="Login form"
                noValidate
              >
                {/* Email Field */}
                <div>
                  <label
                    htmlFor="email"
                    className="block font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black mb-1"
                  >
                    Email Address
                  </label>
                  <input
                    {...register('email')}
                    type="email"
                    id="email"
                    disabled={isLoggingIn}
                    aria-label="Email address"
                    aria-invalid={errors.email ? 'true' : 'false'}
                    aria-describedby={errors.email ? 'email-error' : undefined}
                    aria-required="true"
                    className={`w-full px-2 py-1.5 border font-['Times_New_Roman',Times,serif] text-sm focus:outline-none focus:ring-2 focus:ring-dell-red focus:ring-offset-0
                      ${errors.email ? 'border-red-700' : 'border-black'}
                      ${isLoggingIn ? 'bg-gray-100 cursor-not-allowed' : 'bg-white'}`}
                    placeholder="you@example.com"
                  />
                  {errors.email && (
                    <p id="email-error" className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-700" role="alert">
                      {errors.email.message}
                    </p>
                  )}
                </div>

                {/* Password Field */}
                <div>
                  <label
                    htmlFor="password"
                    className="block font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] text-black mb-1"
                  >
                    Password
                  </label>
                  <div className="relative">
                    <input
                      {...register('password')}
                      type={showPassword ? 'text' : 'password'}
                      id="password"
                      disabled={isLoggingIn}
                      aria-label="Password"
                      aria-invalid={errors.password ? 'true' : 'false'}
                      aria-describedby={errors.password ? 'password-error' : undefined}
                      aria-required="true"
                      className={`w-full px-2 py-1.5 pr-14 border font-['Times_New_Roman',Times,serif] text-sm focus:outline-none focus:ring-2 focus:ring-dell-red focus:ring-offset-0
                        ${errors.password ? 'border-red-700' : 'border-black'}
                        ${isLoggingIn ? 'bg-gray-100 cursor-not-allowed' : 'bg-white'}`}
                      placeholder="••••••••"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      aria-label={showPassword ? 'Hide password' : 'Show password'}
                      aria-pressed={showPassword}
                      className="absolute right-2 top-1/2 -translate-y-1/2 font-[Helvetica,Arial,system-ui,sans-serif] text-[11px] font-bold uppercase text-gray-500 hover:text-black"
                    >
                      {showPassword ? 'HIDE' : 'SHOW'}
                    </button>
                  </div>
                  {errors.password && (
                    <p id="password-error" className="mt-1 font-['Times_New_Roman',Times,serif] text-sm text-red-700" role="alert">
                      {errors.password.message}
                    </p>
                  )}
                </div>

                {/* Submit Button */}
                <button
                  type="submit"
                  disabled={isLoggingIn}
                  className={`w-full flex items-center justify-center gap-2 px-4 py-2 border border-black font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1.5px]
                    ${isLoggingIn
                      ? 'bg-gray-300 cursor-not-allowed text-gray-500'
                      : 'bg-black text-white hover:bg-gray-800'
                    } transition-colors`}
                >
                  {isLoggingIn && <Loader2 className="w-4 h-4 animate-spin" />}
                  {isLoggingIn ? 'SIGNING IN...' : 'SIGN IN'}
                </button>
              </form>

              <div className="mt-4 pt-3 border-t border-black text-center">
                <p className="font-['Times_New_Roman',Times,serif] text-sm text-black">
                  Don&apos;t have an account?{' '}
                  <Link
                    to="/register"
                    className="text-[#0000ee] underline hover:text-[#551a8b]"
                  >
                    Create one now
                  </Link>
                </p>
              </div>
            </div>
          </div>

          {/* Help text */}
          <div className="mt-4 text-center">
            <p className="font-['Times_New_Roman',Times,serif] text-[11px] text-gray-500">
              Having trouble?{' '}
              <a href="mailto:support@keyles.io" className="text-[#0000ee] underline">
                Contact support
              </a>
            </p>
          </div>
        </div>
      </div>
    </>
  );
}
