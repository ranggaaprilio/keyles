/**
 * Auth API client for login and authentication
 */

import axios, { AxiosError } from 'axios';
import { ApiException } from '../../types/api';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000,
});

// Add JWT token to requests if available
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  expires_in: number;
  user: {
    id: string;
    email: string;
    full_name: string;
    role: string;
  };
  tenant: {
    id: string;
    organization_name: string;
    status: string;
  };
}

export interface DashboardResponse {
  tenant: {
    id: string;
    organization_name: string;
    status: string;
    created_at: string;
    verified_at: string | null;
  };
  user: {
    id: string;
    full_name: string;
    email: string;
    role: string;
  };
}

/**
 * Login user and return JWT token
 */
export async function login(data: LoginRequest): Promise<LoginResponse> {
  try {
    const response = await apiClient.post<LoginResponse>('/login', data);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError<{ error: string }>;
      throw new ApiException(
        axiosError.response?.status || 500,
        axiosError.response?.data?.error || 'Failed to login',
        axiosError.code
      );
    }
    throw error;
  }
}

/**
 * Get dashboard data (requires authentication)
 */
export async function getDashboard(): Promise<DashboardResponse> {
  try {
    const response = await apiClient.get<DashboardResponse>('/dashboard');
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError<{ error: string }>;
      throw new ApiException(
        axiosError.response?.status || 500,
        axiosError.response?.data?.error || 'Failed to load dashboard',
        axiosError.code
      );
    }
    throw error;
  }
}

/**
 * Logout user (clear token from storage)
 */
export function logout(): void {
  localStorage.removeItem('auth_token');
  localStorage.removeItem('user_data');
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
  return !!localStorage.getItem('auth_token');
}

/**
 * Get stored user data
 */
export function getStoredUser(): LoginResponse['user'] | null {
  const userData = localStorage.getItem('user_data');
  if (!userData) return null;
  try {
    return JSON.parse(userData);
  } catch {
    return null;
  }
}

/**
 * Store auth data in localStorage
 */
export function storeAuthData(token: string, user: LoginResponse['user']): void {
  localStorage.setItem('auth_token', token);
  localStorage.setItem('user_data', JSON.stringify(user));
}
