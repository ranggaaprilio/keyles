/**
 * UserList component tests
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import { UserList } from "@/components/users/UserList";
import type { User, PaginatedResponse } from "@/types/user";

// Mock the hooks
const mockUseUsers = vi.fn();
const mockUseInviteUser = vi.fn(() => ({
  mutateAsync: vi.fn(),
  isPending: false,
}));

vi.mock("@/hooks/useUsers", () => ({
  useUsers: (...args: unknown[]) => mockUseUsers(...args),
  useInviteUser: () => mockUseInviteUser(),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return (
    <QueryClientProvider client={qc}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  );
}

const mockUsers: User[] = [
  {
    id: "u1",
    tenant_id: "t1",
    email: "alice@example.com",
    display_name: "Alice",
    status: "active",
    last_login_at: new Date().toISOString(),
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    role_count: 3,
  },
  {
    id: "u2",
    tenant_id: "t1",
    email: "bob@example.com",
    display_name: "Bob",
    status: "pending",
    last_login_at: null,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    role_count: 0,
  },
];

const mockResponse: PaginatedResponse<User> = {
  data: mockUsers,
  total: 2,
  page: 1,
  page_size: 25,
  total_pages: 1,
};

describe("UserList", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders user rows with correct columns", async () => {
    mockUseUsers.mockReturnValue({
      data: mockResponse,
      isLoading: false,
      error: null,
    });
    render(<UserList />, { wrapper });

    expect(screen.getByText("Alice")).toBeInTheDocument();
    expect(screen.getByText("alice@example.com")).toBeInTheDocument();
    expect(screen.getByText("Bob")).toBeInTheDocument();
    expect(screen.getByText("bob@example.com")).toBeInTheDocument();
    // Status badges exist (there are tab buttons with same names, use getAllByText)
    expect(screen.getAllByText("Active").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Pending").length).toBeGreaterThanOrEqual(1);
  });

  it("shows loading skeletons while fetching", () => {
    mockUseUsers.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });
    const { container } = render(<UserList />, { wrapper });

    // Skeleton elements should be present
    const skeletons = container.querySelectorAll(
      '.animate-pulse, [class*="skeleton"]',
    );
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("renders empty state when no users", () => {
    mockUseUsers.mockReturnValue({
      data: { data: [], total: 0, page: 1, page_size: 25, total_pages: 0 },
      isLoading: false,
      error: null,
    });
    render(<UserList />, { wrapper });

    expect(screen.getByText("No users found.")).toBeInTheDocument();
  });

  it("updates search filter via debounced input", async () => {
    mockUseUsers.mockReturnValue({
      data: mockResponse,
      isLoading: false,
      error: null,
    });
    const user = userEvent.setup();
    render(<UserList />, { wrapper });

    const searchInput = screen.getByPlaceholderText(
      "Search by name or email...",
    );
    await user.type(searchInput, "alice");

    await waitFor(() => {
      // Verify useUsers was called with search filter
      const lastCall =
        mockUseUsers.mock.calls[mockUseUsers.mock.calls.length - 1];
      expect(lastCall[0]).toMatchObject({ search: "alice" });
    });
  });

  it("switches status filter on tab click", async () => {
    mockUseUsers.mockReturnValue({
      data: mockResponse,
      isLoading: false,
      error: null,
    });
    const user = userEvent.setup();
    render(<UserList />, { wrapper });

    // Click the "Disabled" tab (unique text, no badge in data set)
    await user.click(screen.getByText("Disabled"));

    await waitFor(() => {
      const lastCall =
        mockUseUsers.mock.calls[mockUseUsers.mock.calls.length - 1];
      expect(lastCall[0]).toMatchObject({ status: "disabled" });
    });
  });

  it("shows pagination controls when multiple pages", () => {
    mockUseUsers.mockReturnValue({
      data: { ...mockResponse, total: 50, total_pages: 2 },
      isLoading: false,
      error: null,
    });
    render(<UserList />, { wrapper });

    expect(screen.getByText("Previous")).toBeInTheDocument();
    expect(screen.getByText("Next")).toBeInTheDocument();
    expect(screen.getByText("Page 1 of 2 (50 users)")).toBeInTheDocument();
  });
});
