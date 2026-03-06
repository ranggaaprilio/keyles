/**
 * Accept Invitation Page (public, no auth)
 *
 * Validates an invitation token from the URL, then shows a password-creation form.
 * If the token is expired or already used, shows an error state.
 */

import { useParams, useNavigate } from "react-router-dom";
import {
  useValidateInvitation,
  useAcceptInvitation,
} from "../hooks/useInvitation";
import { AcceptInvitationForm } from "../components/users/AcceptInvitationForm";
import { InvitationExpired } from "../components/users/InvitationExpired";
import axios from "axios";

export function AcceptInvitationPage() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const { data, isLoading, error } = useValidateInvitation(token ?? "");
  const acceptMutation = useAcceptInvitation(token ?? "");

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-pulse text-gray-500">
          Validating invitation...
        </div>
      </div>
    );
  }

  // 410 Gone or any error → show expired/invalid state
  if (error) {
    const status = axios.isAxiosError(error)
      ? error.response?.status
      : undefined;
    if (status === 410 || status === 404) {
      return <InvitationExpired />;
    }
    return <InvitationExpired />;
  }

  if (!data) {
    return <InvitationExpired />;
  }

  const handleAccept = async (password: string) => {
    await acceptMutation.mutateAsync({ password });
    navigate("/login", {
      state: { message: "Account created successfully. Please log in." },
    });
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-lg shadow-sm border p-8">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            Accept Invitation
          </h1>
          <p className="text-gray-600 mb-6">
            Set a password to activate your account.
          </p>
          <AcceptInvitationForm
            email={data.email}
            displayName={data.display_name}
            onSubmit={handleAccept}
            isSubmitting={acceptMutation.isPending}
            error={
              acceptMutation.error && axios.isAxiosError(acceptMutation.error)
                ? (acceptMutation.error.response?.data?.error ??
                  "Something went wrong.")
                : acceptMutation.error
                  ? "Something went wrong."
                  : undefined
            }
          />
        </div>
      </div>
    </div>
  );
}
