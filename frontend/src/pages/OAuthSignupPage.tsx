import { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
  getSignupContext,
  resendSignupCode,
  submitSignup,
  verifySignup,
} from '../services/oauthInteractionService';
import type { OAuthInteractionError } from '../types/oauth';
import { AuthShell } from '../components/layout/AuthShell';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';

const GENERIC_ERROR = 'An error occurred. Please try again.';

function isErrorWithCode(err: unknown): err is OAuthInteractionError {
  return typeof err === 'object' && err !== null && 'error' in err && typeof (err as { error: unknown }).error === 'string';
}

function signupErrorMessage(err: unknown): string {
  if (!isErrorWithCode(err)) return GENERIC_ERROR;
  switch (err.error) {
    case 'invalid_request':
      return err.error_description || 'Please check the details you entered.';
    case 'invalid_transaction':
      return 'This sign-in request has expired. Return to the application and try again.';
    case 'signup_disabled':
      return 'Account creation is not available for this application.';
    case 'signup_unavailable':
      return 'New accounts cannot be created at this time.';
    case 'invalid_code':
      return 'Invalid or expired verification code. Request a new code, or go back and submit your details again.';
    case 'throttled':
      return 'Too many attempts. Please try again later.';
    default:
      return GENERIC_ERROR;
  }
}

type Step = 'details' | 'verify';

export function OAuthSignupPage() {
  const [searchParams] = useSearchParams();
  const transactionId = searchParams.get('transaction_id') || '';
  const loginHref = `/oauth2/login?transaction_id=${encodeURIComponent(transactionId)}`;

  const [clientName, setClientName] = useState<string | null>(null);
  const [unavailable, setUnavailable] = useState<string | null>(
    transactionId ? null : 'Invalid or expired sign-up session.',
  );
  const [step, setStep] = useState<Step>('details');
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [code, setCode] = useState('');
  const [resendIn, setResendIn] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    if (!transactionId) return;
    let cancelled = false;
    getSignupContext(transactionId)
      .then((ctx) => {
        if (cancelled) return;
        setClientName(ctx.client_name);
        if (!ctx.signup_enabled) {
          setUnavailable('Account creation is not available for this application.');
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) setUnavailable(signupErrorMessage(err));
      });
    return () => {
      cancelled = true;
    };
  }, [transactionId]);

  useEffect(() => {
    if (resendIn <= 0) return;
    const timer = setTimeout(() => setResendIn((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [resendIn]);

  const handleDetailsSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      setError('Passwords do not match.');
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const result = await submitSignup({
        transaction_id: transactionId,
        email,
        password,
        display_name: displayName,
      });
      setResendIn(result.resend_after_seconds);
      setStep('verify');
    } catch (err: unknown) {
      setError(signupErrorMessage(err));
    } finally {
      setLoading(false);
    }
  };

  const handleVerifySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setNotice(null);
    try {
      const result = await verifySignup({ transaction_id: transactionId, email, code });
      window.location.href = result.redirect_url;
    } catch (err: unknown) {
      setError(signupErrorMessage(err));
      setLoading(false);
    }
  };

  const handleResend = async () => {
    setError(null);
    setNotice(null);
    try {
      const result = await resendSignupCode({ transaction_id: transactionId, email });
      setResendIn(result.resend_after_seconds);
      setNotice('If a sign-up is pending for this email, a new code has been sent.');
    } catch (err: unknown) {
      setError(signupErrorMessage(err));
    }
  };

  const handleChangeEmail = () => {
    setStep('details');
    setCode('');
    setError(null);
    setNotice(null);
  };

  const description = clientName
    ? `Create an account to continue to ${clientName}.`
    : 'Create an account to continue to the application.';

  if (unavailable) {
    return (
      <AuthShell title="Create an account" description={description}>
        <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
          {unavailable}
        </div>
        {transactionId && (
          <p className="mt-6 text-center text-sm text-muted-foreground">
            <Link to={loginHref} className="font-medium text-foreground underline-offset-4 hover:underline">
              Back to sign in
            </Link>
          </p>
        )}
      </AuthShell>
    );
  }

  if (step === 'verify') {
    return (
      <AuthShell title="Verify your email" description={`Enter the 6-digit code we sent to ${email}.`}>
        {error && (
          <div className="mb-4 rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">{error}</div>
        )}
        {notice && (
          <div className="mb-4 rounded-md border border-border bg-muted p-3 text-sm text-muted-foreground">{notice}</div>
        )}
        <form onSubmit={handleVerifySubmit} className="space-y-5">
          <div className="space-y-1.5">
            <Label htmlFor="code">Verification code</Label>
            <Input
              id="code"
              inputMode="numeric"
              autoComplete="one-time-code"
              pattern="\d{6}"
              maxLength={6}
              required
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
              disabled={loading}
            />
          </div>
          <Button type="submit" className="w-full" disabled={loading || code.length !== 6}>
            {loading ? 'Verifying...' : 'Verify and continue'}
          </Button>
        </form>
        <div className="mt-6 flex items-center justify-between text-sm">
          <button
            type="button"
            onClick={handleChangeEmail}
            className="text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
          >
            Back to details
          </button>
          <button
            type="button"
            onClick={handleResend}
            disabled={resendIn > 0}
            className="font-medium text-foreground underline-offset-4 hover:underline disabled:cursor-not-allowed disabled:text-muted-foreground disabled:no-underline"
          >
            {resendIn > 0 ? `Resend code in ${resendIn}s` : 'Resend code'}
          </button>
        </div>
      </AuthShell>
    );
  }

  return (
    <AuthShell
      title="Create an account"
      description={description}
      footer="If you already have an account, sign in instead."
    >
      {error && (
        <div className="mb-4 rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">{error}</div>
      )}
      <form onSubmit={handleDetailsSubmit} className="space-y-5">
        <div className="space-y-1.5">
          <Label htmlFor="display_name">Full name</Label>
          <Input
            id="display_name"
            autoComplete="name"
            required
            minLength={2}
            maxLength={100}
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            disabled={loading}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            autoComplete="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={loading}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            autoComplete="new-password"
            required
            minLength={8}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={loading}
          />
          <p className="text-xs text-muted-foreground">
            At least 8 characters with uppercase, lowercase, a number and one of @$!%*?&.
          </p>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="confirm_password">Confirm password</Label>
          <Input
            id="confirm_password"
            type="password"
            autoComplete="new-password"
            required
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            disabled={loading}
          />
        </div>
        <Button type="submit" className="w-full" disabled={loading}>
          {loading ? 'Creating account...' : 'Create account'}
        </Button>
      </form>
      <p className="mt-6 text-center text-sm text-muted-foreground">
        Already have an account?{' '}
        <Link to={loginHref} className="font-medium text-foreground underline-offset-4 hover:underline">
          Sign in
        </Link>
      </p>
    </AuthShell>
  );
}
