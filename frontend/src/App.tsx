/**
 * Main Application Component
 */

import { BrowserRouter, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ErrorBoundary } from "./components/ErrorBoundary";
import { RegisterPage } from "./pages/RegisterPage";
import { VerifyOTPPage } from "./pages/VerifyOTPPage";
import { LoginPage } from "./pages/LoginPage";
import { DashboardPage } from "./pages/DashboardPage";
import LandingPage from "./pages/LandingPage";
import ScrollToHashElement from "./components/landing/ScrollToHashElement";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { ClientManagementPage } from "./pages/ClientManagementPage";
import { UserManagementPage } from "./pages/UserManagementPage";
import { AcceptInvitationPage } from "./pages/AcceptInvitationPage";
import { OAuthLoginPage } from './pages/OAuthLoginPage';
import { OAuthSignupPage } from './pages/OAuthSignupPage';
import { OAuthConsentPage } from './pages/OAuthConsentPage';
import { OAuthErrorPage } from './pages/OAuthErrorPage';
import { OAuthLogoutPage } from './pages/OAuthLogoutPage';
import { DashboardLayout } from "./components/dashboard/DashboardLayout";
import { IntegrationGuidePage } from "./pages/IntegrationGuidePage";

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
});

// 404 Page
function NotFoundPage() {
  return (
    <div className="min-h-screen bg-muted flex items-center justify-center px-4">
      <div className="w-full max-w-md rounded-xl border border-border bg-background p-8 text-center shadow-card">
        <p className="text-5xl font-normal tracking-[-0.03em] text-foreground">404</p>
        <p className="mt-3 text-sm text-muted-foreground mb-6">
          Page not found. The requested resource does not exist.
        </p>
        <a
          href="/"
          className="inline-flex h-10 items-center justify-center rounded-full bg-primary px-5 text-sm font-medium text-primary-foreground hover:bg-primary/85 transition-colors duration-fast"
        >
          Go Home
        </a>
      </div>
    </div>
  );
}

export function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <ScrollToHashElement />
          <Routes>
            {/* Landing Page at root */}
            <Route path="/" element={<LandingPage />} />

            {/* Public routes */}
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/verify-otp" element={<VerifyOTPPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/invite/:token" element={<AcceptInvitationPage />} />
            <Route path="/oauth2/login" element={<OAuthLoginPage />} />
            <Route path="/oauth2/signup" element={<OAuthSignupPage />} />
            <Route path="/oauth2/consent" element={<OAuthConsentPage />} />
            <Route path="/oauth2/error" element={<OAuthErrorPage />} />
            <Route path="/oauth2/logout" element={<OAuthLogoutPage />} />
            <Route path="/docs/oauth" element={<IntegrationGuidePage />} />

            {/* Protected routes — wrapped in DashboardLayout */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <DashboardPage />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/clients"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <ClientManagementPage />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/clients/new"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <ClientManagementPage />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/clients/:clientId"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <ClientManagementPage />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/clients/:clientId/edit"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <ClientManagementPage />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/users"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <UserManagementPage />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/users/:userId"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <UserManagementPage />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/integration"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <IntegrationGuidePage />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />

            {/* 404 */}
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}
