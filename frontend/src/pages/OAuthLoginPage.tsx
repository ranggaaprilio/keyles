import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { submitLogin } from '../services/oauthInteractionService';
import type { OAuthInteractionError } from '../types/oauth';
import { AuthShell } from '../components/layout/AuthShell';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';

function isErrorWithCode(err: unknown): err is OAuthInteractionError {
  return typeof err === 'object' && err !== null && 'error' in err && typeof (err as { error: unknown }).error === 'string';
}

export function OAuthLoginPage() {
  const [searchParams] = useSearchParams();
  const transactionId = searchParams.get('transaction_id') || '';
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!transactionId) {
      setError('Invalid or expired login session.');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const result = await submitLogin({
        transaction_id: transactionId,
        email,
        password,
      });
      window.location.href = result.redirect_url;
    } catch (err: unknown) {
      if (isErrorWithCode(err)) {
        if (err.error === 'throttled') {
          setError('Too many login attempts. Please try again later.');
        } else if (err.error === 'invalid_credentials') {
          setError('Invalid email or password.');
        } else {
          setError('An error occurred. Please try again.');
        }
      } else {
        setError('An error occurred. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthShell title="Sign in" description="Continue to the application with your Keyles account.">
      {!transactionId && (
        <div className="mb-4 rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
          Invalid or expired login session.
        </div>
      )}
      {error && (
        <div className="mb-4 rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">{error}</div>
      )}
      <form onSubmit={handleSubmit} className="space-y-5">
        <div className="space-y-1.5">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={loading || !transactionId}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={loading || !transactionId}
          />
        </div>
        <Button
          type="submit"
          className="w-full"
          disabled={loading || !transactionId}
        >
          {loading ? 'Signing in...' : 'Sign In'}
        </Button>
      </form>
    </AuthShell>
  );
}
