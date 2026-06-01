import { useSearchParams } from 'react-router-dom';
import { OAuthErrorPanel } from '../components/auth/OAuthErrorPanel';

export function OAuthErrorPage() {
  const [searchParams] = useSearchParams();
  const errorCode = searchParams.get('error') ?? 'unknown_error';

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <OAuthErrorPanel errorCode={errorCode} />
    </div>
  );
}
