/**
 * Registration page component
 */

import { useEffect } from 'react';
import { RegistrationForm } from '../components/registration/RegistrationForm';

export function RegisterPage() {
  useEffect(() => {
    document.title = 'Register - Keyles SSO';
  }, []);

  return (
    <main className="min-h-screen">
      <RegistrationForm />
    </main>
  );
}
