/**
 * Unit tests for SecretDisplay component
 * T050 - Secret display modal, copy-to-clipboard, save confirmation
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SecretDisplay } from '@/components/clients/SecretDisplay';

// Mock navigator.clipboard
const mockClipboard = { writeText: vi.fn().mockResolvedValue(undefined) };
Object.assign(navigator, { clipboard: mockClipboard });

describe('SecretDisplay', () => {
  const defaultProps = {
    open: true,
    onClose: vi.fn(),
    clientId: 'test-client-123',
    clientSecret: 'super-secret-value-xyz',
    clientName: 'Test App',
  };

  it('displays client_id', () => {
    render(<SecretDisplay {...defaultProps} />);
    expect(screen.getByDisplayValue('test-client-123')).toBeInTheDocument();
  });

  it('displays client_secret for confidential clients', () => {
    render(<SecretDisplay {...defaultProps} />);
    expect(screen.getByDisplayValue('super-secret-value-xyz')).toBeInTheDocument();
  });

  it('shows warning about secret not being retrievable again', () => {
    render(<SecretDisplay {...defaultProps} />);
    expect(screen.getByText(/cannot be retrieved/i)).toBeInTheDocument();
  });

  it('does not render when open is false', () => {
    render(<SecretDisplay {...defaultProps} open={false} />);
    expect(screen.queryByText(/test-client-123/)).not.toBeInTheDocument();
  });

  it('renders copy buttons', () => {
    render(<SecretDisplay {...defaultProps} />);
    // Copy buttons have icon-only content; find by role
    const buttons = screen.getAllByRole('button');
    // Should have at least: 2 copy buttons + Done button
    expect(buttons.length).toBeGreaterThanOrEqual(3);
  });
});
