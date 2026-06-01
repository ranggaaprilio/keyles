import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { submitLogin } from '../services/oauthInteractionService';
import type { OAuthInteractionError } from '../types/oauth';

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
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8">
        <h1 className="text-2xl font-bold text-gray-900 mb-6 text-center">Sign In</h1>
        {!transactionId && (
          <div className="mb-4 p-3 bg-red-50 text-red-700 rounded">
            Invalid or expired login session.
          </div>
        )}
        {error && (
          <div className="mb-4 p-3 bg-red-50 text-red-700 rounded">{error}</div>
        )}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700">Email</label>
            <input
              id="email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={loading || !transactionId}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2 border"
            />
          </div>
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700">Password</label>
            <input
              id="password"
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={loading || !transactionId}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2 border"
            />
          </div>
          <button
            type="submit"
            disabled={loading || !transactionId}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            {loading ? 'Signing in...' : 'Sign In'}
          </button>
        </form>
      </div>
    </div>
  );
}