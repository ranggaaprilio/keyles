/**
 * Registration form component for tenant onboarding
 * Implements TDD approach with real-time validation and availability checking
 */

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { Loader2, CheckCircle2, XCircle, AlertCircle } from 'lucide-react';

import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

import { registrationSchema, RegistrationFormData } from './RegistrationSchema';
import { useTenantRegistration } from '../../hooks/useTenantRegistration';
import { checkAvailability } from '../../services/api/tenant';
import { ApiException } from '../../types/api';

export function RegistrationForm() {
  const navigate = useNavigate();
  const [availabilityStatus, setAvailabilityStatus] = useState<{
    orgName?: boolean;
    email?: boolean;
  }>({});
  const [checkingAvailability, setCheckingAvailability] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    watch,
    setError,
  } = useForm<RegistrationFormData>({
    resolver: zodResolver(registrationSchema),
    mode: 'onChange',
  });

  const mutation = useTenantRegistration({
    onSuccess: (data) => {
      // Navigate to verification page with tenant info
      navigate('/verify-otp', {
        state: {
          tenantId: data.tenant_id,
          organizationName: data.organization_name,
          message: data.message,
        },
      });
    },
    onError: (error: ApiException) => {
      // Map API errors to form errors
      if (error.status === 409) {
        if (error.error.includes('organization name')) {
          setError('organization_name', {
            type: 'manual',
            message: error.error,
          });
        } else if (error.error.includes('email')) {
          setError('email', {
            type: 'manual',
            message: error.error,
          });
        }
      } else {
        setError('root', {
          type: 'manual',
          message: error.error || 'Registration failed. Please try again.',
        });
      }
    },
  });

  // Watch fields for real-time availability checking
  const orgName = watch('organization_name');
  const email = watch('email');

  // Debounced availability check
  const checkFieldAvailability = async () => {
    if (!orgName || !email || orgName.length < 3 || !email.includes('@')) {
      return;
    }

    setCheckingAvailability(true);
    try {
      const result = await checkAvailability({
        organization_name: orgName,
        email: email,
      });
      setAvailabilityStatus({
        orgName: result.organization_name_available,
        email: result.email_available,
      });
    } catch (error) {
      // Silently fail availability check - user can still submit
      console.error('Availability check failed:', error);
    } finally {
      setCheckingAvailability(false);
    }
  };

  const onSubmit = async (data: RegistrationFormData) => {
    mutation.mutate(data);
  };

  const getAvailabilityIcon = (available?: boolean) => {
    if (available === undefined) return null;
    if (available) {
      return <CheckCircle2 className="h-5 w-5 text-green-600" />;
    }
    return <XCircle className="h-5 w-5 text-red-600" />;
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">
            Create Your Organization
          </CardTitle>
          <CardDescription className="text-center">
            Register your organization to get started with Keyles SSO
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            {/* Organization Name */}
            <div className="space-y-2">
              <Label htmlFor="organization_name">Organization Name</Label>
              <div className="relative">
                <Input
                  id="organization_name"
                  placeholder="Acme Corporation"
                  {...register('organization_name')}
                  onBlur={checkFieldAvailability}
                  className={errors.organization_name ? 'border-red-500' : ''}
                />
                {checkingAvailability && (
                  <div className="absolute right-3 top-2.5">
                    <Loader2 className="h-5 w-5 animate-spin text-gray-400" />
                  </div>
                )}
                {!checkingAvailability && availabilityStatus.orgName !== undefined && (
                  <div className="absolute right-3 top-2.5">
                    {getAvailabilityIcon(availabilityStatus.orgName)}
                  </div>
                )}
              </div>
              {errors.organization_name && (
                <p className="text-sm text-red-600 flex items-center gap-1">
                  <AlertCircle className="h-4 w-4" />
                  {errors.organization_name.message}
                </p>
              )}
              {availabilityStatus.orgName === false && (
                <p className="text-sm text-red-600">
                  This organization name is already taken
                </p>
              )}
            </div>

            {/* Email */}
            <div className="space-y-2">
              <Label htmlFor="email">Admin Email</Label>
              <div className="relative">
                <Input
                  id="email"
                  type="email"
                  placeholder="admin@acme.com"
                  {...register('email')}
                  onBlur={checkFieldAvailability}
                  className={errors.email ? 'border-red-500' : ''}
                />
                {checkingAvailability && (
                  <div className="absolute right-3 top-2.5">
                    <Loader2 className="h-5 w-5 animate-spin text-gray-400" />
                  </div>
                )}
                {!checkingAvailability && availabilityStatus.email !== undefined && (
                  <div className="absolute right-3 top-2.5">
                    {getAvailabilityIcon(availabilityStatus.email)}
                  </div>
                )}
              </div>
              {errors.email && (
                <p className="text-sm text-red-600 flex items-center gap-1">
                  <AlertCircle className="h-4 w-4" />
                  {errors.email.message}
                </p>
              )}
              {availabilityStatus.email === false && (
                <p className="text-sm text-red-600">
                  This email is already registered
                </p>
              )}
            </div>

            {/* Password */}
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                {...register('password')}
                className={errors.password ? 'border-red-500' : ''}
              />
              {errors.password && (
                <p className="text-sm text-red-600 flex items-center gap-1">
                  <AlertCircle className="h-4 w-4" />
                  {errors.password.message}
                </p>
              )}
              <p className="text-xs text-gray-500">
                Must be 8+ characters with uppercase, lowercase, number, and special character
              </p>
            </div>

            {/* Full Name */}
            <div className="space-y-2">
              <Label htmlFor="full_name">Full Name</Label>
              <Input
                id="full_name"
                placeholder="John Doe"
                {...register('full_name')}
                className={errors.full_name ? 'border-red-500' : ''}
              />
              {errors.full_name && (
                <p className="text-sm text-red-600 flex items-center gap-1">
                  <AlertCircle className="h-4 w-4" />
                  {errors.full_name.message}
                </p>
              )}
            </div>

            {/* Global Error */}
            {errors.root && (
              <div className="rounded-md bg-red-50 p-3">
                <p className="text-sm text-red-800 flex items-center gap-2">
                  <AlertCircle className="h-4 w-4" />
                  {errors.root.message}
                </p>
              </div>
            )}

            {/* Submit Button */}
            <Button
              type="submit"
              className="w-full"
              disabled={isSubmitting || mutation.isPending}
            >
              {mutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating Organization...
                </>
              ) : (
                'Create Organization'
              )}
            </Button>
          </form>

          <div className="mt-4 text-center text-sm">
            <p className="text-gray-600">
              Already have an account?{' '}
              <a href="/login" className="text-blue-600 hover:underline">
                Sign in
              </a>
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
