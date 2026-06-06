import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { IntegrationGuidePage } from "../pages/IntegrationGuidePage";

describe("IntegrationGuidePage", () => {
  it("renders the OAuth integration journey and endpoint reference", () => {
    render(<IntegrationGuidePage />);

    expect(
      screen.getByRole("heading", { name: "Connect your app to Keyles" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Authorization code flow with PKCE")).toBeInTheDocument();
    expect(screen.getByText("OIDC Discovery")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Production checklist" }),
    ).toBeInTheDocument();
    expect(
      screen.getByText("http://localhost:8080/oauth2/auth"),
    ).toBeInTheDocument();
  });

  it("copies a code sample", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText },
    });

    render(<IntegrationGuidePage />);
    fireEvent.click(screen.getByRole("button", { name: "Copy 1. PKCE utilities" }));

    await waitFor(() => expect(writeText).toHaveBeenCalledOnce());
    expect(
      screen.getByRole("button", { name: "Copy 1. PKCE utilities" }),
    ).toHaveTextContent("Copied");
  });
});
