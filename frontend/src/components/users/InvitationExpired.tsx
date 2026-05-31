/**
 * InvitationExpired — Dell 1996 retro style
 */

import { AlertCircle } from "lucide-react";

export function InvitationExpired() {
  return (
    <div className="min-h-screen bg-white flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        {/* Section eyebrow — salmon */}
        <div className="bg-[#d77a7a] px-4 py-4">
          <h2 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[28px] font-black uppercase leading-[1.0] text-black">
            INVITATION EXPIRED
          </h2>
        </div>

        <div className="border-x border-b border-black bg-white p-4">
          <AlertCircle className="h-8 w-8 text-red-700 mb-3" />
          <p className="font-['Times_New_Roman',Times,serif] text-sm text-black mb-3">
            This invitation link has expired or has already been used.
          </p>
          <p className="font-['Times_New_Roman',Times,serif] text-[11px] text-gray-600">
            Please contact your administrator to request a new invitation.
          </p>
        </div>
      </div>
    </div>
  );
}
