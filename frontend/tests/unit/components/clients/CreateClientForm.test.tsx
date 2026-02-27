/**
 * Unit tests for CreateClientForm component
 * T050 - Form validation, client type selection
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { CreateClientForm } from '@/components/clients/CreateClientForm';

describe('CreateClientForm', () => {
  const mockOnSubmit = vi.fn();
  const mockOnCancel = vi.fn();

  it('renders the form with all fields', () => {
    render(<CreateClientForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    expect(screen.getByText(/application name/i)).toBeInTheDocument();
    expect(screen.getByText(/client type/i)).toBeInTheDocument();
    expect(screen.getByText('Redirect URIs *')).toBeInTheDocument();
  });

  it('renders cancel button', () => {
    render(<CreateClientForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    const cancelBtn = screen.getByRole('button', { name: /cancel/i });
    expect(cancelBtn).toBeInTheDocument();
  });

  it('calls onCancel when cancel button is clicked', async () => {
    const user = userEvent.setup();
    render(<CreateClientForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    await user.click(screen.getByRole('button', { name: /cancel/i }));
    expect(mockOnCancel).toHaveBeenCalled();
  });

  it('shows loading state when isLoading is true', () => {
    render(<CreateClientForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} isLoading={true} />);
    const submitBtn = screen.getByRole('button', { name: /register|creat|submit/i });
    expect(submitBtn).toBeDisabled();
  });

  it('has radio buttons for confidential and public client types', () => {
    render(<CreateClientForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />);
    expect(screen.getByText(/confidential/i)).toBeInTheDocument();
    expect(screen.getByText(/public/i)).toBeInTheDocument();
  });
});
