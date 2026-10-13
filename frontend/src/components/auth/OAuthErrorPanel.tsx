import { AlertTriangle } from 'lucide-react';

const ERROR_MESSAGES: Record<string, string> = {
  invalid_client: 'The application is not registered or is currently unavailable.',
  invalid_request: 'This sign-in request is invalid or has expired.',
  access_denied: 'Access was not granted to the application.',
  temporarily_unavailable: 'Sign-in is temporarily unavailable. Please try again later.',
  login_required: 'Please sign in to continue.',
  consent_required: 'Your approval is required to continue.',
};

interface OAuthErrorPanelProps {
  errorCode: string;
}

export function OAuthErrorPanel({ errorCode }: OAuthErrorPanelProps) {
  const message = ERROR_MESSAGES[errorCode] ?? 'The sign-in request could not be completed.';

  return (
    <div className="w-full max-w-md rounded-xl border border-border bg-background p-6 shadow-card sm:p-8">
      <div className="flex items-center gap-3">
        <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-destructive/10 text-destructive">
          <AlertTriangle className="h-5 w-5" />
        </span>
        <h1 className="text-xl font-medium tracking-tight text-foreground">
          Sign-in Error
        </h1>
      </div>
      <p className="mt-4 text-sm text-muted-foreground">{message}</p>
    </div>
  );
}
