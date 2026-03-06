/**
 * User Management Page
 *
 * Main page for managing tenant users.
 * Renders UserList at the list view, and user detail when a userId is selected.
 */

import { useParams } from "react-router-dom";
import { UserList } from "../components/users/UserList";
import { UserDetail } from "../components/users/UserDetail";

export function UserManagementPage() {
  const { userId } = useParams<{ userId: string }>();

  if (userId) {
    return <UserDetail userId={userId} />;
  }

  return <UserList />;
}
