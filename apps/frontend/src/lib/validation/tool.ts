import * as z from 'zod';
import { commonValidation, enums, validationUtils } from './common';

// MCP Tool validation schema
export const toolSchema = z.object({
  name: commonValidation.name,
  description: commonValidation.description,
  function_name: commonValidation.functionName,
  category: z.enum(enums.toolCategory, {
    errorMap: () => ({ message: 'Please select a valid tool category' })
  }),
  implementation_type: z.enum(enums.toolImplementationType, {
    errorMap: () => ({ message: 'Please select a valid implementation type' })
  }).default('internal'),
  endpoint_url: z.string().optional(),
  timeout_seconds: commonValidation.timeout.default(30),
  max_retries: commonValidation.retries.default(3),
  usage_count: z.number().int().min(0).default(0),
  is_active: commonValidation.isActive,
  is_public: commonValidation.isPublic,
  schema: commonValidation.jsonString,
  examples: commonValidation.jsonString,
  documentation: z.string()
    .max(5000, 'Documentation must be less than 5000 characters')
    .optional(),
  tags: commonValidation.tags,
  access_permissions: commonValidation.jsonString,
  metadata: commonValidation.jsonString,
});

// Implementation type specific validation
export const toolSchemaWithImplementationValidation = toolSchema
  .refine((data) => {
    // Validate endpoint URL based on implementation type
    switch (data.implementation_type) {
      case 'external':
      case 'webhook':
        return data.endpoint_url && data.endpoint_url.trim() !== '';
      case 'internal':
      case 'script':
        return true; // No endpoint URL required
      default:
        return true;
    }
  }, (data) => ({
    message: `Endpoint URL is required for ${data.implementation_type} implementation`,
    path: ['endpoint_url'],
  }))
  .refine((data) => {
    // Validate endpoint URL format
    if (data.endpoint_url && data.endpoint_url.trim() !== '') {
      try {
        new URL(data.endpoint_url);
        return true;
      } catch {
        return false;
      }
    }
    return true;
  }, {
    message: 'Endpoint URL must be a valid HTTP/HTTPS URL',
    path: ['endpoint_url'],
  })
  .refine((data) => {
    // Validate webhook URLs use HTTPS in production
    if (data.implementation_type === 'webhook' && data.endpoint_url) {
      const url = new URL(data.endpoint_url);
      return url.protocol === 'https:' || url.hostname === 'localhost' || url.hostname.startsWith('127.');
    }
    return true;
  }, {
    message: 'Webhook URLs should use HTTPS for security',
    path: ['endpoint_url'],
  });

// JSON Schema validation for tool schema field
export const validateToolSchema = (schemaString: string): { valid: boolean; error?: string; suggestions?: string[] } => {
  if (!schemaString || schemaString.trim() === '') {
    return { valid: true }; // Schema is optional
  }

  const parseResult = validationUtils.parseJson(schemaString);
  if (!parseResult.success) {
    return { valid: false, error: parseResult.error };
  }

  const schema = parseResult.data;

  // Basic JSON Schema validation
  if (typeof schema !== 'object' || schema === null) {
    return { valid: false, error: 'Schema must be an object' };
  }

  // Check for required JSON Schema fields
  if (!schema.type) {
    return {
      valid: false,
      error: 'Schema must have a "type" property',
      suggestions: ['Add "type": "object" for object parameters', 'Add "type": "string" for string parameters']
    };
  }

  // Validate type values
  const validTypes = ['object', 'array', 'string', 'number', 'integer', 'boolean', 'null'];
  if (!validTypes.includes(schema.type)) {
    return {
      valid: false,
      error: `Invalid schema type "${schema.type}"`,
      suggestions: [`Valid types are: ${validTypes.join(', ')}`]
    };
  }

  // For object types, validate properties structure
  if (schema.type === 'object' && schema.properties) {
    if (typeof schema.properties !== 'object') {
      return { valid: false, error: 'Properties must be an object' };
    }

    // Validate each property
    for (const [propName, propSchema] of Object.entries(schema.properties)) {
      if (typeof propSchema !== 'object' || !propSchema) {
        return { valid: false, error: `Property "${propName}" must be an object with a schema` };
      }

      if (!(propSchema as any).type) {
        return { valid: false, error: `Property "${propName}" must have a type` };
      }
    }
  }

  // For array types, validate items structure
  if (schema.type === 'array' && schema.items) {
    if (typeof schema.items !== 'object' || !(schema.items as any).type) {
      return { valid: false, error: 'Array items must have a type specification' };
    }
  }

  return { valid: true };
};

// Tool examples validation
export const validateToolExamples = (examplesString: string): { valid: boolean; error?: string; suggestions?: string[] } => {
  if (!examplesString || examplesString.trim() === '') {
    return { valid: true }; // Examples are optional
  }

  const parseResult = validationUtils.parseJson(examplesString);
  if (!parseResult.success) {
    return { valid: false, error: parseResult.error };
  }

  const examples = parseResult.data;

  if (!Array.isArray(examples)) {
    return {
      valid: false,
      error: 'Examples must be an array',
      suggestions: ['Wrap examples in square brackets: [{"input": {...}, "output": {...}}]']
    };
  }

  // Validate each example
  for (let i = 0; i < examples.length; i++) {
    const example = examples[i];

    if (typeof example !== 'object' || example === null) {
      return {
        valid: false,
        error: `Example ${i + 1} must be an object`,
        suggestions: ['Each example should be an object with "input" and "output" properties']
      };
    }

    if (!example.hasOwnProperty('input')) {
      return {
        valid: false,
        error: `Example ${i + 1} must have an "input" property`,
        suggestions: ['Add "input" property with example parameters']
      };
    }

    if (!example.hasOwnProperty('output')) {
      return {
        valid: false,
        error: `Example ${i + 1} must have an "output" property`,
        suggestions: ['Add "output" property with expected result']
      };
    }
  }

  return { valid: true };
};

// Tool category suggestions based on name/description
export const suggestToolCategory = (name: string, description?: string): typeof enums.toolCategory[number] => {
  const text = `${name} ${description || ''}`.toLowerCase();

  // Category keywords
  const categoryKeywords = {
    data: ['data', 'database', 'query', 'sql', 'analytics', 'csv', 'json'],
    file: ['file', 'upload', 'download', 'storage', 'document', 'pdf', 'image'],
    web: ['http', 'api', 'request', 'web', 'url', 'scrape', 'fetch'],
    system: ['system', 'os', 'process', 'command', 'shell', 'exec'],
    ai: ['ai', 'ml', 'model', 'predict', 'nlp', 'gpt', 'openai'],
    dev: ['git', 'deploy', 'build', 'test', 'debug', 'code', 'npm'],
  };

  // Check for keyword matches
  for (const [category, keywords] of Object.entries(categoryKeywords)) {
    if (keywords.some(keyword => text.includes(keyword))) {
      return category as typeof enums.toolCategory[number];
    }
  }

  return 'general'; // Default category
};

// Implementation type suggestions
export const suggestImplementationType = (endpointUrl?: string): typeof enums.toolImplementationType[number] => {
  if (!endpointUrl) return 'internal';

  if (endpointUrl.includes('webhook') || endpointUrl.includes('callback')) {
    return 'webhook';
  }

  if (endpointUrl.startsWith('http')) {
    return 'external';
  }

  return 'internal';
};

// Generate tool schema template
export const generateToolSchemaTemplate = (category: typeof enums.toolCategory[number]): string => {
  const templates = {
    general: {
      type: 'object',
      properties: {
        input: {
          type: 'string',
          description: 'Input parameter'
        }
      },
      required: ['input']
    },
    data: {
      type: 'object',
      properties: {
        query: {
          type: 'string',
          description: 'Query string'
        },
        format: {
          type: 'string',
          enum: ['json', 'csv', 'xml'],
          description: 'Output format'
        }
      },
      required: ['query']
    },
    file: {
      type: 'object',
      properties: {
        path: {
          type: 'string',
          description: 'File path'
        },
        operation: {
          type: 'string',
          enum: ['read', 'write', 'delete'],
          description: 'File operation'
        }
      },
      required: ['path', 'operation']
    },
    web: {
      type: 'object',
      properties: {
        url: {
          type: 'string',
          format: 'uri',
          description: 'URL to request'
        },
        method: {
          type: 'string',
          enum: ['GET', 'POST', 'PUT', 'DELETE'],
          description: 'HTTP method'
        },
        headers: {
          type: 'object',
          description: 'Request headers'
        }
      },
      required: ['url']
    }
  };

  const template = templates[category as keyof typeof templates] || templates.general;
  return JSON.stringify(template, null, 2);
};

// Tool creation schema
export const createToolSchema = toolSchemaWithImplementationValidation.omit({ usage_count: true });

// Tool update schema
export const updateToolSchema = toolSchemaWithImplementationValidation.partial();

export type ToolFormData = z.infer<typeof toolSchema>;
export type CreateToolData = z.infer<typeof createToolSchema>;
export type UpdateToolData = z.infer<typeof updateToolSchema>;
