/**
 * AcceptInvitationForm component tests
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AcceptInvitationForm } from "@/components/users/AcceptInvitationForm";

describe("AcceptInvitationForm", () => {
  const defaultProps = {
    email: "alice@example.com",
    displayName: "Alice",
    onSubmit: vi.fn(),
    isSubmitting: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders with read-only email pre-populated", () => {
    render(<AcceptInvitationForm {...defaultProps} />);

    const emailInput = screen.getByDisplayValue("alice@example.com");
    expect(emailInput).toBeDisabled();
  });

  it("renders display name when provided", () => {
    render(<AcceptInvitationForm {...defaultProps} />);

    expect(screen.getByDisplayValue("Alice")).toBeInTheDocument();
  });

  it("shows validation error for password under 8 chars", async () => {
    const user = userEvent.setup();
    render(<AcceptInvitationForm {...defaultProps} />);

    await user.type(screen.getByLabelText("Password *"), "Ab1");
    await user.type(screen.getByLabelText("Confirm Password *"), "Ab1");
    await user.click(screen.getByText("Create Account"));

    await waitFor(() => {
      expect(
        screen.getByText("Password must be at least 8 characters"),
      ).toBeInTheDocument();
    });
  });

  it("shows validation error for mismatched passwords", async () => {
    const user = userEvent.setup();
    render(<AcceptInvitationForm {...defaultProps} />);

    await user.type(screen.getByLabelText("Password *"), "StrongPass1");
    await user.type(
      screen.getByLabelText("Confirm Password *"),
      "DifferentPass2",
    );
    await user.click(screen.getByText("Create Account"));

    await waitFor(() => {
      expect(screen.getByText("Passwords do not match")).toBeInTheDocument();
    });
  });

  it("shows password strength indicator", async () => {
    const user = userEvent.setup();
    render(<AcceptInvitationForm {...defaultProps} />);

    await user.type(screen.getByLabelText("Password *"), "StrongPass1!");

    // Strong indicator should appear with a strong password
    await waitFor(() => {
      expect(screen.getByText("Strong")).toBeInTheDocument();
    });
  });

  it("calls onSubmit with password on successful submission", async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    const user = userEvent.setup();
    render(<AcceptInvitationForm {...defaultProps} onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText("Password *"), "StrongPass1");
    await user.type(screen.getByLabelText("Confirm Password *"), "StrongPass1");
    await user.click(screen.getByText("Create Account"));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith("StrongPass1");
    });
  });

  it("shows error message when provided", () => {
    render(<AcceptInvitationForm {...defaultProps} error="Token expired" />);

    expect(screen.getByText("Token expired")).toBeInTheDocument();
  });
});
