/**
 * InviteUserDialog component tests
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { InviteUserDialog } from "@/components/users/InviteUserDialog";

const mockMutateAsync = vi.fn();
vi.mock("@/hooks/useUsers", () => ({
  useInviteUser: () => ({
    mutateAsync: mockMutateAsync,
    isPending: false,
  }),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe("InviteUserDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders email and display name fields", () => {
    render(<InviteUserDialog open={true} onOpenChange={vi.fn()} />, {
      wrapper,
    });

    expect(screen.getByLabelText("Email *")).toBeInTheDocument();
    expect(screen.getByLabelText("Display Name")).toBeInTheDocument();
  });

  it("submits with valid email", async () => {
    mockMutateAsync.mockResolvedValue({});
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(<InviteUserDialog open={true} onOpenChange={onOpenChange} />, {
      wrapper,
    });

    await user.type(screen.getByLabelText("Email *"), "test@example.com");
    await user.click(screen.getByRole("button", { name: "Send Invitation" }));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ email: "test@example.com" }),
      );
    });
  });

  it("shows validation error for invalid email", async () => {
    const user = userEvent.setup();
    render(<InviteUserDialog open={true} onOpenChange={vi.fn()} />, {
      wrapper,
    });

    // Submit without entering any email to trigger zod validation
    await user.click(screen.getByRole("button", { name: "Send Invitation" }));

    await waitFor(() => {
      expect(
        screen.getByText("Enter a valid email address"),
      ).toBeInTheDocument();
    });
    expect(mockMutateAsync).not.toHaveBeenCalled();
  });

  it("shows error message on API failure", async () => {
    mockMutateAsync.mockRejectedValue(new Error("Network error"));
    const user = userEvent.setup();
    render(<InviteUserDialog open={true} onOpenChange={vi.fn()} />, {
      wrapper,
    });

    await user.type(screen.getByLabelText("Email *"), "test@example.com");
    await user.click(screen.getByRole("button", { name: "Send Invitation" }));

    await waitFor(() => {
      expect(screen.getByText("Something went wrong.")).toBeInTheDocument();
    });
  });

  it("closes dialog on success", async () => {
    mockMutateAsync.mockResolvedValue({});
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    render(<InviteUserDialog open={true} onOpenChange={onOpenChange} />, {
      wrapper,
    });

    await user.type(screen.getByLabelText("Email *"), "test@example.com");
    await user.click(screen.getByRole("button", { name: "Send Invitation" }));

    await waitFor(() => {
      expect(onOpenChange).toHaveBeenCalledWith(false);
    });
  });
});
