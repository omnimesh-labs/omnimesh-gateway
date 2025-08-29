import * as z from 'zod';
import { commonValidation, enums } from './common';

// User validation schema
export const userSchema = z.object({
  email: commonValidation.email,
  name: z.string()
    .min(1, 'Name is required')
    .max(255, 'Name must be less than 255 characters')
    .trim(),
  role: z.enum(enums.userRole, {
    errorMap: () => ({ message: 'Please select a valid user role' })
  }).default('user'),
  is_active: commonValidation.isActive,
  email_verified: z.boolean().default(false),
});

// Password validation schema
export const passwordSchema = z.object({
  password: z.string()
    .min(8, 'Password must be at least 8 characters')
    .max(128, 'Password must be less than 128 characters')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number')
    .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character'),
  confirmPassword: z.string()
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword']
});

// User creation schema (includes password)
export const createUserSchema = userSchema.merge(passwordSchema).omit({ confirmPassword: true });

// User registration schema (includes password confirmation)
export const registerUserSchema = userSchema.merge(passwordSchema);

// User update schema (excludes password, allows partial updates)
export const updateUserSchema = userSchema.partial();

// Password change schema
export const changePasswordSchema = z.object({
  currentPassword: z.string().min(1, 'Current password is required'),
  newPassword: z.string()
    .min(8, 'Password must be at least 8 characters')
    .max(128, 'Password must be less than 128 characters')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number')
    .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character'),
  confirmNewPassword: z.string()
}).refine((data) => data.newPassword === data.confirmNewPassword, {
  message: 'New passwords do not match',
  path: ['confirmNewPassword']
}).refine((data) => data.currentPassword !== data.newPassword, {
  message: 'New password must be different from current password',
  path: ['newPassword']
});

// Profile update schema (user can update their own info)
export const updateProfileSchema = z.object({
  name: z.string()
    .min(1, 'Name is required')
    .max(255, 'Name must be less than 255 characters')
    .trim(),
  email: commonValidation.email,
});

// Role validation helpers
export const rolePermissions = {
  admin: {
    canManageUsers: true,
    canManageOrganization: true,
    canManageServers: true,
    canViewLogs: true,
    canManageSettings: true,
    description: 'Full administrative access to all features'
  },
  user: {
    canManageUsers: false,
    canManageOrganization: false,
    canManageServers: true,
    canViewLogs: false,
    canManageSettings: false,
    description: 'Can manage servers and use MCP features'
  },
  viewer: {
    canManageUsers: false,
    canManageOrganization: false,
    canManageServers: false,
    canViewLogs: false,
    canManageSettings: false,
    description: 'Read-only access to view servers and sessions'
  },
  api_user: {
    canManageUsers: false,
    canManageOrganization: false,
    canManageServers: true,
    canViewLogs: false,
    canManageSettings: false,
    description: 'API access for programmatic server management'
  }
} as const;

// Validate role assignment
export const validateRoleAssignment = (
  assignerRole: typeof enums.userRole[number],
  targetRole: typeof enums.userRole[number]
): { valid: boolean; error?: string } => {
  // Only admins can assign admin role
  if (targetRole === 'admin' && assignerRole !== 'admin') {
    return { valid: false, error: 'Only administrators can assign admin role' };
  }

  // Users cannot assign roles
  if (assignerRole === 'user' || assignerRole === 'viewer' || assignerRole === 'api_user') {
    return { valid: false, error: 'Insufficient permissions to assign roles' };
  }

  return { valid: true };
};

// Password strength validation
export const validatePasswordStrength = (password: string): {
  score: number;
  feedback: string[];
  strength: 'weak' | 'fair' | 'good' | 'strong';
} => {
  let score = 0;
  const feedback: string[] = [];

  // Length check
  if (password.length >= 8) score += 1;
  else feedback.push('Use at least 8 characters');

  if (password.length >= 12) score += 1;
  else if (password.length >= 8) feedback.push('Use 12+ characters for better security');

  // Character variety
  if (/[a-z]/.test(password)) score += 1;
  else feedback.push('Add lowercase letters');

  if (/[A-Z]/.test(password)) score += 1;
  else feedback.push('Add uppercase letters');

  if (/[0-9]/.test(password)) score += 1;
  else feedback.push('Add numbers');

  if (/[^A-Za-z0-9]/.test(password)) score += 1;
  else feedback.push('Add special characters (!@#$%^&*)');

  // Additional strength checks
  if (password.length >= 16) score += 1;
  if (/[^A-Za-z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) score += 1;

  // Common patterns (reduce score)
  if (/123456|abcdef|qwerty|password/i.test(password)) {
    score -= 2;
    feedback.push('Avoid common patterns and dictionary words');
  }

  const strength = score <= 2 ? 'weak' : score <= 4 ? 'fair' : score <= 6 ? 'good' : 'strong';

  return {
    score: Math.max(0, score),
    feedback,
    strength
  };
};

// Email validation beyond basic format
export const validateEmailDomain = (email: string): { valid: boolean; error?: string } => {
  const domain = email.split('@')[1];

  if (!domain) {
    return { valid: false, error: 'Invalid email format' };
  }

  // Check for common typos in popular domains
  const commonDomainTypos = {
    'gmail.co': 'gmail.com',
    'gmail.cm': 'gmail.com',
    'gmial.com': 'gmail.com',
    'yahooo.com': 'yahoo.com',
    'hotmial.com': 'hotmail.com',
    'outlok.com': 'outlook.com'
  };

  if (commonDomainTypos[domain as keyof typeof commonDomainTypos]) {
    return {
      valid: false,
      error: `Did you mean ${commonDomainTypos[domain as keyof typeof commonDomainTypos]}?`
    };
  }

  // Basic domain format validation - allows subdomains like co.uk
  if (!/^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*\.([a-zA-Z]{2,}|xn--[a-zA-Z0-9]+)$/.test(domain)) {
    return { valid: false, error: 'Invalid email domain format' };
  }

  return { valid: true };
};

// Generate username suggestion from email
export const generateUsernameFromEmail = (email: string): string => {
  const username = email.split('@')[0];
  return username.replace(/[^a-zA-Z0-9_]/g, '').toLowerCase();
};

// Login schema
export const loginSchema = z.object({
  email: commonValidation.email,
  password: z.string().min(1, 'Password is required'),
  remember: z.boolean().default(false),
});

// Password reset request schema
export const passwordResetRequestSchema = z.object({
  email: commonValidation.email,
});

// Password reset schema
export const passwordResetSchema = z.object({
  token: z.string().min(1, 'Reset token is required'),
  password: z.string()
    .min(8, 'Password must be at least 8 characters')
    .max(128, 'Password must be less than 128 characters')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number')
    .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character'),
  confirmPassword: z.string()
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword']
});

export type UserFormData = z.infer<typeof userSchema>;
export type CreateUserData = z.infer<typeof createUserSchema>;
export type RegisterUserData = z.infer<typeof registerUserSchema>;
export type UpdateUserData = z.infer<typeof updateUserSchema>;
export type UpdateProfileData = z.infer<typeof updateProfileSchema>;
export type ChangePasswordData = z.infer<typeof changePasswordSchema>;
export type LoginData = z.infer<typeof loginSchema>;
export type PasswordResetRequestData = z.infer<typeof passwordResetRequestSchema>;
export type PasswordResetData = z.infer<typeof passwordResetSchema>;
