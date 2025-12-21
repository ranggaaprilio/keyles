/**
 * Tenant-related types for the multi-tenant SSO platform
 */

import { TenantStatus } from './api';

export interface Tenant {
  id: string;
  organizationName: string;
  status: TenantStatus;
  createdAt: string;
  verifiedAt?: string;
  updatedAt: string;
}

export interface AdminUser {
  id: string;
  tenantId: string;
  fullName: string;
  email: string;
  role: UserRole;
  createdAt: string;
  updatedAt: string;
}

export type UserRole = 'admin' | 'owner' | 'member' | 'viewer';

export interface RegisterTenantRequest {
  organization_name: string;
  email: string;
  password: string;
  full_name: string;
}

export interface RegisterTenantResponse {
  tenant_id: string;
  organization_name: string;
  status: TenantStatus;
  message: string;
}

export interface CheckAvailabilityRequest {
  organization_name: string;
  email: string;
}

export interface CheckAvailabilityResponse {
  organization_name_available: boolean;
  email_available: boolean;
}

export interface VerifyOTPRequest {
  tenant_id: string;
  otp_code: string;
}

export interface VerifyOTPResponse {
  tenant_id: string;
  status: TenantStatus;
  message: string;
}

export interface ResendOTPRequest {
  tenant_id: string;
}

export interface ResendOTPResponse {
  tenant_id: string;
  message: string;
}
