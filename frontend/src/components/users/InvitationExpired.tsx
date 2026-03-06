/**
 * InvitationExpired — error state for expired or used invitation
 */

import { AlertCircle } from "lucide-react";

export function InvitationExpired() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="w-full max-w-md text-center">
        <div className="bg-white rounded-lg shadow-sm border p-8">
          <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-gray-900 mb-2">
            Invitation Expired
          </h2>
          <p className="text-gray-600 mb-6">
            This invitation link has expired or has already been used.
          </p>
          <p className="text-sm text-gray-500">
            Please contact your administrator to request a new invitation.
          </p>
        </div>
      </div>
    </div>
  );
}
