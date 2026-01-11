/**
 * RoleManagement Component Unit Tests
 * Tests for user role assignment and management UI
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { describe, test, expect, beforeEach, vi } from 'vitest';
import { RoleManagement } from '../../../src/components/admin/RoleManagement';
import * as roleService from '../../../src/services/roleService';

// Mock the role service
vi.mock('../../../src/services/roleService');

describe('RoleManagement Component', () => {
  const mockTenantId = 'tenant-123';

  const mockRoles = [
    {
      id: 'role-1',
      user_id: 'user-1',
      client_id: 'client-1',
      role: 'admin',
      assigned_at: '2025-01-01T00:00:00Z',
      assigned_by: 'admin-user',
      is_active: true,
      user_email: 'user@example.com',
      client_name: 'Test Client',
    },
    {
      id: 'role-2',
      user_id: 'user-1',
      client_id: 'client-2',
      role: 'viewer',
      assigned_at: '2025-01-02T00:00:00Z',
      assigned_by: 'admin-user',
      is_active: true,
      user_email: 'user@example.com',
      client_name: 'Another Client',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  test('renders role management interface', () => {
    render(<RoleManagement tenantId={mockTenantId} />);
    
    expect(screen.getByText('Role Management')).toBeInTheDocument();
    expect(screen.getByText('Assign Role')).toBeInTheDocument();
    expect(screen.getByText('Filter By')).toBeInTheDocument();
  });

  test('loads and displays user roles when filter is applied', async () => {
    vi.mocked(roleService.listUserRoles).mockResolvedValue({
      roles: mockRoles,
      total: 2,
    });

    render(<RoleManagement tenantId={mockTenantId} />);
    
    // Set filter to user
    const filterSelect = screen.getByDisplayValue('User');
    expect(filterSelect).toBeInTheDocument();

    // Enter user ID
    const userIdInput = screen.getByPlaceholderText('Enter user ID...');
    fireEvent.change(userIdInput, { target: { value: 'user-1' } });

    // Click load button
    const loadButton = screen.getByText('Load Roles');
    fireEvent.click(loadButton);

    // Wait for roles to load
    await waitFor(() => {
      expect(roleService.listUserRoles).toHaveBeenCalledWith('user-1');
      expect(screen.getByText('Test Client')).toBeInTheDocument();
      expect(screen.getByText('Another Client')).toBeInTheDocument();
    });

    expect(screen.getByText('Total roles: 2')).toBeInTheDocument();
  });

  test('loads and displays client roles when filter is changed', async () => {
    vi.mocked(roleService.listClientRoles).mockResolvedValue({
      roles: [mockRoles[0]],
      total: 1,
    });

    render(<RoleManagement tenantId={mockTenantId} />);
    
    // Change filter to client
    const filterSelect = screen.getByDisplayValue('User');
    fireEvent.change(filterSelect, { target: { value: 'client' } });

    // Enter client ID
    const clientIdInput = screen.getByPlaceholderText('Enter client ID...');
    fireEvent.change(clientIdInput, { target: { value: 'client-1' } });

    // Click load button
    const loadButton = screen.getByText('Load Roles');
    fireEvent.click(loadButton);

    // Wait for roles to load
    await waitFor(() => {
      expect(roleService.listClientRoles).toHaveBeenCalledWith('client-1');
      expect(screen.getByText('Test Client')).toBeInTheDocument();
    });
  });

  test('shows assign role form when button is clicked', () => {
    render(<RoleManagement tenantId={mockTenantId} />);
    
    const assignButton = screen.getByText('Assign Role');
    fireEvent.click(assignButton);

    expect(screen.getByText('Assign New Role')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('user-uuid')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('client-id')).toBeInTheDocument();
    // The form has role select with options
    expect(screen.getByText('user')).toBeInTheDocument();
    expect(screen.getByText('admin')).toBeInTheDocument();
  });

  test('submits assign role form successfully', async () => {
    const newRole = {
      id: 'role-3',
      user_id: 'user-2',
      client_id: 'client-1',
      role: 'editor',
      assigned_at: '2025-01-03T00:00:00Z',
      assigned_by: 'admin-user',
      is_active: true,
    };

    vi.mocked(roleService.assignRole).mockResolvedValue(newRole);

    render(<RoleManagement tenantId={mockTenantId} />);
    
    // Open form
    const assignButton = screen.getByText('Assign Role');
    fireEvent.click(assignButton);

    // Fill form
    const userIdInput = screen.getByPlaceholderText('user-uuid');
    const clientIdInput = screen.getByPlaceholderText('client-id');
    // Get all selects (one for filter, one for role in form)
    const selects = screen.getAllByRole('combobox');
    const roleSelect = selects[1]; // Second select is the role selector in the form

    fireEvent.change(userIdInput, { target: { value: 'user-2' } });
    fireEvent.change(clientIdInput, { target: { value: 'client-1' } });
    fireEvent.change(roleSelect, { target: { value: 'editor' } });

    // Submit form
    const submitButton = screen.getByText('Assign Role');
    fireEvent.click(submitButton);

    // Wait for success
    await waitFor(() => {
      expect(roleService.assignRole).toHaveBeenCalledWith({
        user_id: 'user-2',
        client_id: 'client-1',
        role: 'editor',
      });
      expect(screen.getByText('Role assigned successfully')).toBeInTheDocument();
    });
  });

  test('handles assign role error', async () => {
    vi.mocked(roleService.assignRole).mockRejectedValue({
      response: {
        data: {
          message: 'User does not exist',
        },
      },
    });

    render(<RoleManagement tenantId={mockTenantId} />);
    
    // Open and fill form
    const assignButton = screen.getByText('Assign Role');
    fireEvent.click(assignButton);

    const userIdInput = screen.getByPlaceholderText('user-uuid');
    const clientIdInput = screen.getByPlaceholderText('client-id');

    fireEvent.change(userIdInput, { target: { value: 'invalid-user' } });
    fireEvent.change(clientIdInput, { target: { value: 'client-1' } });

    // Submit
    const submitButton = screen.getByText('Assign Role');
    fireEvent.click(submitButton);

    // Wait for error
    await waitFor(() => {
      expect(screen.getByText('User does not exist')).toBeInTheDocument();
    });
  });

  test('revokes role with confirmation', async () => {
    vi.mocked(roleService.listUserRoles).mockResolvedValue({
      roles: mockRoles,
      total: 2,
    });
    vi.mocked(roleService.revokeRole).mockResolvedValue();

    // Mock window.confirm
    const confirmSpy = vi.spyOn(window, 'confirm');
    confirmSpy.mockReturnValue(true);

    render(<RoleManagement tenantId={mockTenantId} />);
    
    // Load roles
    const userIdInput = screen.getByPlaceholderText('Enter user ID...');
    fireEvent.change(userIdInput, { target: { value: 'user-1' } });
    const loadButton = screen.getByText('Load Roles');
    fireEvent.click(loadButton);

    // Wait for roles to load
    await waitFor(() => {
      expect(screen.getByText('Test Client')).toBeInTheDocument();
    });

    // Click revoke button
    const revokeButtons = screen.getAllByText('Revoke');
    fireEvent.click(revokeButtons[0]);

    // Verify confirmation and service call
    await waitFor(() => {
      expect(confirmSpy).toHaveBeenCalled();
      expect(roleService.revokeRole).toHaveBeenCalledWith({
        user_id: 'user-1',
        client_id: 'client-1',
        role: 'admin',
      });
    });

    confirmSpy.mockRestore();
  });

  test('cancels role revocation when user declines confirmation', async () => {
    vi.mocked(roleService.listUserRoles).mockResolvedValue({
      roles: mockRoles,
      total: 2,
    });

    // Mock window.confirm to return false
    const confirmSpy = vi.spyOn(window, 'confirm');
    confirmSpy.mockReturnValue(false);

    render(<RoleManagement tenantId={mockTenantId} />);
    
    // Load roles
    const userIdInput = screen.getByPlaceholderText('Enter user ID...');
    fireEvent.change(userIdInput, { target: { value: 'user-1' } });
    const loadButton = screen.getByText('Load Roles');
    fireEvent.click(loadButton);

    await waitFor(() => {
      expect(screen.getByText('Test Client')).toBeInTheDocument();
    });

    // Click revoke button
    const revokeButtons = screen.getAllByText('Revoke');
    fireEvent.click(revokeButtons[0]);

    // Verify revoke was NOT called
    expect(roleService.revokeRole).not.toHaveBeenCalled();

    confirmSpy.mockRestore();
  });

  test('displays empty state when no roles exist', async () => {
    vi.mocked(roleService.listUserRoles).mockResolvedValue({
      roles: [],
      total: 0,
    });

    render(<RoleManagement tenantId={mockTenantId} />);
    
    const userIdInput = screen.getByPlaceholderText('Enter user ID...');
    fireEvent.change(userIdInput, { target: { value: 'user-1' } });
    const loadButton = screen.getByText('Load Roles');
    fireEvent.click(loadButton);

    await waitFor(() => {
      expect(screen.getByText(/No roles found/)).toBeInTheDocument();
    });
  });

  test('displays loading state', () => {
    vi.mocked(roleService.listUserRoles).mockImplementation(() => new Promise(() => {}));

    render(<RoleManagement tenantId={mockTenantId} />);
    
    const userIdInput = screen.getByPlaceholderText('Enter user ID...');
    fireEvent.change(userIdInput, { target: { value: 'user-1' } });
    const loadButton = screen.getByText('Load Roles');
    fireEvent.click(loadButton);

    expect(screen.getByText('Loading roles...')).toBeInTheDocument();
  });

  test('displays role status badges correctly', async () => {
    const mixedRoles = [
      { ...mockRoles[0], is_active: true },
      { ...mockRoles[1], is_active: false },
    ];

    vi.mocked(roleService.listUserRoles).mockResolvedValue({
      roles: mixedRoles,
      total: 2,
    });

    render(<RoleManagement tenantId={mockTenantId} />);
    
    const userIdInput = screen.getByPlaceholderText('Enter user ID...');
    fireEvent.change(userIdInput, { target: { value: 'user-1' } });
    const loadButton = screen.getByText('Load Roles');
    fireEvent.click(loadButton);

    await waitFor(() => {
      expect(screen.getByText('Active')).toBeInTheDocument();
      expect(screen.getByText('Inactive')).toBeInTheDocument();
    });
  });

  test('displays FR-006e token revocation warning', () => {
    render(<RoleManagement tenantId={mockTenantId} />);
    
    expect(screen.getByText(/Revoking a role will also invalidate all refresh tokens/)).toBeInTheDocument();
    expect(screen.getByText(/FR-006e/)).toBeInTheDocument();
  });
});
