/**
 * Tenant API client for registration and availability checking
 */

import axios, { AxiosError } from 'axios';
import {
  RegisterTenantRequest,
  RegisterTenantResponse,
  CheckAvailabilityRequest,
  CheckAvailabilityResponse,
} from '../../types/tenant';
import { ApiException } from '../../types/api';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000,
});

/**
 * Register a new tenant organization
 */
export async function registerTenant(
  data: RegisterTenantRequest
): Promise<RegisterTenantResponse> {
  try {
    const response = await apiClient.post<RegisterTenantResponse>('/register', data);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError<{ error: string }>;
      throw new ApiException(
        axiosError.response?.status || 500,
        axiosError.response?.data?.error || 'Failed to register tenant',
        axiosError.code
      );
    }
    throw error;
  }
}

/**
 * Check if organization name and email are available
 */
export async function checkAvailability(
  data: CheckAvailabilityRequest
): Promise<CheckAvailabilityResponse> {
  try {
    const response = await apiClient.get<CheckAvailabilityResponse>(
      '/check-availability',
      {
        params: {
          organization_name: data.organization_name,
          email: data.email,
        },
      }
    );
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError<{ error: string }>;
      throw new ApiException(
        axiosError.response?.status || 500,
        axiosError.response?.data?.error || 'Failed to check availability',
        axiosError.code
      );
    }
    throw error;
  }
}
