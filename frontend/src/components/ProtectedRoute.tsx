/**
 * Protected Route Component
 * Redirects to /login if user is not authenticated
 */

import { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { isAuthenticated } from '../services/api/auth';

interface ProtectedRouteProps {
  children: ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}
