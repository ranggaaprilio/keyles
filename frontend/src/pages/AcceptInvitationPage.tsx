/**
 * Accept Invitation Page — Dell 1996 retro style
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
      <div className="min-h-screen bg-white flex items-center justify-center">
        <div className="animate-pulse font-['Times_New_Roman',Times,serif] text-sm text-gray-600">
          Validating invitation...
        </div>
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
    <div className="min-h-screen bg-white flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        {/* Section eyebrow — peach */}
        <div className="bg-[#e6915d] px-4 py-4 mb-0">
          <h1 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[28px] font-black uppercase leading-[1.0] text-black">
            ACCEPT INVITATION
          </h1>
        </div>

        {/* Form card */}
        <div className="border-x border-b border-black">
          <div className="border-b border-black bg-white px-3 py-1.5">
            <h3 className="font-[Helvetica,Arial,system-ui,sans-serif] text-sm font-bold uppercase text-black">
              SET YOUR PASSWORD
            </h3>
          </div>
          <div className="bg-[#e6915d] px-4 py-4">
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
    </div>
  );
}
