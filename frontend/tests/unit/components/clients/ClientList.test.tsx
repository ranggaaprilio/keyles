/**
 * Unit tests for ClientList component
 * T051 - Pagination, search, empty state, loading
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ClientList } from "@/components/clients/ClientList";
import type { Client, PaginatedResponse } from "@/types/client";

// Mock the useClients hook
const mockUseClients = vi.fn();
vi.mock("@/hooks/useClients", () => ({
  useClients: (...args: unknown[]) => mockUseClients(...args),
}));

const mockClients: Client[] = [
  {
    client_id: "client-1",
    client_name: "App One",
    description: "First app",
    client_type: "confidential",
    redirect_uris: ["https://one.example.com/callback"],
    is_active: true,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  },
  {
    client_id: "client-2",
    client_name: "App Two",
    description: null,
    client_type: "public",
    redirect_uris: ["https://two.example.com/callback"],
    is_active: true,
    created_at: "2024-01-02T00:00:00Z",
    updated_at: "2024-01-02T00:00:00Z",
  },
];

const paginatedResponse: PaginatedResponse<Client> = {
  clients: mockClients,
  total: 2,
  page: 1,
  page_size: 10,
  total_pages: 1,
};

describe("ClientList", () => {
  const mockOnSelectClient = vi.fn();
  const mockOnCreateNew = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseClients.mockReturnValue({
      data: paginatedResponse,
      isLoading: false,
      isError: false,
    });
  });

  it("renders client cards from data", () => {
    render(
      <ClientList
        onSelectClient={mockOnSelectClient}
        onCreateNew={mockOnCreateNew}
      />,
    );
    expect(screen.getByText("App One")).toBeInTheDocument();
    expect(screen.getByText("App Two")).toBeInTheDocument();
  });

  it("shows loading skeleton when loading", () => {
    mockUseClients.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    });
    render(
      <ClientList
        onSelectClient={mockOnSelectClient}
        onCreateNew={mockOnCreateNew}
      />,
    );
    // Should not show client names while loading
    expect(screen.queryByText("App One")).not.toBeInTheDocument();
  });

  it("shows empty state when no clients", () => {
    mockUseClients.mockReturnValue({
      data: { clients: [], total: 0, page: 1, page_size: 10, total_pages: 0 },
      isLoading: false,
      isError: false,
    });
    render(
      <ClientList
        onSelectClient={mockOnSelectClient}
        onCreateNew={mockOnCreateNew}
      />,
    );
    expect(
      screen.getByText(/no client applications found/i),
    ).toBeInTheDocument();
  });

  it("renders search input", () => {
    render(
      <ClientList
        onSelectClient={mockOnSelectClient}
        onCreateNew={mockOnCreateNew}
      />,
    );
    expect(screen.getByPlaceholderText(/search/i)).toBeInTheDocument();
  });

  it("renders Register Client button", () => {
    render(
      <ClientList
        onSelectClient={mockOnSelectClient}
        onCreateNew={mockOnCreateNew}
      />,
    );
    const registerBtn = screen.getByRole("button", { name: /register|new/i });
    expect(registerBtn).toBeInTheDocument();
  });

  it("calls onCreateNew when Register button is clicked", async () => {
    const user = userEvent.setup();
    render(
      <ClientList
        onSelectClient={mockOnSelectClient}
        onCreateNew={mockOnCreateNew}
      />,
    );
    const registerBtn = screen.getByRole("button", { name: /register|new/i });
    await user.click(registerBtn);
    expect(mockOnCreateNew).toHaveBeenCalled();
  });

  it("renders pagination controls for multi-page results", () => {
    mockUseClients.mockReturnValue({
      data: {
        clients: mockClients,
        total: 25,
        page: 1,
        page_size: 10,
        total_pages: 3,
      },
      isLoading: false,
      isError: false,
    });
    render(
      <ClientList
        onSelectClient={mockOnSelectClient}
        onCreateNew={mockOnCreateNew}
      />,
    );
    expect(screen.getByText(/page 1 of 3/i)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /previous/i }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /next/i })).toBeInTheDocument();
  });
});
