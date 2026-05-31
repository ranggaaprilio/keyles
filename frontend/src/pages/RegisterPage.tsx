/**
 * Registration page component — Dell 1996 retro style
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
