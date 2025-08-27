import * as z from 'zod';
import { commonValidation, enums } from './common';

// Organization validation schema
export const organizationSchema = z.object({
  name: commonValidation.name,
  slug: commonValidation.slug,
  plan_type: z.enum(enums.planType, {
    errorMap: () => ({ message: 'Please select a valid plan type' })
  }).default('free'),
  max_servers: z.number()
    .int('Max servers must be a whole number')
    .min(1, 'Must allow at least 1 server')
    .max(10000, 'Cannot exceed 10,000 servers')
    .default(10),
  max_sessions: z.number()
    .int('Max sessions must be a whole number')
    .min(1, 'Must allow at least 1 session')
    .max(100000, 'Cannot exceed 100,000 sessions')
    .default(100),
  log_retention_days: z.number()
    .int('Log retention must be a whole number of days')
    .min(1, 'Must retain logs for at least 1 day')
    .max(365, 'Cannot retain logs for more than 365 days')
    .default(7),
  is_active: commonValidation.isActive,
});

// Organization creation schema (excludes system-generated fields)
export const createOrganizationSchema = organizationSchema.omit({});

// Organization update schema (allows partial updates)
export const updateOrganizationSchema = organizationSchema.partial();

// Plan limits configuration
export const planLimits = {
  free: {
    max_servers: 5,
    max_sessions: 50,
    log_retention_days: 7,
  },
  pro: {
    max_servers: 50,
    max_sessions: 500,
    log_retention_days: 30,
  },
  enterprise: {
    max_servers: 1000,
    max_sessions: 10000,
    log_retention_days: 365,
  },
} as const;

// Validation function to check plan limits
export const validatePlanLimits = (data: z.infer<typeof organizationSchema>) => {
  const limits = planLimits[data.plan_type];
  const errors: string[] = [];

  if (data.max_servers > limits.max_servers) {
    errors.push(`${data.plan_type} plan allows maximum ${limits.max_servers} servers`);
  }

  if (data.max_sessions > limits.max_sessions) {
    errors.push(`${data.plan_type} plan allows maximum ${limits.max_sessions} sessions`);
  }

  if (data.log_retention_days > limits.log_retention_days) {
    errors.push(`${data.plan_type} plan allows maximum ${limits.log_retention_days} days log retention`);
  }

  return {
    valid: errors.length === 0,
    errors,
  };
};

// Enhanced organization schema with plan limit validation
export const organizationSchemaWithPlanValidation = organizationSchema.refine(
  (data) => {
    const validation = validatePlanLimits(data);
    return validation.valid;
  },
  (data) => {
    const validation = validatePlanLimits(data);
    return {
      message: validation.errors[0] || 'Plan limits exceeded',
      path: ['plan_type'],
    };
  }
);

export type OrganizationFormData = z.infer<typeof organizationSchema>;
export type CreateOrganizationData = z.infer<typeof createOrganizationSchema>;
export type UpdateOrganizationData = z.infer<typeof updateOrganizationSchema>;
