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
    <div className="min-h-screen bg-white flex items-center justify-center px-4">
      <div className="w-full max-w-md text-center">
        <div className="bg-[#d77a7a] px-4 py-6">
          <h1 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[48px] font-black uppercase leading-[1.0] text-black">404</h1>
        </div>
        <div className="border-x border-b border-black bg-white p-4">
          <p className="font-['Times_New_Roman',Times,serif] text-sm text-black mb-4">
            Page not found. The requested resource does not exist.
          </p>
          <a
            href="/"
            className="inline-block px-4 py-1.5 border border-black bg-black text-white font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1.5px] hover:bg-gray-800 transition-colors"
          >
            Go Home
          </a>
        </div>
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
