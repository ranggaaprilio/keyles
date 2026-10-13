/**
 * Login Page Component
 */

import { useState, useEffect } from 'react';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '../hooks/useAuth';
import { ToastContainer, useToast } from '../components/ui/toast';
import { AuthShell } from '../components/layout/AuthShell';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
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
      // Clear via the router so location.state actually changes; the raw
      // history API leaves React Router's state intact and re-triggers this.
      navigate(location.pathname, { replace: true, state: null });
    }
  }, [location.state, location.pathname, navigate, success]);

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
      <AuthShell
        title="Welcome back"
        description="Sign in to your account to continue."
        footer={
          <>
            Having trouble?{' '}
            <a href="mailto:support@keyles.io" className="text-foreground underline underline-offset-4 hover:text-muted-foreground">
              Contact support
            </a>
          </>
        }
      >
        <form
          onSubmit={handleSubmit(onSubmit)}
          className="space-y-5"
          aria-label="Login form"
          noValidate
        >
          {/* Email Field */}
          <div className="space-y-1.5">
            <Label htmlFor="email">Email Address</Label>
            <Input
              {...register('email')}
              type="email"
              id="email"
              disabled={isLoggingIn}
              aria-label="Email address"
              aria-invalid={errors.email ? 'true' : 'false'}
              aria-describedby={errors.email ? 'email-error' : undefined}
              aria-required="true"
              className={errors.email ? 'border-destructive' : ''}
              placeholder="you@example.com"
            />
            {errors.email && (
              <p id="email-error" className="text-sm text-destructive" role="alert">
                {errors.email.message}
              </p>
            )}
          </div>

          {/* Password Field */}
          <div className="space-y-1.5">
            <Label htmlFor="password">Password</Label>
            <div className="relative">
              <Input
                {...register('password')}
                type={showPassword ? 'text' : 'password'}
                id="password"
                disabled={isLoggingIn}
                aria-label="Password"
                aria-invalid={errors.password ? 'true' : 'false'}
                aria-describedby={errors.password ? 'password-error' : undefined}
                aria-required="true"
                className={`pr-16 ${errors.password ? 'border-destructive' : ''}`}
                placeholder="••••••••"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
                aria-pressed={showPassword}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-medium text-muted-foreground hover:text-foreground"
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
            </div>
            {errors.password && (
              <p id="password-error" className="text-sm text-destructive" role="alert">
                {errors.password.message}
              </p>
            )}
          </div>

          <Button type="submit" className="w-full" disabled={isLoggingIn}>
            {isLoggingIn && <Loader2 className="w-4 h-4 animate-spin" />}
            {isLoggingIn ? 'Signing in...' : 'Sign in'}
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          Don&apos;t have an account?{' '}
          <Link to="/register" className="font-medium text-foreground underline underline-offset-4 hover:text-muted-foreground">
            Create one now
          </Link>
        </p>
      </AuthShell>
    </>
  );
}
