import { useEffect, useState } from 'react';
import { submitLogout } from '../services/oauthInteractionService';

export function OAuthLogoutPage() {
  const [complete, setComplete] = useState(false);

  useEffect(() => {
    void (async () => {
      try {
        await submitLogout();
      } catch {
        // Cookie expiry is best-effort from the browser's perspective.
      }
      setComplete(true);
    })();
  }, []);

  return (
    <div className="min-h-screen bg-muted flex items-center justify-center p-4">
      <div className="w-full max-w-md rounded-xl border border-border bg-background p-6 shadow-card sm:p-8">
        <h1 className="text-xl font-medium tracking-tight text-foreground">
          {complete ? 'Signed Out' : 'Signing Out'}
        </h1>
        <p className="mt-3 text-sm text-muted-foreground">
          {complete ? 'Your Keyles browser session has ended.' : 'Ending your Keyles browser session...'}
        </p>
      </div>
    </div>
  );
}
