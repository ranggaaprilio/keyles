import { fireEvent, render, screen } from '@testing-library/react';
import { Link, MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { ClientManagementPage } from '../pages/ClientManagementPage';

vi.mock('../hooks/useClients', () => ({
  useClients: () => ({
    data: { clients: [], total: 0, page: 1, total_pages: 0 },
    isLoading: false,
    isError: false,
  }),
  useClient: () => ({ data: undefined, isLoading: true, isError: false }),
  useCreateClient: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateClient: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteClient: () => ({ mutate: vi.fn(), isPending: false }),
  useRotateSecret: () => ({ mutate: vi.fn(), isPending: false }),
}));

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      {/* Stands in for the header back link, which lives outside the page */}
      <Link to="/dashboard/clients">back</Link>
      <Routes>
        <Route path="/dashboard/clients" element={<ClientManagementPage />} />
        <Route path="/dashboard/clients/new" element={<ClientManagementPage />} />
        <Route path="/dashboard/clients/:clientId" element={<ClientManagementPage />} />
        <Route path="/dashboard/clients/:clientId/edit" element={<ClientManagementPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('ClientManagementPage', () => {
  it('shows the create form when opened directly at /new', () => {
    renderAt('/dashboard/clients/new');
    expect(screen.getByText('Register New Client Application')).toBeInTheDocument();
  });

  it('returns to the list when the URL goes back to /dashboard/clients', () => {
    renderAt('/dashboard/clients/new');
    expect(screen.getByText('Register New Client Application')).toBeInTheDocument();

    fireEvent.click(screen.getByText('back'));

    expect(screen.getByRole('heading', { name: 'Client Applications' })).toBeInTheDocument();
    expect(screen.queryByText('Register New Client Application')).not.toBeInTheDocument();
  });
});
