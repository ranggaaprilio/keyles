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
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md border border-black bg-white p-5 shadow-[2px_2px_0_#000]">
        <h1 className="font-[Helvetica,Arial,system-ui,sans-serif] text-lg font-bold uppercase tracking-[1px]">
          {complete ? 'Signed Out' : 'Signing Out'}
        </h1>
        <p className="mt-3 font-['Times_New_Roman',Times,serif] text-sm">
          {complete ? 'Your Keyles browser session has ended.' : 'Ending your Keyles browser session...'}
        </p>
      </div>
    </div>
  );
}
