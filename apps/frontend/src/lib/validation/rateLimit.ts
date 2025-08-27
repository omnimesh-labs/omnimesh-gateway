import * as z from 'zod';
import { commonValidation, enums } from './common';

// Rate Limit validation schema
export const rateLimitSchema = z.object({
  scope: z.enum(enums.rateLimitScope, {
    errorMap: () => ({ message: 'Please select a valid rate limit scope' })
  }),
  scope_id: z.string()
    .max(255, 'Scope ID must be less than 255 characters')
    .optional(),
  requests_per_minute: z.number()
    .int('Requests per minute must be a whole number')
    .min(1, 'Must allow at least 1 request per minute')
    .max(10000, 'Cannot exceed 10,000 requests per minute'),
  requests_per_hour: z.number()
    .int('Requests per hour must be a whole number')
    .min(1, 'Must allow at least 1 request per hour')
    .max(100000, 'Cannot exceed 100,000 requests per hour')
    .optional(),
  requests_per_day: z.number()
    .int('Requests per day must be a whole number')
    .min(1, 'Must allow at least 1 request per day')
    .max(1000000, 'Cannot exceed 1,000,000 requests per day')
    .optional(),
  burst_limit: z.number()
    .int('Burst limit must be a whole number')
    .min(1, 'Burst limit must be at least 1')
    .max(1000, 'Burst limit cannot exceed 1000')
    .optional(),
  is_active: commonValidation.isActive,
});

// Enhanced rate limit schema with scope and consistency validation
export const rateLimitSchemaWithValidation = rateLimitSchema
  .refine((data) => {
    // Validate scope_id requirement based on scope
    switch (data.scope) {
      case 'user':
      case 'api_key':
      case 'endpoint':
        return data.scope_id && data.scope_id.trim() !== '';
      case 'global':
      case 'organization':
        return true; // scope_id is optional/managed by system
      default:
        return true;
    }
  }, (data) => ({
    message: `Scope ID is required for ${data.scope} scope`,
    path: ['scope_id'],
  }))
  .refine((data) => {
    // Validate time period consistency (hour >= minute, day >= hour)
    if (data.requests_per_hour && data.requests_per_minute) {
      return data.requests_per_hour >= data.requests_per_minute;
    }
    return true;
  }, {
    message: 'Hourly limit must be greater than or equal to per-minute limit',
    path: ['requests_per_hour'],
  })
  .refine((data) => {
    // Validate daily vs hourly limits
    if (data.requests_per_day && data.requests_per_hour) {
      return data.requests_per_day >= data.requests_per_hour;
    }
    return true;
  }, {
    message: 'Daily limit must be greater than or equal to hourly limit',
    path: ['requests_per_day'],
  })
  .refine((data) => {
    // Validate burst limit vs per-minute limit
    if (data.burst_limit && data.requests_per_minute) {
      return data.burst_limit >= data.requests_per_minute;
    }
    return true;
  }, {
    message: 'Burst limit should be greater than or equal to per-minute limit',
    path: ['burst_limit'],
  });

// Rate limit scope configurations
export const scopeConfigurations = {
  global: {
    description: 'Applies to all requests across the entire system',
    requiresScopeId: false,
    recommended: {
      requests_per_minute: 1000,
      requests_per_hour: 50000,
      requests_per_day: 1000000,
      burst_limit: 100
    }
  },
  organization: {
    description: 'Applies to all requests within the organization',
    requiresScopeId: false,
    recommended: {
      requests_per_minute: 500,
      requests_per_hour: 25000,
      requests_per_day: 500000,
      burst_limit: 50
    }
  },
  user: {
    description: 'Applies to requests from a specific user',
    requiresScopeId: true,
    scopeIdLabel: 'User ID',
    recommended: {
      requests_per_minute: 100,
      requests_per_hour: 5000,
      requests_per_day: 100000,
      burst_limit: 10
    }
  },
  api_key: {
    description: 'Applies to requests using a specific API key',
    requiresScopeId: true,
    scopeIdLabel: 'API Key ID',
    recommended: {
      requests_per_minute: 200,
      requests_per_hour: 10000,
      requests_per_day: 200000,
      burst_limit: 20
    }
  },
  endpoint: {
    description: 'Applies to requests to a specific endpoint',
    requiresScopeId: true,
    scopeIdLabel: 'Endpoint Path',
    recommended: {
      requests_per_minute: 300,
      requests_per_hour: 15000,
      requests_per_day: 300000,
      burst_limit: 30
    }
  }
} as const;

// Validate scope ID format based on scope type
export const validateScopeId = (
  scopeId: string,
  scope: typeof enums.rateLimitScope[number]
): { valid: boolean; error?: string } => {
  if (!scopeConfigurations[scope].requiresScopeId) {
    return { valid: true };
  }

  if (!scopeId || scopeId.trim() === '') {
    return {
      valid: false,
      error: `${scopeConfigurations[scope].scopeIdLabel} is required for ${scope} scope`
    };
  }

  switch (scope) {
    case 'user':
      // UUID format for user ID
      if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(scopeId)) {
        return { valid: false, error: 'User ID must be a valid UUID format' };
      }
      break;

    case 'api_key':
      // API key format validation
      if (!/^[a-zA-Z0-9_-]{20,}$/.test(scopeId)) {
        return { valid: false, error: 'API Key ID must be at least 20 characters with valid format' };
      }
      break;

    case 'endpoint':
      // Endpoint path validation
      if (!scopeId.startsWith('/')) {
        return { valid: false, error: 'Endpoint path must start with /' };
      }
      if (!/^\/[a-zA-Z0-9/_-]*$/.test(scopeId)) {
        return { valid: false, error: 'Endpoint path contains invalid characters' };
      }
      break;
  }

  return { valid: true };
};

// Calculate derived limits
export const calculateDerivedLimits = (data: Partial<z.infer<typeof rateLimitSchema>>) => {
  const result = { ...data };

  // Auto-calculate hourly limit if not provided
  if (!result.requests_per_hour && result.requests_per_minute) {
    result.requests_per_hour = result.requests_per_minute * 60;
  }

  // Auto-calculate daily limit if not provided
  if (!result.requests_per_day && result.requests_per_hour) {
    result.requests_per_day = result.requests_per_hour * 24;
  }

  // Auto-calculate burst limit if not provided
  if (!result.burst_limit && result.requests_per_minute) {
    result.burst_limit = Math.max(1, Math.floor(result.requests_per_minute * 0.1));
  }

  return result;
};

// Get recommended limits for scope
export const getRecommendedLimits = (scope: typeof enums.rateLimitScope[number]) => {
  return scopeConfigurations[scope].recommended;
};

// Validate rate limit effectiveness
export const validateRateLimitEffectiveness = (
  data: z.infer<typeof rateLimitSchema>
): { warnings: string[]; recommendations: string[] } => {
  const warnings: string[] = [];
  const recommendations: string[] = [];

  // Check if limits are too permissive
  if (data.requests_per_minute > 500) {
    warnings.push('Very high per-minute limit may not provide effective rate limiting');
  }

  if (data.burst_limit && data.burst_limit > data.requests_per_minute * 0.5) {
    warnings.push('Burst limit is more than 50% of per-minute limit');
  }

  // Check if limits are too restrictive
  if (data.requests_per_minute < 5 && data.scope === 'user') {
    warnings.push('Very low per-minute limit may impact user experience');
  }

  // Recommendations
  const recommended = scopeConfigurations[data.scope].recommended;

  if (data.requests_per_minute > recommended.requests_per_minute * 2) {
    recommendations.push(`Consider reducing per-minute limit (recommended: ${recommended.requests_per_minute})`);
  }

  if (data.requests_per_minute < recommended.requests_per_minute * 0.1) {
    recommendations.push(`Consider increasing per-minute limit (recommended: ${recommended.requests_per_minute})`);
  }

  if (!data.burst_limit) {
    recommendations.push('Consider setting a burst limit for better traffic handling');
  }

  return { warnings, recommendations };
};

// Rate limit usage schema (for monitoring)
export const rateLimitUsageSchema = z.object({
  rate_limit_id: commonValidation.requiredUuid,
  identifier: z.string().min(1, 'Identifier is required'),
  window_start: z.date(),
  request_count: z.number().int().min(0),
  expires_at: z.date(),
});

// Rate limit creation schema
export const createRateLimitSchema = rateLimitSchemaWithValidation.omit({});

// Rate limit update schema
export const updateRateLimitSchema = rateLimitSchemaWithValidation.partial();

export type RateLimitFormData = z.infer<typeof rateLimitSchema>;
export type CreateRateLimitData = z.infer<typeof createRateLimitSchema>;
export type UpdateRateLimitData = z.infer<typeof updateRateLimitSchema>;
export type RateLimitUsageData = z.infer<typeof rateLimitUsageSchema>;
