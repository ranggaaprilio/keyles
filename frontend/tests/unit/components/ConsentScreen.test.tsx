/**
 * ConsentScreen Component Tests
 * Tests for OAuth consent screen functionality
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ConsentScreen } from "../../../src/components/auth/ConsentScreen";
import type { ClientInfo } from "../../../src/types/oauth";

describe("ConsentScreen", () => {
  const mockClient: ClientInfo = {
    client_id: "test-client-123",
    client_name: "Test Application",
    logo_uri: "https://example.com/logo.png",
    policy_uri: "https://example.com/privacy",
    tos_uri: "https://example.com/terms",
  };

  const mockScopes = ["openid", "profile", "email"];
  const mockUser = "test@example.com";
  let mockOnApprove: ReturnType<typeof vi.fn>;
  let mockOnDeny: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockOnApprove = vi.fn();
    mockOnDeny = vi.fn();
  });

  it("renders client name and logo", () => {
    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    expect(screen.getByText("Test Application")).toBeInTheDocument();
    expect(screen.getByAltText("Test Application logo")).toBeInTheDocument();
  });

  it("renders user email", () => {
    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    expect(screen.getByText("test@example.com")).toBeInTheDocument();
  });

  it("displays all requested scopes", () => {
    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    expect(screen.getByText(/OpenID/)).toBeInTheDocument();
    expect(screen.getByText(/Profile/)).toBeInTheDocument();
    expect(screen.getByText(/Email/)).toBeInTheDocument();
  });

  it("displays scope descriptions", () => {
    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    expect(screen.getByText("Verify your identity")).toBeInTheDocument();
    expect(
      screen.getByText("Access your name and profile picture")
    ).toBeInTheDocument();
    expect(screen.getByText("Access your email address")).toBeInTheDocument();
  });

  it("calls onApprove when Allow button is clicked", async () => {
    mockOnApprove.mockResolvedValue(undefined);

    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    fireEvent.click(screen.getByText("Allow"));

    await waitFor(() => {
      expect(mockOnApprove).toHaveBeenCalledTimes(1);
    });
  });

  it("calls onDeny when Deny button is clicked", async () => {
    mockOnDeny.mockResolvedValue(undefined);

    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    fireEvent.click(screen.getByText("Deny"));

    await waitFor(() => {
      expect(mockOnDeny).toHaveBeenCalledTimes(1);
    });
  });

  it("disables buttons while loading", () => {
    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
        isLoading={true}
      />
    );

    expect(screen.getByText("Allow")).toBeDisabled();
    expect(screen.getByText("Deny")).toBeDisabled();
  });

  it("shows loading state during approval", async () => {
    // Make onApprove take some time
    mockOnApprove.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    fireEvent.click(screen.getByText("Allow"));

    await waitFor(() => {
      expect(screen.getByText("Allowing...")).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.queryByText("Allowing...")).not.toBeInTheDocument();
    });
  });

  it("shows loading state during denial", async () => {
    mockOnDeny.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    fireEvent.click(screen.getByText("Deny"));

    await waitFor(() => {
      expect(screen.getByText("Denying...")).toBeInTheDocument();
    });
  });

  it("displays policy and terms links when provided", () => {
    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    const termsLink = screen.getByText("Terms of Service");
    const privacyLink = screen.getByText("Privacy Policy");

    expect(termsLink).toHaveAttribute("href", "https://example.com/terms");
    expect(privacyLink).toHaveAttribute("href", "https://example.com/privacy");
  });

  it("handles client without logo", () => {
    const clientWithoutLogo: ClientInfo = {
      client_id: "test-client-123",
      client_name: "Test Application",
    };

    render(
      <ConsentScreen
        client={clientWithoutLogo}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    // Should not crash and should show fallback
    expect(screen.getByText("Test Application")).toBeInTheDocument();
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });

  it("handles unknown scopes", () => {
    const unknownScopes = ["openid", "custom_scope"];

    render(
      <ConsentScreen
        client={mockClient}
        scopes={unknownScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    expect(screen.getByText(/OpenID/)).toBeInTheDocument();
    expect(screen.getAllByText(/custom_scope/)).toHaveLength(2);
    expect(screen.getByText("Access to custom_scope")).toBeInTheDocument();
  });

  it("displays security warning", () => {
    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    expect(
      screen.getByText(/Only grant access to applications you trust/i)
    ).toBeInTheDocument();
  });

  it("displays Keyles branding", () => {
    render(
      <ConsentScreen
        client={mockClient}
        scopes={mockScopes}
        user={mockUser}
        onApprove={mockOnApprove}
        onDeny={mockOnDeny}
      />
    );

    expect(screen.getByText("Keyles SSO")).toBeInTheDocument();
  });
});
