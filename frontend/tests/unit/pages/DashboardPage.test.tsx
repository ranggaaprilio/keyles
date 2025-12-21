/**
 * DashboardPage Unit Tests
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { BrowserRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { DashboardPage } from "../../../src/pages/DashboardPage";
import * as authApi from "../../../src/services/api/auth";

// Mock auth service
vi.mock("../../../src/services/api/auth", () => ({
  getDashboard: vi.fn(),
  isAuthenticated: vi.fn(),
  logout: vi.fn(),
  getStoredUser: vi.fn(),
  login: vi.fn(),
  storeAuthData: vi.fn(),
}));

// Mock router navigation
const mockNavigate = vi.fn();
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Create a fresh query client for each test
function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
}

function renderDashboardPage() {
  const testQueryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={testQueryClient}>
      <BrowserRouter>
        <DashboardPage />
      </BrowserRouter>
    </QueryClientProvider>
  );
}

const mockDashboardData = {
  tenant: {
    id: "tenant-123",
    organization_name: "Test Organization",
    status: "active",
    created_at: "2024-01-15T10:00:00Z",
    verified_at: "2024-01-15T11:00:00Z",
  },
  user: {
    id: "user-123",
    full_name: "Admin User",
    email: "admin@example.com",
    role: "admin",
  },
};

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (authApi.isAuthenticated as any).mockReturnValue(true);
  });

  it("redirects to login if not authenticated", () => {
    (authApi.isAuthenticated as any).mockReturnValue(false);

    renderDashboardPage();

    expect(mockNavigate).toHaveBeenCalledWith("/login", { replace: true });
  });

  it("shows loading state while fetching data", () => {
    (authApi.getDashboard as any).mockImplementation(
      () => new Promise(() => {})
    ); // Never resolves

    renderDashboardPage();

    expect(screen.getByText(/Loading dashboard/i)).toBeInTheDocument();
  });

  it("displays dashboard data after successful load", async () => {
    // Setup mocks BEFORE render
    const dashboardData = { ...mockDashboardData };
    (authApi.getDashboard as any).mockResolvedValue(dashboardData);
    (authApi.isAuthenticated as any).mockReturnValue(true);

    renderDashboardPage();

    // Wait for the loading to complete
    await waitFor(
      () => {
        expect(
          screen.queryByText(/Loading dashboard/i)
        ).not.toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    // Now check the data - use getAllByText since org name may appear multiple times
    expect(screen.getAllByText("Test Organization").length).toBeGreaterThan(0);
    expect(screen.getByText(/Welcome back, Admin User/i)).toBeInTheDocument();
    expect(screen.getByText("admin@example.com")).toBeInTheDocument();
  });

  it("shows error message on load failure", async () => {
    (authApi.getDashboard as any).mockRejectedValue(
      new Error("Failed to load dashboard")
    );
    (authApi.isAuthenticated as any).mockReturnValue(true);

    renderDashboardPage();

    await waitFor(
      () => {
        expect(
          screen.queryByText(/Loading dashboard/i)
        ).not.toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    // The error heading appears - use getAllByText in case it appears in multiple places
    expect(
      screen.getAllByText(/Failed to Load Dashboard/i).length
    ).toBeGreaterThan(0);
  });

  it("handles logout action", async () => {
    const user = userEvent.setup();
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    (authApi.isAuthenticated as any).mockReturnValue(true);

    renderDashboardPage();

    await waitFor(
      () => {
        expect(
          screen.queryByText(/Loading dashboard/i)
        ).not.toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    const logoutButton = screen.getByRole("button", { name: /Logout/i });
    await user.click(logoutButton);

    expect(authApi.logout).toHaveBeenCalled();
  });

  it("displays tenant status badge", async () => {
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    (authApi.isAuthenticated as any).mockReturnValue(true);

    renderDashboardPage();

    await waitFor(
      () => {
        expect(
          screen.queryByText(/Loading dashboard/i)
        ).not.toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    expect(screen.getByText(/active/i)).toBeInTheDocument();
  });

  it("displays user role", async () => {
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    (authApi.isAuthenticated as any).mockReturnValue(true);

    renderDashboardPage();

    await waitFor(
      () => {
        expect(
          screen.queryByText(/Loading dashboard/i)
        ).not.toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    // User role appears in a badge - use getAllByText since it may appear multiple times
    expect(screen.getAllByText(/ADMIN/i).length).toBeGreaterThan(0);
  });

  it("handles tenant with null verified_at", async () => {
    const dataWithNullVerified = {
      ...mockDashboardData,
      tenant: {
        ...mockDashboardData.tenant,
        verified_at: null,
      },
    };
    (authApi.getDashboard as any).mockResolvedValue(dataWithNullVerified);
    (authApi.isAuthenticated as any).mockReturnValue(true);

    renderDashboardPage();

    await waitFor(
      () => {
        expect(
          screen.queryByText(/Loading dashboard/i)
        ).not.toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    expect(screen.getAllByText("Test Organization").length).toBeGreaterThan(0);
  });

  it("formats dates correctly", async () => {
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    (authApi.isAuthenticated as any).mockReturnValue(true);

    renderDashboardPage();

    await waitFor(
      () => {
        expect(
          screen.queryByText(/Loading dashboard/i)
        ).not.toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    // Should display formatted date strings - the date contains "2024" or a month name
    const dateText = screen.getAllByText(/\d{4}|January|February/i);
    expect(dateText.length).toBeGreaterThan(0);
  });

  it("shows quick actions menu", async () => {
    (authApi.getDashboard as any).mockResolvedValue(mockDashboardData);
    (authApi.isAuthenticated as any).mockReturnValue(true);

    renderDashboardPage();

    // Wait for the loading to complete
    await waitFor(
      () => {
        expect(
          screen.queryByText(/Loading dashboard/i)
        ).not.toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    // Check quick actions section exists
    expect(screen.getByText(/Quick Actions/i)).toBeInTheDocument();
    // Check for buttons in the quick actions section
    expect(
      screen.getByRole("button", { name: /Manage Users/i })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /^Settings$/i })
    ).toBeInTheDocument();
  });
});
