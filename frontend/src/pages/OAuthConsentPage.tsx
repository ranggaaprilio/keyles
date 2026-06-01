import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ConsentScreen } from '../components/auth/ConsentScreen';
import { OAuthErrorPanel } from '../components/auth/OAuthErrorPanel';
import { getConsentDetails, submitConsentDecision } from '../services/oauthInteractionService';
import type { OAuthConsentDetails, OAuthInteractionError } from '../types/oauth';

function errorCode(error: unknown): string {
  return typeof error === 'object' && error !== null && 'error' in error
    ? String((error as OAuthInteractionError).error)
    : 'temporarily_unavailable';
}

export function OAuthConsentPage() {
  const [searchParams] = useSearchParams();
  const transactionId = searchParams.get('transaction_id') ?? '';
  const [details, setDetails] = useState<OAuthConsentDetails | null>(null);
  const [error, setError] = useState<string | null>(transactionId ? null : 'invalid_request');
  const [loading, setLoading] = useState(Boolean(transactionId));

  useEffect(() => {
    if (!transactionId) return;
    getConsentDetails(transactionId)
      .then(setDetails)
      .catch((reason: unknown) => setError(errorCode(reason)))
      .finally(() => setLoading(false));
  }, [transactionId]);

  const decide = async (approved: boolean) => {
    if (!details) return;
    try {
      const result = await submitConsentDecision({
        transaction_id: details.transaction_id,
        interaction_csrf_token: details.interaction_csrf_token,
        approved,
      });
      window.location.href = result.redirect_url;
    } catch (reason: unknown) {
      setError(errorCode(reason));
    }
  };

  if (error) {
    return <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4"><OAuthErrorPanel errorCode={error} /></div>;
  }
  if (loading || !details) {
    return <div className="min-h-screen bg-gray-100 flex items-center justify-center">Loading authorization request...</div>;
  }

  return (
    <ConsentScreen
      client={{
        client_id: details.client_id,
        client_name: details.client_name,
        ...(details.client_logo_uri ? { logo_uri: details.client_logo_uri } : {}),
      }}
      scopes={details.scopes}
      user={details.user_display}
      onApprove={() => decide(true)}
      onDeny={() => decide(false)}
    />
  );
}
