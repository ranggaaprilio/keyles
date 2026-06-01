import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { OAuthErrorPage } from '../pages/OAuthErrorPage';

describe('OAuthErrorPage', () => {
  it.each([
    ['invalid_client', 'The application is not registered or is currently unavailable.'],
    ['invalid_request', 'This sign-in request is invalid or has expired.'],
    ['access_denied', 'Access was not granted to the application.'],
    ['temporarily_unavailable', 'Sign-in is temporarily unavailable. Please try again later.'],
  ])('maps %s to a friendly message', (code, message) => {
    render(<MemoryRouter initialEntries={[`/oauth2/error?error=${code}`]}><OAuthErrorPage /></MemoryRouter>);
    expect(screen.getByText(message)).toBeInTheDocument();
  });

  it('does not render an untrusted description', () => {
    render(<MemoryRouter initialEntries={['/oauth2/error?error=unknown&error_description=secret-details']}><OAuthErrorPage /></MemoryRouter>);
    expect(screen.getByText('The sign-in request could not be completed.')).toBeInTheDocument();
    expect(screen.queryByText('secret-details')).not.toBeInTheDocument();
  });
});
