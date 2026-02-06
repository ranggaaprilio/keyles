/**
 * Unit tests for ClientManagement component
 * Tests client list rendering, CRUD operations, and form validation
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { ClientManagement } from "@/components/admin/ClientManagement";
import * as clientService from "@/services/clientService";
import type {
  Client,
  CreateClientResponse,
  RotateSecretResponse,
} from "@/types/client";

// Mock the client service
vi.mock("@/services/clientService", () => ({
  clientService: {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    rotateSecret: vi.fn(),
    get: vi.fn(),
  },
}));

// Mock navigator.clipboard
const mockClipboard = {
  writeText: vi.fn(),
};
Object.assign(navigator, { clipboard: mockClipboard });

describe("ClientManagement", () => {
  const mockClients: Client[] = [
    {
      client_id: "client-123",
      client_name: "Test App",
      redirect_uris: ["https://example.com/callback"],
      is_active: true,
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z",
    },
    {
      client_id: "client-456",
      client_name: "Another App",
      redirect_uris: [
        "https://app.example.com/oauth",
        "https://app.example.com/callback",
      ],
      is_active: false,
      created_at: "2024-01-02T00:00:00Z",
      updated_at: "2024-01-02T00:00:00Z",
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockClipboard.writeText.mockResolvedValue(undefined);
  });

  describe("Client List Rendering", () => {
    it("should display loading state initially", () => {
      vi.mocked(clientService.clientService.list).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(<ClientManagement />);

      expect(screen.getByText(/loading clients/i)).toBeInTheDocument();
    });

    it("should render client list after loading", async () => {
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText("Test App")).toBeInTheDocument();
        expect(screen.getByText("Another App")).toBeInTheDocument();
      });
    });

    it("should display client status correctly", async () => {
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText("Active")).toBeInTheDocument();
        expect(screen.getByText("Inactive")).toBeInTheDocument();
      });
    });

    it("should display empty state when no clients exist", async () => {
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: [],
        total: 0,
      });

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
        expect(
          screen.getByText(/create your first oauth client/i)
        ).toBeInTheDocument();
      });
    });

    it("should display error state when loading fails", async () => {
      vi.mocked(clientService.clientService.list).mockRejectedValue(
        new Error("Failed to fetch clients")
      );

      render(<ClientManagement />);

      await waitFor(() => {
        expect(
          screen.getByText(/failed to fetch clients/i)
        ).toBeInTheDocument();
      });
    });

    it("should truncate long client IDs", async () => {
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });

      render(<ClientManagement />);

      await waitFor(() => {
        // Client ID should be truncated to 12 chars + "..."
        expect(screen.getByText(/client-123\.\.\./i)).toBeInTheDocument();
      });
    });

    it("should show redirect URI count when more than 2", async () => {
      const baseClient = mockClients[0]!;
      const clientWithManyUris: Client = {
        client_id: baseClient.client_id,
        client_name: baseClient.client_name,
        is_active: baseClient.is_active,
        created_at: baseClient.created_at,
        updated_at: baseClient.updated_at,
        redirect_uris: [
          "https://example.com/1",
          "https://example.com/2",
          "https://example.com/3",
          "https://example.com/4",
        ],
      };

      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: [clientWithManyUris],
        total: 1,
      });

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText(/\+2 more/i)).toBeInTheDocument();
      });
    });
  });

  describe("Create Client", () => {
    it("should show create form when clicking New Client button", async () => {
      const user = userEvent.setup();
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText("Test App")).toBeInTheDocument();
      });

      await user.click(screen.getByRole("button", { name: /new client/i }));

      expect(screen.getByText(/create new client/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/client name/i)).toBeInTheDocument();
    });

    it("should display credentials modal after successful creation", async () => {
      const user = userEvent.setup();
      const createdClient: CreateClientResponse = {
        client_id: "new-client-id",
        client_secret: "secret-12345",
        client_name: "New App",
        redirect_uris: ["https://newapp.com/callback"],
        is_active: true,
        created_at: "2024-01-03T00:00:00Z",
      };

      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: [],
        total: 0,
      });
      vi.mocked(clientService.clientService.create).mockResolvedValue(
        createdClient
      );

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
      });

      await user.click(screen.getByRole("button", { name: /create client/i }));

      // Fill in the form
      await user.type(screen.getByLabelText(/client name/i), "New App");
      await user.type(
        screen.getByPlaceholderText(/https:\/\/example.com\/callback/i),
        "https://newapp.com/callback"
      );

      // Update mock for refetch
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: [
          { ...createdClient, client_secret: undefined } as unknown as Client,
        ],
        total: 1,
      });

      await user.click(screen.getByRole("button", { name: /create client/i }));

      // Check for credentials modal
      await waitFor(() => {
        expect(screen.getByText(/client created/i)).toBeInTheDocument();
        expect(screen.getByText("new-client-id")).toBeInTheDocument();
        expect(screen.getByText("secret-12345")).toBeInTheDocument();
      });
    });

    it("should return to list view when canceling create", async () => {
      const user = userEvent.setup();
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText("Test App")).toBeInTheDocument();
      });

      await user.click(screen.getByRole("button", { name: /new client/i }));
      expect(screen.getByText(/create new client/i)).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /cancel/i }));

      await waitFor(() => {
        expect(screen.getByText("Test App")).toBeInTheDocument();
      });
    });
  });

  describe("Delete Client", () => {
    it("should delete client when clicking delete button", async () => {
      const user = userEvent.setup();
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });
      vi.mocked(clientService.clientService.delete).mockResolvedValue(
        undefined
      );

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText("Test App")).toBeInTheDocument();
      });

      // Find the delete button for the first client
      const rows = screen.getAllByRole("row");
      const firstDataRow = rows[1]!; // Skip header row
      const deleteButton = within(firstDataRow).getByTitle(/delete client/i);

      // Update mock for after deletion
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: [mockClients[1]!],
        total: 1,
      });

      await user.click(deleteButton);

      await waitFor(() => {
        expect(clientService.clientService.delete).toHaveBeenCalledWith(
          "client-123"
        );
      });
    });

    it("should display error when delete fails", async () => {
      const user = userEvent.setup();
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });
      vi.mocked(clientService.clientService.delete).mockRejectedValue(
        new Error("Cannot delete client")
      );

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText("Test App")).toBeInTheDocument();
      });

      const rows = screen.getAllByRole("row");
      const firstDataRow = rows[1]!;
      const deleteButton = within(firstDataRow).getByTitle(/delete client/i);

      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText(/cannot delete client/i)).toBeInTheDocument();
      });
    });
  });

  describe("Rotate Secret", () => {
    it("should rotate secret and show new credentials", async () => {
      const user = userEvent.setup();
      const rotateResponse: RotateSecretResponse = {
        client_id: "client-123",
        client_secret: "new-secret-xyz",
        rotated_at: "2024-01-03T00:00:00Z",
      };

      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });
      vi.mocked(clientService.clientService.rotateSecret).mockResolvedValue(
        rotateResponse
      );

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText("Test App")).toBeInTheDocument();
      });

      const rows = screen.getAllByRole("row");
      const firstDataRow = rows[1]!;
      const rotateButton = within(firstDataRow).getByTitle(/rotate secret/i);

      await user.click(rotateButton);

      await waitFor(() => {
        expect(clientService.clientService.rotateSecret).toHaveBeenCalledWith(
          "client-123"
        );
        expect(screen.getByText(/secret rotated/i)).toBeInTheDocument();
        expect(screen.getByText("new-secret-xyz")).toBeInTheDocument();
      });
    });

    it("should display error when rotate fails", async () => {
      const user = userEvent.setup();
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });
      vi.mocked(clientService.clientService.rotateSecret).mockRejectedValue(
        new Error("Cannot rotate secret")
      );

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText("Test App")).toBeInTheDocument();
      });

      const rows = screen.getAllByRole("row");
      const firstDataRow = rows[1]!;
      const rotateButton = within(firstDataRow).getByTitle(/rotate secret/i);

      await user.click(rotateButton);

      await waitFor(() => {
        expect(screen.getByText(/cannot rotate secret/i)).toBeInTheDocument();
      });
    });
  });

  describe("Edit Client", () => {
    it("should show edit form when clicking edit button", async () => {
      const user = userEvent.setup();
      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: mockClients,
        total: mockClients.length,
      });

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText("Test App")).toBeInTheDocument();
      });

      const rows = screen.getAllByRole("row");
      const firstDataRow = rows[1]!;
      const editButton = within(firstDataRow).getByTitle(/edit client/i);

      await user.click(editButton);

      expect(screen.getByText(/edit client/i)).toBeInTheDocument();
      expect(screen.getByDisplayValue("Test App")).toBeInTheDocument();
    });
  });

  describe("Credentials Modal", () => {
    it("should display credentials modal with client ID and secret after creation", async () => {
      const user = userEvent.setup();
      const createdClient: CreateClientResponse = {
        client_id: "copy-test-id",
        client_secret: "copy-test-secret",
        client_name: "Copy Test",
        redirect_uris: ["https://test.com/callback"],
        is_active: true,
        created_at: "2024-01-03T00:00:00Z",
      };

      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: [],
        total: 0,
      });
      vi.mocked(clientService.clientService.create).mockResolvedValue(
        createdClient
      );

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
      });

      await user.click(screen.getByRole("button", { name: /create client/i }));

      await user.type(screen.getByLabelText(/client name/i), "Copy Test");
      await user.type(
        screen.getByPlaceholderText(/https:\/\/example.com\/callback/i),
        "https://test.com/callback"
      );

      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: [
          { ...createdClient, client_secret: undefined } as unknown as Client,
        ],
        total: 1,
      });

      await user.click(screen.getByRole("button", { name: /create client/i }));

      // Verify credentials are displayed in the modal
      await waitFor(() => {
        expect(screen.getByText("copy-test-id")).toBeInTheDocument();
        expect(screen.getByText("copy-test-secret")).toBeInTheDocument();
        expect(
          screen.getByText(/save these credentials securely/i)
        ).toBeInTheDocument();
      });
    });

    it("should close modal when clicking dismiss button", async () => {
      const user = userEvent.setup();
      const createdClient: CreateClientResponse = {
        client_id: "modal-test-id",
        client_secret: "modal-test-secret",
        client_name: "Modal Test",
        redirect_uris: ["https://test.com/callback"],
        is_active: true,
        created_at: "2024-01-03T00:00:00Z",
      };

      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: [],
        total: 0,
      });
      vi.mocked(clientService.clientService.create).mockResolvedValue(
        createdClient
      );

      render(<ClientManagement />);

      await waitFor(() => {
        expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
      });

      await user.click(screen.getByRole("button", { name: /create client/i }));

      await user.type(screen.getByLabelText(/client name/i), "Modal Test");
      await user.type(
        screen.getByPlaceholderText(/https:\/\/example.com\/callback/i),
        "https://test.com/callback"
      );

      vi.mocked(clientService.clientService.list).mockResolvedValue({
        clients: [
          { ...createdClient, client_secret: undefined } as unknown as Client,
        ],
        total: 1,
      });

      await user.click(screen.getByRole("button", { name: /create client/i }));

      await waitFor(() => {
        expect(screen.getByText(/client created/i)).toBeInTheDocument();
      });

      await user.click(
        screen.getByRole("button", { name: /i've saved the credentials/i })
      );

      await waitFor(() => {
        expect(screen.queryByText(/client created/i)).not.toBeInTheDocument();
      });
    });
  });
});

describe("ClientForm Validation", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should validate client name is required", async () => {
    const user = userEvent.setup();
    vi.mocked(clientService.clientService.list).mockResolvedValue({
      clients: [],
      total: 0,
    });

    render(<ClientManagement />);

    await waitFor(() => {
      expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /create client/i }));

    // Try to submit without filling the form
    await user.click(screen.getByRole("button", { name: /create client/i }));

    await waitFor(() => {
      expect(screen.getByText(/client name is required/i)).toBeInTheDocument();
    });
  });

  it("should validate client name minimum length", async () => {
    const user = userEvent.setup();
    vi.mocked(clientService.clientService.list).mockResolvedValue({
      clients: [],
      total: 0,
    });

    render(<ClientManagement />);

    await waitFor(() => {
      expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /create client/i }));

    await user.type(screen.getByLabelText(/client name/i), "AB");
    await user.type(
      screen.getByPlaceholderText(/https:\/\/example.com\/callback/i),
      "https://test.com/callback"
    );

    await user.click(screen.getByRole("button", { name: /create client/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/client name must be at least 3 characters/i)
      ).toBeInTheDocument();
    });
  });

  it("should validate redirect URI requires HTTPS", async () => {
    const user = userEvent.setup();
    vi.mocked(clientService.clientService.list).mockResolvedValue({
      clients: [],
      total: 0,
    });

    render(<ClientManagement />);

    await waitFor(() => {
      expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /create client/i }));

    await user.type(screen.getByLabelText(/client name/i), "Test App");
    await user.type(
      screen.getByPlaceholderText(/https:\/\/example.com\/callback/i),
      "http://insecure.com/callback"
    );

    await user.click(screen.getByRole("button", { name: /create client/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/redirect uri must use https/i)
      ).toBeInTheDocument();
    });
  });

  it("should allow localhost for development", async () => {
    const user = userEvent.setup();
    const createdClient: CreateClientResponse = {
      client_id: "local-client",
      client_secret: "local-secret",
      client_name: "Local App",
      redirect_uris: ["http://localhost:3000/callback"],
      is_active: true,
      created_at: "2024-01-03T00:00:00Z",
    };

    vi.mocked(clientService.clientService.list).mockResolvedValue({
      clients: [],
      total: 0,
    });
    vi.mocked(clientService.clientService.create).mockResolvedValue(
      createdClient
    );

    render(<ClientManagement />);

    await waitFor(() => {
      expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /create client/i }));

    await user.type(screen.getByLabelText(/client name/i), "Local App");
    await user.type(
      screen.getByPlaceholderText(/https:\/\/example.com\/callback/i),
      "http://localhost:3000/callback"
    );

    vi.mocked(clientService.clientService.list).mockResolvedValue({
      clients: [
        { ...createdClient, client_secret: undefined } as unknown as Client,
      ],
      total: 1,
    });

    await user.click(screen.getByRole("button", { name: /create client/i }));

    // Should not show validation error - should create successfully
    await waitFor(() => {
      expect(clientService.clientService.create).toHaveBeenCalled();
    });
  });

  it("should validate redirect URI cannot contain fragments", async () => {
    const user = userEvent.setup();
    vi.mocked(clientService.clientService.list).mockResolvedValue({
      clients: [],
      total: 0,
    });

    render(<ClientManagement />);

    await waitFor(() => {
      expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /create client/i }));

    await user.type(screen.getByLabelText(/client name/i), "Fragment App");
    await user.type(
      screen.getByPlaceholderText(/https:\/\/example.com\/callback/i),
      "https://example.com/callback#fragment"
    );

    await user.click(screen.getByRole("button", { name: /create client/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/redirect uri cannot contain fragments/i)
      ).toBeInTheDocument();
    });
  });

  it("should require at least one redirect URI", async () => {
    const user = userEvent.setup();
    vi.mocked(clientService.clientService.list).mockResolvedValue({
      clients: [],
      total: 0,
    });

    render(<ClientManagement />);

    await waitFor(() => {
      expect(screen.getByText(/no clients yet/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /create client/i }));

    await user.type(screen.getByLabelText(/client name/i), "No URI App");
    // Leave redirect URI empty

    await user.click(screen.getByRole("button", { name: /create client/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/at least one redirect uri is required/i)
      ).toBeInTheDocument();
    });
  });
});
