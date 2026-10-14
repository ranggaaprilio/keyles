import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { EditClientForm } from "@/components/clients/EditClientForm";
import { useClient, useUpdateClient } from "@/hooks/useClients";
import type { Client } from "@/types/client";

vi.mock("@/hooks/useClients", () => ({
  useClient: vi.fn(),
  useUpdateClient: vi.fn(),
}));

const baseClient: Client = {
  client_id: "client-1",
  client_name: "Acme Portal",
  description: null,
  client_type: "public",
  redirect_uris: ["https://app.example.com/callback"],
  is_active: true,
  allow_self_signup: false,
  self_signup_default_role: "",
  created_at: "2026-10-01T00:00:00Z",
  updated_at: "2026-10-01T00:00:00Z",
};

describe("EditClientForm self sign-up settings", () => {
  const mutate = vi.fn();

  beforeEach(() => {
    mutate.mockReset();
    vi.mocked(useUpdateClient).mockReturnValue({ mutate, isPending: false } as unknown as ReturnType<typeof useUpdateClient>);
  });

  function renderForm(client: Client = baseClient) {
    vi.mocked(useClient).mockReturnValue({ data: client, isLoading: false } as unknown as ReturnType<typeof useClient>);
    return render(<EditClientForm clientId="client-1" onCancel={vi.fn()} onSuccess={vi.fn()} />);
  }

  it("reveals the default role field only when self sign-up is enabled", async () => {
    renderForm();
    expect(screen.queryByLabelText("Default role *")).not.toBeInTheDocument();

    fireEvent.click(await screen.findByLabelText("Allow self sign-up"));

    expect(screen.getByLabelText("Default role *")).toBeInTheDocument();
  });

  it("requires a default role when enabling self sign-up", async () => {
    renderForm();
    fireEvent.click(await screen.findByLabelText("Allow self sign-up"));
    fireEvent.click(screen.getByRole("button", { name: "Save Changes" }));

    expect(await screen.findByText("A default role is required when self sign-up is enabled")).toBeInTheDocument();
    expect(mutate).not.toHaveBeenCalled();
  });

  it("sends the self sign-up settings with the update", async () => {
    renderForm();
    fireEvent.click(await screen.findByLabelText("Allow self sign-up"));
    fireEvent.change(screen.getByLabelText("Default role *"), { target: { value: " member " } });
    fireEvent.click(screen.getByRole("button", { name: "Save Changes" }));

    await waitFor(() => expect(mutate).toHaveBeenCalled());
    expect(mutate.mock.calls[0][0]).toEqual({
      clientId: "client-1",
      data: expect.objectContaining({
        allow_self_signup: true,
        self_signup_default_role: "member",
      }),
    });
  });

  it("pre-fills existing settings", async () => {
    renderForm({ ...baseClient, allow_self_signup: true, self_signup_default_role: "viewer" });

    expect(await screen.findByLabelText("Allow self sign-up")).toBeChecked();
    expect(screen.getByLabelText("Default role *")).toHaveValue("viewer");
  });
});
