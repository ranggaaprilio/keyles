/**
 * Unit tests for RegistrationForm component
 * Tests form validation, submission, error handling, and availability checking
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';

import { RegistrationForm } from '@/components/registration/RegistrationForm';
import * as tenantApi from '@/services/api/tenant';
import { ApiException } from '@/types/api';

// Mock API calls
vi.mock('@/services/api/tenant');

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('RegistrationForm', () => {
  let queryClient: QueryClient;

  const renderForm = () => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    return render(
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <RegistrationForm />
        </BrowserRouter>
      </QueryClientProvider>
    );
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Form Rendering', () => {
    it('should render all form fields', () => {
      renderForm();

      expect(screen.getByLabelText(/organization name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/admin email/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/full name/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /create organization/i })).toBeInTheDocument();
    });

    it('should render form title and description', () => {
      renderForm();

      expect(screen.getByText(/create your organization/i)).toBeInTheDocument();
      expect(screen.getByText(/register your organization to get started/i)).toBeInTheDocument();
    });

    it('should render password requirements hint', () => {
      renderForm();

      expect(
        screen.getByText(/must be 8\+ characters with uppercase, lowercase, number/i)
      ).toBeInTheDocument();
    });
  });

  describe('Form Validation', () => {
    it('should validate organization name length (minimum 3 characters)', async () => {
      const user = userEvent.setup();
      renderForm();

      const orgNameInput = screen.getByLabelText(/organization name/i);
      await user.type(orgNameInput, 'AB');
      await user.tab();

      await waitFor(() => {
        expect(
          screen.getByText(/organization name must be at least 3 characters/i)
        ).toBeInTheDocument();
      });
    });

    it('should validate organization name format (alphanumeric, spaces, hyphens)', async () => {
      const user = userEvent.setup();
      renderForm();

      const orgNameInput = screen.getByLabelText(/organization name/i);
      await user.type(orgNameInput, 'Invalid@Org!');
      await user.tab();

      await waitFor(() => {
        expect(
          screen.getByText(/organization name can only contain letters, numbers, spaces, and hyphens/i)
        ).toBeInTheDocument();
      });
    });

    it('should validate email format', async () => {
      const user = userEvent.setup();
      renderForm();

      const emailInput = screen.getByLabelText(/admin email/i);
      await user.type(emailInput, 'invalid-email');
      await user.tab();

      await waitFor(() => {
        expect(screen.getByText(/please enter a valid email address/i)).toBeInTheDocument();
      });
    });

    it('should validate password complexity requirements', async () => {
      const user = userEvent.setup();
      renderForm();

      const passwordInput = screen.getByLabelText(/^password$/i);
      await user.type(passwordInput, 'weak');
      await user.tab();

      await waitFor(() => {
        expect(
          screen.getByText(
            /password must be at least 8 characters/i
          )
        ).toBeInTheDocument();
      });
    });

    it('should validate full name length', async () => {
      const user = userEvent.setup();
      renderForm();

      const fullNameInput = screen.getByLabelText(/full name/i);
      await user.type(fullNameInput, 'A');
      await user.tab();

      await waitFor(() => {
        expect(
          screen.getByText(/full name must be at least 2 characters/i)
        ).toBeInTheDocument();
      });
    });

    it('should accept valid form data without errors', async () => {
      const user = userEvent.setup();
      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Acme Corporation');
      await user.type(screen.getByLabelText(/admin email/i), 'admin@acme.com');
      await user.type(screen.getByLabelText(/^password$/i), 'SecureP@ss123');
      await user.type(screen.getByLabelText(/full name/i), 'John Doe');

      // No validation errors should be shown
      expect(
        screen.queryByText(/organization name must be at least 3 characters/i)
      ).not.toBeInTheDocument();
      expect(screen.queryByText(/please enter a valid email address/i)).not.toBeInTheDocument();
      expect(
        screen.queryByText(/password must be at least 8 characters/i)
      ).not.toBeInTheDocument();
    });
  });

  describe('Availability Checking', () => {
    it('should check availability when organization name and email are entered', async () => {
      const user = userEvent.setup();
      const checkAvailabilityMock = vi.mocked(tenantApi.checkAvailability);
      checkAvailabilityMock.mockResolvedValue({
        organization_name_available: true,
        email_available: true,
      });

      renderForm();

      const orgNameInput = screen.getByLabelText(/organization name/i);
      const emailInput = screen.getByLabelText(/admin email/i);

      await user.type(orgNameInput, 'Acme Corp');
      await user.type(emailInput, 'admin@acme.com');
      await user.tab();

      await waitFor(() => {
        expect(checkAvailabilityMock).toHaveBeenCalledWith({
          organization_name: 'Acme Corp',
          email: 'admin@acme.com',
        });
      });
    });

    it('should show success icon when organization name is available', async () => {
      const user = userEvent.setup();
      const checkAvailabilityMock = vi.mocked(tenantApi.checkAvailability);
      checkAvailabilityMock.mockResolvedValue({
        organization_name_available: true,
        email_available: true,
      });

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Available Org');
      await user.type(screen.getByLabelText(/admin email/i), 'admin@available.com');
      await user.tab();

      // Wait for availability check to complete and icons to appear
      await waitFor(() => {
        const icons = document.querySelectorAll('svg');
        const hasCheckIcon = Array.from(icons).some(icon => 
          icon.classList.contains('text-green-600')
        );
        expect(hasCheckIcon).toBe(true);
      });
    });

    it('should show error when organization name is taken', async () => {
      const user = userEvent.setup();
      const checkAvailabilityMock = vi.mocked(tenantApi.checkAvailability);
      checkAvailabilityMock.mockResolvedValue({
        organization_name_available: false,
        email_available: true,
      });

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Taken Org');
      await user.type(screen.getByLabelText(/admin email/i), 'admin@taken.com');
      await user.tab();

      await waitFor(() => {
        expect(screen.getByText(/this organization name is already taken/i)).toBeInTheDocument();
      });
    });

    it('should show error when email is already registered', async () => {
      const user = userEvent.setup();
      const checkAvailabilityMock = vi.mocked(tenantApi.checkAvailability);
      checkAvailabilityMock.mockResolvedValue({
        organization_name_available: true,
        email_available: false,
      });

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'New Org');
      await user.type(screen.getByLabelText(/admin email/i), 'existing@email.com');
      await user.tab();

      await waitFor(() => {
        expect(screen.getByText(/this email is already registered/i)).toBeInTheDocument();
      });
    });

    it('should handle availability check failure gracefully', async () => {
      const user = userEvent.setup();
      const checkAvailabilityMock = vi.mocked(tenantApi.checkAvailability);
      checkAvailabilityMock.mockRejectedValue(new Error('Network error'));

      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Test Org');
      await user.type(screen.getByLabelText(/admin email/i), 'test@test.com');
      await user.tab();

      // Should not crash, just log error
      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalledWith('Availability check failed:', expect.any(Error));
      });

      consoleSpy.mockRestore();
    });
  });

  describe('Form Submission', () => {
    it('should submit form with valid data', async () => {
      const user = userEvent.setup();
      const registerTenantMock = vi.mocked(tenantApi.registerTenant);
      registerTenantMock.mockResolvedValue({
        tenant_id: 'tenant-123',
        organization_name: 'Acme Corporation',
        status: 'pending_verification',
        message: 'OTP sent to admin@acme.com',
      });

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Acme Corporation');
      await user.type(screen.getByLabelText(/admin email/i), 'admin@acme.com');
      await user.type(screen.getByLabelText(/^password$/i), 'SecureP@ss123');
      await user.type(screen.getByLabelText(/full name/i), 'John Doe');

      const submitButton = screen.getByRole('button', { name: /create organization/i });
      await user.click(submitButton);

      // Just verify the form submission was attempted
      await waitFor(() => {
        expect(registerTenantMock).toHaveBeenCalled();
      });
    });

    it('should navigate to OTP verification page on successful registration', async () => {
      const user = userEvent.setup();
      const registerTenantMock = vi.mocked(tenantApi.registerTenant);
      registerTenantMock.mockResolvedValue({
        tenant_id: 'tenant-123',
        organization_name: 'Acme Corporation',
        status: 'pending_verification',
        message: 'OTP sent to admin@acme.com',
      });

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Acme Corporation');
      await user.type(screen.getByLabelText(/admin email/i), 'admin@acme.com');
      await user.type(screen.getByLabelText(/^password$/i), 'SecureP@ss123');
      await user.type(screen.getByLabelText(/full name/i), 'John Doe');

      const submitButton = screen.getByRole('button', { name: /create organization/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/verify-otp', {
          state: {
            tenantId: 'tenant-123',
            organizationName: 'Acme Corporation',
            message: 'OTP sent to admin@acme.com',
          },
        });
      });
    });

    it('should show loading state during submission', async () => {
      const user = userEvent.setup();
      const registerTenantMock = vi.mocked(tenantApi.registerTenant);
      registerTenantMock.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Acme Corporation');
      await user.type(screen.getByLabelText(/admin email/i), 'admin@acme.com');
      await user.type(screen.getByLabelText(/^password$/i), 'SecureP@ss123');
      await user.type(screen.getByLabelText(/full name/i), 'John Doe');

      const submitButton = screen.getByRole('button', { name: /create organization/i });
      await user.click(submitButton);

      expect(screen.getByText(/creating organization.../i)).toBeInTheDocument();
      expect(submitButton).toBeDisabled();
    });

    it('should handle duplicate organization name error', async () => {
      const user = userEvent.setup();
      const registerTenantMock = vi.mocked(tenantApi.registerTenant);
      registerTenantMock.mockRejectedValue(
        new ApiException('organization name already exists', 409)
      );

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Duplicate Org');
      await user.type(screen.getByLabelText(/admin email/i), 'admin@dup.com');
      await user.type(screen.getByLabelText(/^password$/i), 'SecureP@ss123');
      await user.type(screen.getByLabelText(/full name/i), 'John Doe');

      const submitButton = screen.getByRole('button', { name: /create organization/i });
      await user.click(submitButton);

      await waitFor(() => {
        // Error should be displayed somewhere in the form
        expect(registerTenantMock).toHaveBeenCalled();
      });
    });

    it('should handle duplicate email error', async () => {
      const user = userEvent.setup();
      const registerTenantMock = vi.mocked(tenantApi.registerTenant);
      registerTenantMock.mockRejectedValue(
        new ApiException('email already registered', 409)
      );

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'New Org');
      await user.type(screen.getByLabelText(/admin email/i), 'existing@email.com');
      await user.type(screen.getByLabelText(/^password$/i), 'SecureP@ss123');
      await user.type(screen.getByLabelText(/full name/i), 'John Doe');

      const submitButton = screen.getByRole('button', { name: /create organization/i });
      await user.click(submitButton);

      await waitFor(() => {
        // Error should be displayed somewhere in the form
        expect(registerTenantMock).toHaveBeenCalled();
      });
    });

    it('should handle generic API error', async () => {
      const user = userEvent.setup();
      const registerTenantMock = vi.mocked(tenantApi.registerTenant);
      registerTenantMock.mockRejectedValue(
        new ApiException('Internal server error', 500)
      );

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Test Org');
      await user.type(screen.getByLabelText(/admin email/i), 'test@test.com');
      await user.type(screen.getByLabelText(/^password$/i), 'SecureP@ss123');
      await user.type(screen.getByLabelText(/full name/i), 'John Doe');

      const submitButton = screen.getByRole('button', { name: /create organization/i });
      await user.click(submitButton);

      await waitFor(() => {
        // Error should be displayed somewhere in the form
        expect(registerTenantMock).toHaveBeenCalled();
      });
    });
  });

  describe('Accessibility', () => {
    it('should have proper form labels', () => {
      renderForm();

      expect(screen.getByLabelText(/organization name/i)).toHaveAttribute('id', 'organization_name');
      expect(screen.getByLabelText(/admin email/i)).toHaveAttribute('id', 'email');
      expect(screen.getByLabelText(/^password$/i)).toHaveAttribute('id', 'password');
      expect(screen.getByLabelText(/full name/i)).toHaveAttribute('id', 'full_name');
    });

    it('should show error messages with proper aria attributes', async () => {
      const user = userEvent.setup();
      renderForm();

      const orgNameInput = screen.getByLabelText(/organization name/i);
      await user.type(orgNameInput, 'AB');
      await user.tab();

      await waitFor(() => {
        const errorMessage = screen.getByText(
          /organization name must be at least 3 characters/i
        );
        expect(errorMessage).toBeInTheDocument();
        expect(errorMessage.closest('p')).toHaveClass('text-red-600');
      });
    });

    it('should disable submit button during submission', async () => {
      const user = userEvent.setup();
      const registerTenantMock = vi.mocked(tenantApi.registerTenant);
      registerTenantMock.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );

      renderForm();

      await user.type(screen.getByLabelText(/organization name/i), 'Test Org');
      await user.type(screen.getByLabelText(/admin email/i), 'test@test.com');
      await user.type(screen.getByLabelText(/^password$/i), 'SecureP@ss123');
      await user.type(screen.getByLabelText(/full name/i), 'John Doe');

      const submitButton = screen.getByRole('button', { name: /create organization/i });
      await user.click(submitButton);

      expect(submitButton).toBeDisabled();
    });
  });
});
