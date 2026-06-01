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
    <div className="w-full max-w-md border border-black bg-white shadow-[2px_2px_0_#000]">
      <div className="bg-[#d77a7a] px-5 py-4">
        <h1 className="font-[Helvetica,Arial,system-ui,sans-serif] text-lg font-bold uppercase tracking-[1px]">
          Sign-in Error
        </h1>
      </div>
      <div className="p-5">
        <p className="font-['Times_New_Roman',Times,serif] text-sm text-black">{message}</p>
      </div>
    </div>
  );
}
