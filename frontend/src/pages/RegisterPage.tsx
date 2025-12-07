/**
 * Registration page component
 * Wraps RegistrationForm with proper layout and metadata
 */

import { useEffect } from 'react';
import { RegistrationForm } from '../components/registration/RegistrationForm';

export function RegisterPage() {
  useEffect(() => {
    // Set page title
    document.title = 'Register - Keyles SSO';
  }, []);

  return (
    <main className="min-h-screen">
      <RegistrationForm />
    </main>
  );
}
