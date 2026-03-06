/**
 * UserRoles component tests
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { UserRoles } from "@/components/users/UserRoles";
import type { RoleAssignment } from "@/types/user";

const mockUseUserRoles = vi.fn();
const mockRevokeMutateAsync = vi.fn();
const mockAssignMutateAsync = vi.fn();

vi.mock("@/hooks/useRoles", () => ({
  useUserRoles: (...args: unknown[]) => mockUseUserRoles(...args),
  useAssignRole: () => ({
    mutateAsync: mockAssignMutateAsync,
    isPending: false,
  }),
  useRevokeRole: () => ({
    mutateAsync: mockRevokeMutateAsync,
    isPending: false,
  }),
}));

vi.mock("@/hooks/useClients", () => ({
  useClients: () => ({ data: { clients: [] } }),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

const mockRoles: RoleAssignment[] = [
  {
    id: 1,
    user_id: "u1",
    client_id: "c1",
    client_name: "My App",
    tenant_id: "t1",
    role: "editor",
    is_active: true,
    granted_at: "2024-01-01T00:00:00Z",
    granted_by: "admin1",
  },
  {
    id: 2,
    user_id: "u1",
    client_id: "c1",
    client_name: "My App",
    tenant_id: "t1",
    role: "viewer",
    is_active: true,
    granted_at: "2024-01-02T00:00:00Z",
    granted_by: "admin1",
  },
  {
    id: 3,
    user_id: "u1",
    client_id: "c2",
    client_name: "Other App",
    tenant_id: "t1",
    role: "admin",
    is_active: true,
    granted_at: "2024-01-03T00:00:00Z",
    granted_by: "admin1",
  },
];

describe("UserRoles", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders roles grouped by client", () => {
    mockUseUserRoles.mockReturnValue({ data: mockRoles, isLoading: false });
    render(<UserRoles userId="u1" />, { wrapper });

    expect(screen.getByText("My App")).toBeInTheDocument();
    expect(screen.getByText("Other App")).toBeInTheDocument();
    expect(screen.getByText("editor")).toBeInTheDocument();
    expect(screen.getByText("viewer")).toBeInTheDocument();
    expect(screen.getByText("admin")).toBeInTheDocument();
  });

  it("shows empty state when no roles", () => {
    mockUseUserRoles.mockReturnValue({ data: [], isLoading: false });
    render(<UserRoles userId="u1" />, { wrapper });

    expect(
      screen.getByText(
        "No roles assigned. Assign a role to grant access to a client application.",
      ),
    ).toBeInTheDocument();
  });

  it("opens AssignRoleDialog on button click", async () => {
    mockUseUserRoles.mockReturnValue({ data: [], isLoading: false });
    const user = userEvent.setup();
    render(<UserRoles userId="u1" />, { wrapper });

    await user.click(screen.getByText("Assign Role"));

    await waitFor(() => {
      expect(screen.getByText("Client Application *")).toBeInTheDocument();
    });
  });

  it("shows loading skeletons", () => {
    mockUseUserRoles.mockReturnValue({ data: undefined, isLoading: true });
    const { container } = render(<UserRoles userId="u1" />, { wrapper });

    const skeletons = container.querySelectorAll(
      '[class*="skeleton"], .animate-pulse',
    );
    expect(skeletons.length).toBeGreaterThan(0);
  });
});
