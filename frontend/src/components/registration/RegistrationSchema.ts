/**
 * Zod validation schema for tenant registration form
 */

import { z } from 'zod';

/**
 * Registration form validation schema
 * Matches backend validation rules exactly
 */
export const registrationSchema = z.object({
  organization_name: z
    .string()
    .min(3, 'Organization name must be at least 3 characters')
    .max(100, 'Organization name must be at most 100 characters')
    .regex(
      /^[a-zA-Z0-9\s\-]{3,100}$/,
      'Organization name can only contain letters, numbers, spaces, and hyphens'
    ),
  
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Please enter a valid email address')
    .max(255, 'Email must be at most 255 characters'),
  
  password: z
    .string()
    .min(8, 'Password must be at least 8 characters long')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number')
    .regex(/[@$!%*?&]/, 'Password must contain at least one special character (@$!%*?&)'),
  
  full_name: z
    .string()
    .min(2, 'Full name must be at least 2 characters')
    .max(100, 'Full name must be at most 100 characters')
    .regex(/^[a-zA-Z\s\-'.]+$/, 'Full name can only contain letters, spaces, hyphens, apostrophes, and periods'),
});

export type RegistrationFormData = z.infer<typeof registrationSchema>;

/**
 * Availability check validation schema
 */
export const availabilitySchema = z.object({
  organization_name: z
    .string()
    .min(3, 'Organization name must be at least 3 characters')
    .max(100, 'Organization name must be at most 100 characters'),
  
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Please enter a valid email address'),
});

export type AvailabilityFormData = z.infer<typeof availabilitySchema>;
