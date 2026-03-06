/**
 * UserDetail — profile header, action buttons, and tabbed content
 */

import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { Skeleton } from "../ui/skeleton";
import { UserStatusBadge } from "./UserStatusBadge";
import { UserRoles } from "./UserRoles";
import { UserSessions } from "./UserSessions";
import { UserActivity } from "./UserActivity";
import { EnableDisableDialog } from "./EnableDisableDialog";
import { DeleteUserDialog } from "./DeleteUserDialog";
import {
  useUser,
  useUpdateUser,
  useResendInvitation,
} from "../../hooks/useUsers";
import {
  ArrowLeft,
  Pencil,
  ShieldOff,
  ShieldCheck,
  Trash2,
  Send,
} from "lucide-react";
import { Input } from "../ui/input";

interface UserDetailProps {
  userId: string;
}

export function UserDetail({ userId }: UserDetailProps) {
  const navigate = useNavigate();
  const { data: user, isLoading, refetch } = useUser(userId);
  const updateUser = useUpdateUser();
  const resendInvitation = useResendInvitation();

  const [editing, setEditing] = useState(false);
  const [displayName, setDisplayName] = useState("");
  const [statusDialogOpen, setStatusDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-64 w-full" />
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-gray-900">
            User not found
          </h2>
          <Button
            variant="outline"
            className="mt-4"
            onClick={() => navigate("/dashboard/users")}
          >
            Back to Users
          </Button>
        </div>
      </div>
    );
  }

  const initials = (user.display_name || user.email)
    .split(" ")
    .map((w) => w[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();

  const startEditing = () => {
    setDisplayName(user.display_name || "");
    setEditing(true);
  };

  const saveDisplayName = async () => {
    await updateUser.mutateAsync({ id: userId, displayName });
    setEditing(false);
    refetch();
  };

  const handleResend = async () => {
    await resendInvitation.mutateAsync(userId);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Back */}
        <Button
          variant="ghost"
          size="sm"
          className="mb-4"
          onClick={() => navigate("/dashboard/users")}
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Users
        </Button>

        {/* Profile header */}
        <div className="bg-white rounded-lg border shadow-sm p-6 mb-6">
          <div className="flex items-start gap-4">
            <div className="h-14 w-14 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center text-lg font-semibold">
              {initials}
            </div>
            <div className="flex-1 min-w-0">
              {editing ? (
                <div className="flex items-center gap-2">
                  <Input
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    className="max-w-xs"
                    autoFocus
                  />
                  <Button
                    size="sm"
                    onClick={saveDisplayName}
                    disabled={updateUser.isPending}
                  >
                    Save
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => setEditing(false)}
                  >
                    Cancel
                  </Button>
                </div>
              ) : (
                <h2 className="text-xl font-semibold text-gray-900">
                  {user.display_name || user.email.split("@")[0]}
                </h2>
              )}
              <p className="text-sm text-gray-500 mt-1">{user.email}</p>
              <div className="flex items-center gap-3 mt-2">
                <UserStatusBadge status={user.status} />
                {user.last_login_at && (
                  <span className="text-xs text-gray-400">
                    Last login: {new Date(user.last_login_at).toLocaleString()}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Action buttons */}
          <div className="flex flex-wrap gap-2 mt-4 border-t pt-4">
            {!editing && (
              <Button variant="outline" size="sm" onClick={startEditing}>
                <Pencil className="h-4 w-4 mr-1" />
                Edit Name
              </Button>
            )}
            {user.status === "pending" && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleResend}
                disabled={resendInvitation.isPending}
              >
                <Send className="h-4 w-4 mr-1" />
                {resendInvitation.isPending
                  ? "Sending..."
                  : "Resend Invitation"}
              </Button>
            )}
            {(user.status === "active" || user.status === "disabled") && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setStatusDialogOpen(true)}
              >
                {user.status === "active" ? (
                  <>
                    <ShieldOff className="h-4 w-4 mr-1" />
                    Disable
                  </>
                ) : (
                  <>
                    <ShieldCheck className="h-4 w-4 mr-1" />
                    Enable
                  </>
                )}
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              className="text-red-600 hover:text-red-700"
              onClick={() => setDeleteDialogOpen(true)}
            >
              <Trash2 className="h-4 w-4 mr-1" />
              Delete
            </Button>
          </div>
        </div>

        {/* Tabbed content */}
        <Tabs defaultValue="roles">
          <TabsList>
            <TabsTrigger value="roles">Roles</TabsTrigger>
            <TabsTrigger value="sessions">Sessions</TabsTrigger>
            <TabsTrigger value="activity">Activity</TabsTrigger>
          </TabsList>
          <TabsContent value="roles" className="mt-4">
            <UserRoles userId={userId} />
          </TabsContent>
          <TabsContent value="sessions" className="mt-4">
            <UserSessions userId={userId} />
          </TabsContent>
          <TabsContent value="activity" className="mt-4">
            <UserActivity userId={userId} />
          </TabsContent>
        </Tabs>

        {/* Dialogs */}
        <EnableDisableDialog
          open={statusDialogOpen}
          onOpenChange={setStatusDialogOpen}
          userId={userId}
          email={user.email}
          currentStatus={user.status}
          onSuccess={() => refetch()}
        />
        <DeleteUserDialog
          open={deleteDialogOpen}
          onOpenChange={setDeleteDialogOpen}
          userId={userId}
          email={user.email}
          onDeleted={() => navigate("/dashboard/users")}
        />
      </div>
    </div>
  );
}
