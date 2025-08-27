import * as z from 'zod';
import { commonValidation, enums } from './common';

// MCP Server validation schema
export const serverSchema = z.object({
  name: commonValidation.name,
  description: commonValidation.description,
  protocol: z.enum(enums.protocol, {
    errorMap: () => ({ message: 'Please select a valid protocol' })
  }),
  url: z.string().optional(),
  command: z.string().max(255, 'Command must be less than 255 characters').optional(),
  args: z.array(z.string().max(100, 'Each argument must be less than 100 characters'))
    .max(20, 'Maximum 20 arguments allowed')
    .default([]),
  environment: z.array(z.string().max(200, 'Each environment variable must be less than 200 characters'))
    .max(50, 'Maximum 50 environment variables allowed')
    .default([]),
  working_dir: z.string()
    .max(500, 'Working directory path must be less than 500 characters')
    .optional(),
  version: z.string()
    .max(50, 'Version must be less than 50 characters')
    .optional(),
  timeout_seconds: commonValidation.timeout.default(300),
  max_retries: commonValidation.retries.default(3),
  status: z.enum(enums.serverStatus, {
    errorMap: () => ({ message: 'Please select a valid server status' })
  }).default('active'),
  health_check_url: commonValidation.optionalUrl,
  is_active: commonValidation.isActive,
  metadata: commonValidation.jsonString,
  tags: commonValidation.tags,
});

// Protocol-specific validation
const protocolValidation = {
  http: z.object({
    url: z.string().url('Valid HTTP URL is required').min(1, 'URL is required for HTTP protocol'),
    command: z.string().optional(),
  }),
  https: z.object({
    url: z.string().url('Valid HTTPS URL is required').min(1, 'URL is required for HTTPS protocol'),
    command: z.string().optional(),
  }),
  stdio: z.object({
    command: z.string().min(1, 'Command is required for STDIO protocol'),
    url: z.string().optional(),
  }),
  websocket: z.object({
    url: z.string()
      .refine((url) => url.startsWith('ws://') || url.startsWith('wss://'),
        'WebSocket URL must start with ws:// or wss://')
      .min(1, 'WebSocket URL is required'),
    command: z.string().optional(),
  }),
  sse: z.object({
    url: z.string().url('Valid URL is required for SSE protocol').min(1, 'URL is required for SSE protocol'),
    command: z.string().optional(),
  }),
};

// Enhanced server schema with protocol-specific validation
export const serverSchemaWithProtocolValidation = serverSchema
  .refine((data) => {
    switch (data.protocol) {
      case 'http':
      case 'https':
      case 'sse':
        return data.url && data.url.trim() !== '';
      case 'stdio':
        return data.command && data.command.trim() !== '';
      case 'websocket':
        return data.url && (data.url.startsWith('ws://') || data.url.startsWith('wss://'));
      default:
        return true;
    }
  }, (data) => {
    switch (data.protocol) {
      case 'http':
      case 'https':
      case 'sse':
        return {
          message: `URL is required for ${data.protocol.toUpperCase()} protocol`,
          path: ['url'],
        };
      case 'stdio':
        return {
          message: 'Command is required for STDIO protocol',
          path: ['command'],
        };
      case 'websocket':
        return {
          message: 'Valid WebSocket URL (ws:// or wss://) is required',
          path: ['url'],
        };
      default:
        return {
          message: 'Invalid protocol configuration',
          path: ['protocol'],
        };
    }
  })
  .refine((data) => {
    // Validate that health check URL matches protocol if provided
    if (data.health_check_url) {
      switch (data.protocol) {
        case 'http':
          return data.health_check_url.startsWith('http://');
        case 'https':
          return data.health_check_url.startsWith('https://');
        case 'websocket':
          return data.health_check_url.startsWith('http://') || data.health_check_url.startsWith('https://');
        case 'sse':
          return data.health_check_url.startsWith('http://') || data.health_check_url.startsWith('https://');
        default:
          return true;
      }
    }
    return true;
  }, {
    message: 'Health check URL protocol should match or be compatible with server protocol',
    path: ['health_check_url'],
  });

// Environment variable validation helper
export const validateEnvironmentVariable = (envVar: string): { valid: boolean; error?: string } => {
  if (!envVar.includes('=')) {
    return { valid: false, error: 'Environment variable must be in format KEY=VALUE' };
  }

  const [key, ...valueParts] = envVar.split('=');
  const value = valueParts.join('=');

  if (!key || key.trim() === '') {
    return { valid: false, error: 'Environment variable key cannot be empty' };
  }

  if (!/^[A-Z_][A-Z0-9_]*$/i.test(key)) {
    return { valid: false, error: 'Environment variable key can only contain letters, numbers, and underscores' };
  }

  if (value === undefined) {
    return { valid: false, error: 'Environment variable value cannot be empty' };
  }

  return { valid: true };
};

// Command argument validation helper
export const validateCommandArgument = (arg: string): { valid: boolean; error?: string } => {
  if (arg.trim() === '') {
    return { valid: false, error: 'Command argument cannot be empty' };
  }

  if (arg.length > 100) {
    return { valid: false, error: 'Command argument must be less than 100 characters' };
  }

  return { valid: true };
};

// Server creation schema (excludes system-generated fields)
export const createServerSchema = serverSchemaWithProtocolValidation.omit({});

// Server update schema (allows partial updates)
export const updateServerSchema = serverSchemaWithProtocolValidation.partial();

export type ServerFormData = z.infer<typeof serverSchema>;
export type CreateServerData = z.infer<typeof createServerSchema>;
export type UpdateServerData = z.infer<typeof updateServerSchema>;
