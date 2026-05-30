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
import UserManagementPage from "./pages/UserManagementPage";
import { AcceptInvitationPage } from "./pages/AcceptInvitationPage";

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
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="text-center">
        <h1 className="text-6xl font-bold text-gray-900 mb-4">404</h1>
        <p className="text-xl text-gray-600 mb-8">Page not found</p>
        <a
          href="/"
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
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

            {/* Protected routes */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/clients"
              element={
                <ProtectedRoute>
                  <ClientManagementPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/clients/new"
              element={
                <ProtectedRoute>
                  <ClientManagementPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/clients/:clientId"
              element={
                <ProtectedRoute>
                  <ClientManagementPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard/clients/:clientId/edit"
              element={
                <ProtectedRoute>
                  <ClientManagementPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin/users"
              element={
                <ProtectedRoute>
                  <UserManagementPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin/users/:id"
              element={
                <ProtectedRoute>
                  <UserManagementPage />
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
