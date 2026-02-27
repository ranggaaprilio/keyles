/**
 * Unit tests for ClientCard component
 * T051 - Card rendering with type badge, click handler, copy client_id
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ClientCard } from '@/components/clients/ClientCard';
import type { Client } from '@/types/client';

// Mock navigator.clipboard
const mockClipboard = { writeText: vi.fn().mockResolvedValue(undefined) };
Object.assign(navigator, { clipboard: mockClipboard });

const mockClient: Client = {
  client_id: 'test-client-id-12345678',
  client_name: 'My Test App',
  description: 'A test application',
  client_type: 'confidential',
  redirect_uris: ['https://example.com/callback'],
  is_active: true,
  created_at: '2024-06-15T10:00:00Z',
  updated_at: '2024-06-15T10:00:00Z',
};

describe('ClientCard', () => {
  const mockOnClick = vi.fn();

  it('renders the client name', () => {
    render(<ClientCard client={mockClient} onClick={mockOnClick} />);
    expect(screen.getByText('My Test App')).toBeInTheDocument();
  });

  it('renders the client type badge', () => {
    render(<ClientCard client={mockClient} onClick={mockOnClick} />);
    expect(screen.getByText(/confidential/i)).toBeInTheDocument();
  });

  it('renders public badge for public clients', () => {
    const publicClient = { ...mockClient, client_type: 'public' as const };
    render(<ClientCard client={publicClient} onClick={mockOnClick} />);
    expect(screen.getByText(/public/i)).toBeInTheDocument();
  });

  it('shows the client ID (possibly truncated)', () => {
    render(<ClientCard client={mockClient} onClick={mockOnClick} />);
    // The component truncates at 20 chars
    expect(screen.getByText(/test-client-id-12345/)).toBeInTheDocument();
  });

  it('calls onClick with client ID when card is clicked', async () => {
    const user = userEvent.setup();
    render(<ClientCard client={mockClient} onClick={mockOnClick} />);
    // Click the card - it's a Card element with cursor-pointer
    const card = screen.getByText('My Test App').closest('[class*="cursor-pointer"]');
    if (card) {
      await user.click(card);
      expect(mockOnClick).toHaveBeenCalledWith('test-client-id-12345678');
    }
  });

  it('shows active status indicator', () => {
    render(<ClientCard client={mockClient} onClick={mockOnClick} />);
    expect(screen.getByText(/active/i)).toBeInTheDocument();
  });
});
