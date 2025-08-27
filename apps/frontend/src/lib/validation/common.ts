import * as z from 'zod';

// Common validation patterns and utilities
export const commonValidation = {
  // UUID validation
  uuid: z.string().uuid('Invalid UUID format'),

  // Required UUID (non-nullable)
  requiredUuid: z.string().uuid('Invalid UUID format'),

  // Optional UUID (nullable)
  optionalUuid: z.string().uuid('Invalid UUID format').optional(),

  // Name validation (common across many models)
  name: z.string()
    .min(1, 'Name is required')
    .max(255, 'Name must be less than 255 characters')
    .trim(),

  // Optional description
  description: z.string()
    .max(1000, 'Description must be less than 1000 characters')
    .trim()
    .optional()
    .or(z.literal('')),

  // Email validation
  email: z.string()
    .email('Invalid email address')
    .max(255, 'Email must be less than 255 characters')
    .toLowerCase(),

  // URL validation
  url: z.string()
    .url('Invalid URL format')
    .max(500, 'URL must be less than 500 characters'),

  // Optional URL
  optionalUrl: z.string()
    .url('Invalid URL format')
    .max(500, 'URL must be less than 500 characters')
    .optional()
    .or(z.literal('')),

  // URI validation (more flexible than URL)
  uri: z.string()
    .min(1, 'URI is required')
    .max(500, 'URI must be less than 500 characters')
    .refine((uri) => {
      // Allow various URI schemes: http, https, file, ftp, etc.
      return /^[a-z][a-z0-9+.-]*:/i.test(uri) || uri.startsWith('/') || uri.startsWith('./');
    }, 'Invalid URI format'),

  // Tags array validation
  tags: z.array(z.string().trim().min(1).max(50))
    .max(20, 'Maximum 20 tags allowed')
    .default([]),

  // JSON string validation
  jsonString: z.string()
    .optional()
    .refine((val) => {
      if (!val || val.trim() === '') return true;
      try {
        JSON.parse(val);
        return true;
      } catch {
        return false;
      }
    }, 'Invalid JSON format'),

  // Required JSON string validation
  requiredJsonString: z.string()
    .min(1, 'JSON is required')
    .refine((val) => {
      try {
        JSON.parse(val);
        return true;
      } catch {
        return false;
      }
    }, 'Invalid JSON format'),

  // Slug validation (URL-safe identifier)
  slug: z.string()
    .min(1, 'Slug is required')
    .max(100, 'Slug must be less than 100 characters')
    .regex(/^[a-z0-9-_]+$/, 'Slug can only contain lowercase letters, numbers, hyphens, and underscores')
    .refine((slug) => !slug.startsWith('-') && !slug.endsWith('-'), 'Slug cannot start or end with a hyphen'),

  // Function name validation (for tools)
  functionName: z.string()
    .min(1, 'Function name is required')
    .max(100, 'Function name must be less than 100 characters')
    .regex(/^[a-zA-Z_][a-zA-Z0-9_]*$/, 'Function name must be a valid identifier (letters, numbers, underscores only, cannot start with number)'),

  // MIME type validation
  mimeType: z.string()
    .regex(/^[a-z]+\/[a-z0-9\-\+\.]+$/i, 'Invalid MIME type format')
    .max(100, 'MIME type must be less than 100 characters')
    .optional(),

  // File size validation (in bytes)
  fileSize: z.number()
    .int('File size must be a whole number')
    .min(0, 'File size cannot be negative')
    .max(1024 * 1024 * 1024, 'File size cannot exceed 1GB') // 1GB limit
    .optional(),

  // Timeout validation (in seconds)
  timeout: z.number()
    .int('Timeout must be a whole number')
    .min(1, 'Timeout must be at least 1 second')
    .max(300, 'Timeout cannot exceed 300 seconds'),

  // Retry count validation
  retries: z.number()
    .int('Retry count must be a whole number')
    .min(0, 'Retry count cannot be negative')
    .max(10, 'Retry count cannot exceed 10'),

  // Rate limit validation
  rateLimit: z.number()
    .int('Rate limit must be a whole number')
    .min(1, 'Rate limit must be at least 1 request'),

  // Boolean with default
  booleanDefault: (defaultValue: boolean) => z.boolean().default(defaultValue),

  // Active status
  isActive: z.boolean().default(true),

  // Public status
  isPublic: z.boolean().default(false),
};

// Utility functions for validation
export const validationUtils = {
  // Convert string to slug format
  slugify: (str: string): string => {
    return str
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9\s-]/g, '') // Remove special characters
      .replace(/\s+/g, '-') // Replace spaces with hyphens
      .replace(/-+/g, '-') // Replace multiple hyphens with single
      .replace(/^-|-$/g, ''); // Remove leading/trailing hyphens
  },

  // Validate and parse JSON with better error messages
  parseJson: <T>(jsonString: string): { success: true; data: T } | { success: false; error: string } => {
    try {
      const parsed = JSON.parse(jsonString);
      return { success: true, data: parsed };
    } catch (error) {
      const err = error as Error;
      return {
        success: false,
        error: `Invalid JSON: ${err.message}`
      };
    }
  },

  // Validate JSON schema structure for tools
  validateToolSchema: (schemaString: string): { valid: boolean; error?: string } => {
    try {
      const schema = JSON.parse(schemaString);

      // Basic JSON Schema validation
      if (typeof schema !== 'object' || schema === null) {
        return { valid: false, error: 'Schema must be an object' };
      }

      if (!schema.type) {
        return { valid: false, error: 'Schema must have a "type" property' };
      }

      return { valid: true };
    } catch {
      return { valid: false, error: 'Invalid JSON format' };
    }
  },

  // Validate prompt parameters
  validatePromptParameters: (parametersString: string): { valid: boolean; error?: string } => {
    try {
      const parameters = JSON.parse(parametersString);

      if (!Array.isArray(parameters)) {
        return { valid: false, error: 'Parameters must be an array' };
      }

      for (const param of parameters) {
        if (typeof param !== 'object' || !param.name) {
          return { valid: false, error: 'Each parameter must be an object with a "name" property' };
        }
      }

      return { valid: true };
    } catch {
      return { valid: false, error: 'Invalid JSON format' };
    }
  },

  // Generate validation error message
  getValidationError: (errors: any[]): string => {
    if (errors.length === 0) return '';
    return errors[0].message || 'Validation error';
  },
};

// Enum definitions that mirror backend enums
export const enums = {
  planType: ['free', 'pro', 'enterprise'] as const,
  protocol: ['http', 'https', 'stdio', 'websocket', 'sse'] as const,
  serverStatus: ['active', 'inactive', 'error', 'pending'] as const,
  sessionStatus: ['initializing', 'active', 'closed', 'error'] as const,
  processStatus: ['running', 'stopped', 'error'] as const,
  userRole: ['admin', 'user', 'viewer', 'api_user'] as const,
  resourceType: ['file', 'url', 'database', 'api', 'memory', 'custom'] as const,
  toolCategory: ['general', 'data', 'file', 'web', 'system', 'ai', 'dev', 'custom'] as const,
  toolImplementationType: ['internal', 'external', 'webhook', 'script'] as const,
  promptCategory: ['general', 'coding', 'analysis', 'creative', 'educational', 'business', 'custom'] as const,
  adapterType: ['REST', 'GraphQL', 'gRPC', 'SOAP'] as const,
  logLevel: ['debug', 'info', 'warn', 'error'] as const,
  windowType: ['minute', 'hour', 'day'] as const,
  rateLimitScope: ['global', 'organization', 'user', 'api_key', 'endpoint'] as const,
  auditAction: ['create', 'update', 'delete', 'read'] as const,
  importStatus: ['pending', 'in_progress', 'completed', 'failed'] as const,
  conflictStrategy: ['skip', 'overwrite', 'merge'] as const,
} as const;

export type PlanType = typeof enums.planType[number];
export type Protocol = typeof enums.protocol[number];
export type ServerStatus = typeof enums.serverStatus[number];
export type SessionStatus = typeof enums.sessionStatus[number];
export type ProcessStatus = typeof enums.processStatus[number];
export type UserRole = typeof enums.userRole[number];
export type ResourceType = typeof enums.resourceType[number];
export type ToolCategory = typeof enums.toolCategory[number];
export type ToolImplementationType = typeof enums.toolImplementationType[number];
export type PromptCategory = typeof enums.promptCategory[number];
export type AdapterType = typeof enums.adapterType[number];
export type LogLevel = typeof enums.logLevel[number];
export type WindowType = typeof enums.windowType[number];
export type RateLimitScope = typeof enums.rateLimitScope[number];
export type AuditAction = typeof enums.auditAction[number];
export type ImportStatus = typeof enums.importStatus[number];
export type ConflictStrategy = typeof enums.conflictStrategy[number];
