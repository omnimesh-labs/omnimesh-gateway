import * as z from 'zod';
import { commonValidation, enums } from './common';

// MCP Resource validation schema
export const resourceSchema = z.object({
  name: commonValidation.name,
  description: commonValidation.description,
  resource_type: z.enum(enums.resourceType, {
    errorMap: () => ({ message: 'Please select a valid resource type' })
  }),
  uri: commonValidation.uri,
  mime_type: commonValidation.mimeType,
  size_bytes: commonValidation.fileSize,
  is_active: commonValidation.isActive,
  tags: commonValidation.tags,
  access_permissions: commonValidation.jsonString,
  metadata: commonValidation.jsonString,
});

// Resource type specific validation patterns
const resourceTypePatterns = {
  file: {
    uriPattern: /^(\/|\.\/|file:\/\/)/,
    description: 'File path must start with /, ./, or file://',
    allowedMimeTypes: [
      'text/plain', 'text/csv', 'application/json', 'application/xml',
      'image/jpeg', 'image/png', 'image/gif', 'application/pdf',
      'application/zip', 'application/octet-stream'
    ]
  },
  url: {
    uriPattern: /^https?:\/\//,
    description: 'URL must start with http:// or https://',
    allowedMimeTypes: [
      'text/html', 'application/json', 'application/xml', 'text/plain',
      'image/jpeg', 'image/png', 'application/pdf'
    ]
  },
  database: {
    uriPattern: /^(postgresql|mysql|sqlite|mongodb):\/\//,
    description: 'Database URI must specify a valid protocol (postgresql, mysql, sqlite, mongodb)',
    allowedMimeTypes: ['application/sql', 'application/json']
  },
  api: {
    uriPattern: /^https?:\/\/.*\/api/i,
    description: 'API URI should be a valid HTTP URL containing /api',
    allowedMimeTypes: ['application/json', 'application/xml', 'text/plain']
  },
  memory: {
    uriPattern: /^(memory|cache|redis):\/\//,
    description: 'Memory URI must specify a valid protocol (memory, cache, redis)',
    allowedMimeTypes: ['application/json', 'text/plain']
  },
  custom: {
    uriPattern: /.+/,  // Allow any URI for custom resources
    description: 'Custom resource URI can be any valid format',
    allowedMimeTypes: [] // Allow any MIME type for custom resources
  }
};

// Enhanced resource schema with type-specific validation
export const resourceSchemaWithTypeValidation = resourceSchema
  .refine((data) => {
    const typePattern = resourceTypePatterns[data.resource_type];
    return typePattern.uriPattern.test(data.uri);
  }, (data) => ({
    message: resourceTypePatterns[data.resource_type].description,
    path: ['uri'],
  }))
  .refine((data) => {
    if (!data.mime_type) return true; // MIME type is optional

    const typeConfig = resourceTypePatterns[data.resource_type];
    if (typeConfig.allowedMimeTypes.length === 0) return true; // Custom type allows anything

    return typeConfig.allowedMimeTypes.includes(data.mime_type);
  }, (data) => ({
    message: `MIME type not typically used for ${data.resource_type} resources`,
    path: ['mime_type'],
  }))
  .refine((data) => {
    // Validate file size constraints based on resource type
    if (!data.size_bytes) return true;

    switch (data.resource_type) {
      case 'file':
        return data.size_bytes <= 100 * 1024 * 1024; // 100MB for files
      case 'url':
        return data.size_bytes <= 10 * 1024 * 1024;  // 10MB for URLs (cached content)
      case 'database':
        return true; // No size limit for database resources
      case 'api':
        return data.size_bytes <= 10 * 1024 * 1024;  // 10MB for API responses
      case 'memory':
        return data.size_bytes <= 1024 * 1024;       // 1MB for memory resources
      case 'custom':
        return true; // No specific limit for custom resources
    }
    return true;
  }, (data) => {
    const limits = {
      file: '100MB',
      url: '10MB',
      api: '10MB',
      memory: '1MB',
    };
    return {
      message: `File size exceeds recommended limit for ${data.resource_type} resources (${limits[data.resource_type as keyof typeof limits] || 'unknown'})`,
      path: ['size_bytes'],
    };
  });

// Access permissions validation helper
export const validateAccessPermissions = (permissionsString: string): { valid: boolean; error?: string } => {
  if (!permissionsString || permissionsString.trim() === '') return { valid: true };

  try {
    const permissions = JSON.parse(permissionsString);

    if (typeof permissions !== 'object' || permissions === null) {
      return { valid: false, error: 'Access permissions must be an object' };
    }

    // Common permission fields
    const allowedFields = ['read', 'write', 'delete', 'admin', 'users', 'roles', 'groups'];
    const providedFields = Object.keys(permissions);

    for (const field of providedFields) {
      if (!allowedFields.includes(field)) {
        console.warn(`Unknown permission field: ${field}`);
      }
    }

    return { valid: true };
  } catch {
    return { valid: false, error: 'Invalid JSON format for access permissions' };
  }
};

// URI validation helper based on resource type
export const validateUriForType = (uri: string, resourceType: typeof enums.resourceType[number]): { valid: boolean; error?: string } => {
  const typeConfig = resourceTypePatterns[resourceType];

  if (!typeConfig.uriPattern.test(uri)) {
    return {
      valid: false,
      error: typeConfig.description
    };
  }

  // Additional specific validations
  switch (resourceType) {
    case 'url':
    case 'api':
      try {
        new URL(uri); // Validate as proper URL
      } catch {
        return { valid: false, error: 'Invalid URL format' };
      }
      break;
    case 'file':
      if (uri.includes('..')) {
        return { valid: false, error: 'File paths cannot contain parent directory references (..)' };
      }
      break;
    case 'database':
      if (!uri.includes('://')) {
        return { valid: false, error: 'Database URI must include protocol' };
      }
      break;
  }

  return { valid: true };
};

// Suggest MIME type based on URI
export const suggestMimeType = (uri: string, resourceType: typeof enums.resourceType[number]): string | null => {
  const extension = uri.split('.').pop()?.toLowerCase();

  const mimeTypeMap: Record<string, string> = {
    // Text files
    'txt': 'text/plain',
    'csv': 'text/csv',
    'json': 'application/json',
    'xml': 'application/xml',
    'html': 'text/html',

    // Images
    'jpg': 'image/jpeg',
    'jpeg': 'image/jpeg',
    'png': 'image/png',
    'gif': 'image/gif',

    // Documents
    'pdf': 'application/pdf',
    'zip': 'application/zip',

    // Database
    'sql': 'application/sql',
    'db': 'application/octet-stream',
  };

  if (extension && mimeTypeMap[extension]) {
    return mimeTypeMap[extension];
  }

  // Default MIME types by resource type
  switch (resourceType) {
    case 'api':
      return 'application/json';
    case 'database':
      return 'application/json';
    case 'memory':
      return 'application/json';
    default:
      return null;
  }
};

// Resource creation schema
export const createResourceSchema = resourceSchemaWithTypeValidation.omit({});

// Resource update schema
export const updateResourceSchema = resourceSchemaWithTypeValidation.partial();

export type ResourceFormData = z.infer<typeof resourceSchema>;
export type CreateResourceData = z.infer<typeof createResourceSchema>;
export type UpdateResourceData = z.infer<typeof updateResourceSchema>;
