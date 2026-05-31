/**
 * Unit tests for IntegrationDocs component
 * T040 - Verify OAuth flow documentation rendering
 */

import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { IntegrationDocs } from '@/components/clients/IntegrationDocs';
import type { Client } from '@/types/client';

const confidentialClient: Client = {
  client_id: 'test-client-123',
  client_name: 'Test Backend App',
  description: 'A backend application',
  client_type: 'confidential',
  redirect_uris: ['https://app.example.com/callback'],
  is_active: true,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

const publicClient: Client = {
  client_id: 'test-public-456',
  client_name: 'Test SPA',
  description: 'A public SPA application',
  client_type: 'public',
  redirect_uris: ['https://spa.example.com/callback'],
  is_active: true,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('IntegrationDocs', () => {
  it('renders integration guide title', () => {
    render(<IntegrationDocs client={confidentialClient} />);
    expect(screen.getByText('Integration Guide')).toBeInTheDocument();
  });

  it('displays client type badge', () => {
    render(<IntegrationDocs client={confidentialClient} />);
    expect(screen.getByText('confidential')).toBeInTheDocument();
  });

  it('interpolates client_id into code examples for confidential', () => {
    render(<IntegrationDocs client={confidentialClient} />);
    const codeBlocks = document.querySelectorAll('code');
    const allCode = Array.from(codeBlocks).map(el => el.textContent).join('\n');
    expect(allCode).toContain('test-client-123');
  });

  it('interpolates client_id into code examples for public', () => {
    render(<IntegrationDocs client={publicClient} />);
    const codeBlocks = document.querySelectorAll('code');
    const allCode = Array.from(codeBlocks).map(el => el.textContent).join('\n');
    expect(allCode).toContain('test-public-456');
  });

  it('shows Authorization Code tab for confidential clients', () => {
    render(<IntegrationDocs client={confidentialClient} />);
    expect(screen.getByText('Authorization Code')).toBeInTheDocument();
  });

  it('shows PKCE Flow tab for public clients', () => {
    render(<IntegrationDocs client={publicClient} />);
    expect(screen.getByText('PKCE Flow')).toBeInTheDocument();
  });

  it('shows client_secret placeholder for confidential clients', () => {
    render(<IntegrationDocs client={confidentialClient} />);
    const codeBlocks = document.querySelectorAll('code');
    const allCode = Array.from(codeBlocks).map(el => el.textContent).join('\n');
    expect(allCode).toContain('YOUR_CLIENT_SECRET');
  });

  it('shows PKCE parameters for public clients', () => {
    render(<IntegrationDocs client={publicClient} />);
    const codeBlocks = document.querySelectorAll('code');
    const allCode = Array.from(codeBlocks).map(el => el.textContent).join('\n');
    expect(allCode).toContain('code_challenge');
    expect(allCode).toContain('code_verifier');
  });

  it('shows mandatory PKCE parameters for confidential clients', () => {
    render(<IntegrationDocs client={confidentialClient} />);
    const codeBlocks = document.querySelectorAll('code');
    const allCode = Array.from(codeBlocks).map(el => el.textContent).join('\n');
    expect(allCode).toContain('code_challenge');
    expect(allCode).toContain('code_verifier');
  });

  it('renders Token Exchange tab', () => {
    render(<IntegrationDocs client={confidentialClient} />);
    expect(screen.getByText('Token Exchange')).toBeInTheDocument();
  });

  it('renders Token Refresh tab', () => {
    render(<IntegrationDocs client={confidentialClient} />);
    expect(screen.getByText('Token Refresh')).toBeInTheDocument();
  });

  it('shows redirect URI requirements note', () => {
    render(<IntegrationDocs client={confidentialClient} />);
    expect(screen.getByText(/Redirect URI Requirements/)).toBeInTheDocument();
  });
});
