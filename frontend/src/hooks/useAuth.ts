/**
 * useAuth hook for authentication state management
 */

import { useMutation, useQuery, useQueryClient, UseMutationResult, UseQueryResult } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import {
  login,
  logout as logoutApi,
  getDashboard,
  isAuthenticated as checkAuth,
  getStoredUser,
  storeAuthData,
  LoginRequest,
  LoginResponse,
  DashboardResponse,
} from '../services/api/auth';
import { ApiException } from '../types/api';

interface UseAuthReturn {
  // Login mutation
  loginMutation: UseMutationResult<LoginResponse, ApiException, LoginRequest>;
  isLoggingIn: boolean;
  
  // Dashboard query
  dashboardQuery: UseQueryResult<DashboardResponse, ApiException>;
  
  // Auth state
  isAuthenticated: boolean;
  user: LoginResponse['user'] | null;
  
  // Actions
  login: (data: LoginRequest) => void;
  logout: () => void;
}

export function useAuth(): UseAuthReturn {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isAuthenticated = checkAuth();
  const user = getStoredUser();

  // Login mutation
  const loginMutation = useMutation<LoginResponse, ApiException, LoginRequest>({
    mutationFn: login,
    onSuccess: (data) => {
      // Store auth data
      storeAuthData(data.token, data.user);
      
      // Invalidate queries to refetch with new auth
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });

  // Dashboard query (only fetch if authenticated)
  const dashboardQuery = useQuery<DashboardResponse, ApiException>({
    queryKey: ['dashboard'],
    queryFn: getDashboard,
    enabled: isAuthenticated,
    retry: false,
  });

  // Logout function
  const handleLogout = () => {
    logoutApi();
    queryClient.clear();
    navigate('/login', { replace: true });
  };

  // Login function
  const handleLogin = (data: LoginRequest) => {
    loginMutation.mutate(data);
  };

  return {
    loginMutation,
    isLoggingIn: loginMutation.isPending,
    dashboardQuery,
    isAuthenticated,
    user,
    login: handleLogin,
    logout: handleLogout,
  };
}
