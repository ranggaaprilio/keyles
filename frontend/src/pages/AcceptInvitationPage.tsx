/**
 * Accept Invitation Page
 */

import { useParams, useNavigate } from "react-router-dom";
import {
  useValidateInvitation,
  useAcceptInvitation,
} from "../hooks/useInvitation";
import { AcceptInvitationForm } from "../components/users/AcceptInvitationForm";
import { InvitationExpired } from "../components/users/InvitationExpired";
import { AuthShell } from "../components/layout/AuthShell";
import axios from "axios";

export function AcceptInvitationPage() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const { data, isLoading, error } = useValidateInvitation(token ?? "");
  const acceptMutation = useAcceptInvitation(token ?? "");

  if (isLoading) {
    return (
      <div className="min-h-screen bg-muted flex items-center justify-center">
        <p className="animate-pulse text-sm text-muted-foreground">
          Validating invitation...
        </p>
      </div>
    );
  }

  if (error) {
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
    <AuthShell
      title="Accept invitation"
      description="Set a password to finish creating your account."
    >
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
    </AuthShell>
  );
}
