import * as z from 'zod';
import {
  toolSchema,
  toolSchemaWithImplementationValidation,
  createToolSchema,
  updateToolSchema,
  validateToolSchema,
  validateToolExamples,
  suggestToolCategory,
  suggestImplementationType,
  generateToolSchemaTemplate
} from '../tool';

describe('Tool Validation Schemas', () => {
  describe('toolSchema', () => {
    it('should validate basic tool data', () => {
      const validTool = {
        name: 'File Reader',
        description: 'Reads files from the filesystem',
        function_name: 'read_file',
        category: 'file' as const,
        implementation_type: 'internal' as const,
        endpoint_url: 'https://api.example.com/tools/file-reader',
        timeout_seconds: 30,
        max_retries: 3,
        usage_count: 0,
        is_active: true,
        is_public: false,
        schema: '{"type": "object", "properties": {"path": {"type": "string"}}}',
        examples: '[{"input": {"path": "/test.txt"}, "output": "File contents"}]',
        documentation: 'This tool reads files from the local filesystem.',
        tags: ['file', 'io', 'utility'],
        access_permissions: '["read"]',
        metadata: '{"version": "1.0"}'
      };

      expect(() => toolSchema.parse(validTool)).not.toThrow();
    });

    it('should apply defaults for optional fields', () => {
      const minimalTool = {
        name: 'Simple Tool',
        function_name: 'simple_tool',
        category: 'general' as const
      };

      const parsed = toolSchema.parse(minimalTool);
      expect(parsed.implementation_type).toBe('internal');
      expect(parsed.timeout_seconds).toBe(30);
      expect(parsed.max_retries).toBe(3);
      expect(parsed.usage_count).toBe(0);
      expect(parsed.is_active).toBe(true);
      expect(parsed.is_public).toBe(false);
      expect(parsed.tags).toEqual([]);
    });

    it('should validate tool categories', () => {
      const validCategories = ['general', 'data', 'file', 'web', 'system', 'ai', 'dev', 'custom'];

      validCategories.forEach(category => {
        const tool = {
          name: 'Test Tool',
          function_name: 'test_tool',
          category: category as any
        };

        expect(() => toolSchema.parse(tool)).not.toThrow();
      });
    });

    it('should reject invalid tool categories', () => {
      const tool = {
        name: 'Test Tool',
        function_name: 'test_tool',
        category: 'invalid-category'
      };

      expect(() => toolSchema.parse(tool)).toThrow();
    });

    it('should validate implementation types', () => {
      const validTypes = ['internal', 'external', 'webhook', 'script'];

      validTypes.forEach(type => {
        const tool = {
          name: 'Test Tool',
          function_name: 'test_tool',
          category: 'general' as const,
          implementation_type: type as any
        };

        expect(() => toolSchema.parse(tool)).not.toThrow();
      });
    });

    it('should reject invalid implementation types', () => {
      const tool = {
        name: 'Test Tool',
        function_name: 'test_tool',
        category: 'general' as const,
        implementation_type: 'invalid-type'
      };

      expect(() => toolSchema.parse(tool)).toThrow();
    });

    it('should validate function names', () => {
      const validFunctionNames = ['myFunction', 'my_function', '_privateFunction', 'func123'];

      validFunctionNames.forEach(functionName => {
        const tool = {
          name: 'Test Tool',
          function_name: functionName,
          category: 'general' as const
        };

        expect(() => toolSchema.parse(tool)).not.toThrow();
      });
    });

    it('should reject invalid function names', () => {
      const invalidFunctionNames = ['123function', 'func-name', 'function name', ''];

      invalidFunctionNames.forEach(functionName => {
        const tool = {
          name: 'Test Tool',
          function_name: functionName,
          category: 'general' as const
        };

        expect(() => toolSchema.parse(tool)).toThrow();
      });
    });

    it('should validate documentation length', () => {
      const tool = {
        name: 'Test Tool',
        function_name: 'test_tool',
        category: 'general' as const,
        documentation: 'x'.repeat(5001)
      };

      expect(() => toolSchema.parse(tool)).toThrow('Documentation must be less than 5000 characters');
    });

    it('should validate tags array', () => {
      const tooManyTags = Array.from({length: 21}, (_, i) => `tag${i}`);
      const tool = {
        name: 'Test Tool',
        function_name: 'test_tool',
        category: 'general' as const,
        tags: tooManyTags
      };

      expect(() => toolSchema.parse(tool)).toThrow('Maximum 20 tags allowed');
    });
  });

  describe('toolSchemaWithImplementationValidation', () => {
    describe('External implementation validation', () => {
      it('should require endpoint URL for external implementation', () => {
        const externalTool = {
          name: 'External API Tool',
          function_name: 'external_api',
          category: 'web' as const,
          implementation_type: 'external' as const
        };

        expect(() => toolSchemaWithImplementationValidation.parse(externalTool))
          .toThrow();
      });

      it('should accept valid external tool with endpoint', () => {
        const externalTool = {
          name: 'External API Tool',
          function_name: 'external_api',
          category: 'web' as const,
          implementation_type: 'external' as const,
          endpoint_url: 'https://api.example.com/tool'
        };

        expect(() => toolSchemaWithImplementationValidation.parse(externalTool)).not.toThrow();
      });
    });

    describe('Webhook implementation validation', () => {
      it('should require endpoint URL for webhook implementation', () => {
        const webhookTool = {
          name: 'Webhook Tool',
          function_name: 'webhook_tool',
          category: 'web' as const,
          implementation_type: 'webhook' as const
        };

        expect(() => toolSchemaWithImplementationValidation.parse(webhookTool))
          .toThrow();
      });

      it('should prefer HTTPS for webhook URLs', () => {
        const httpWebhookTool = {
          name: 'HTTP Webhook Tool',
          function_name: 'http_webhook',
          category: 'web' as const,
          implementation_type: 'webhook' as const,
          endpoint_url: 'http://api.example.com/webhook'
        };

        expect(() => toolSchemaWithImplementationValidation.parse(httpWebhookTool))
          .toThrow('Webhook URLs should use HTTPS for security');
      });

      it('should allow HTTP for localhost webhooks', () => {
        const localhostWebhooks = [
          {
            name: 'Localhost Webhook',
            function_name: 'localhost_webhook',
            category: 'dev' as const,
            implementation_type: 'webhook' as const,
            endpoint_url: 'http://localhost:3000/webhook'
          },
          {
            name: '127.0.0.1 Webhook',
            function_name: 'local_ip_webhook',
            category: 'dev' as const,
            implementation_type: 'webhook' as const,
            endpoint_url: 'http://127.0.0.1:8080/webhook'
          }
        ];

        localhostWebhooks.forEach(tool => {
          expect(() => toolSchemaWithImplementationValidation.parse(tool)).not.toThrow();
        });
      });

      it('should accept HTTPS webhook URLs', () => {
        const httpsWebhookTool = {
          name: 'HTTPS Webhook Tool',
          function_name: 'https_webhook',
          category: 'web' as const,
          implementation_type: 'webhook' as const,
          endpoint_url: 'https://api.example.com/webhook'
        };

        expect(() => toolSchemaWithImplementationValidation.parse(httpsWebhookTool)).not.toThrow();
      });
    });

    describe('Internal/Script implementation validation', () => {
      it('should not require endpoint URL for internal implementation', () => {
        const internalTool = {
          name: 'Internal Tool',
          function_name: 'internal_tool',
          category: 'system' as const,
          implementation_type: 'internal' as const
        };

        expect(() => toolSchemaWithImplementationValidation.parse(internalTool)).not.toThrow();
      });

      it('should not require endpoint URL for script implementation', () => {
        const scriptTool = {
          name: 'Script Tool',
          function_name: 'script_tool',
          category: 'system' as const,
          implementation_type: 'script' as const
        };

        expect(() => toolSchemaWithImplementationValidation.parse(scriptTool)).not.toThrow();
      });
    });

    describe('URL format validation', () => {
      it('should reject invalid URL formats', () => {
        const invalidUrlTool = {
          name: 'Invalid URL Tool',
          function_name: 'invalid_url',
          category: 'web' as const,
          implementation_type: 'external' as const,
          endpoint_url: 'not-a-valid-url'
        };

        expect(() => toolSchemaWithImplementationValidation.parse(invalidUrlTool))
          .toThrow('Endpoint URL must be a valid HTTP/HTTPS URL');
      });

      it('should accept valid HTTP/HTTPS URLs', () => {
        const validUrls = [
          'https://api.example.com/tool',
          'http://localhost:3000/api',
          'https://subdomain.example.com:8080/path'
        ];

        validUrls.forEach(url => {
          const tool = {
            name: 'Test Tool',
            function_name: 'test_tool',
            category: 'web' as const,
            implementation_type: 'external' as const,
            endpoint_url: url
          };

          expect(() => toolSchemaWithImplementationValidation.parse(tool)).not.toThrow();
        });
      });
    });
  });

  describe('Schema compositions', () => {
    it('should create tool with createToolSchema', () => {
      const createTool = {
        name: 'New Tool',
        function_name: 'new_tool',
        category: 'general' as const,
        implementation_type: 'internal' as const
      };

      expect(() => createToolSchema.parse(createTool)).not.toThrow();
    });

    it('should exclude usage_count from createToolSchema', () => {
      const createTool = {
        name: 'New Tool',
        function_name: 'new_tool',
        category: 'general' as const
      };

      const parsed = createToolSchema.parse(createTool);
      expect(parsed.usage_count).toBeUndefined(); // Field should be excluded
    });

    it('should allow partial updates with updateToolSchema', () => {
      const partialUpdate = {
        name: 'Updated Tool Name',
        timeout_seconds: 60
      };

      expect(() => updateToolSchema.parse(partialUpdate)).not.toThrow();
    });
  });
});

describe('Tool Schema Validation', () => {
  describe('validateToolSchema', () => {
    it('should allow empty schema', () => {
      const result = validateToolSchema('');
      expect(result.valid).toBe(true);
    });

    it('should validate basic object schema', () => {
      const objectSchema = JSON.stringify({
        type: 'object',
        properties: {
          name: { type: 'string' },
          age: { type: 'number' }
        },
        required: ['name']
      });

      const result = validateToolSchema(objectSchema);
      expect(result.valid).toBe(true);
    });

    it('should validate array schema', () => {
      const arraySchema = JSON.stringify({
        type: 'array',
        items: { type: 'string' }
      });

      const result = validateToolSchema(arraySchema);
      expect(result.valid).toBe(true);
    });

    it('should validate primitive type schemas', () => {
      const primitiveSchemas = ['string', 'number', 'integer', 'boolean', 'null'];

      primitiveSchemas.forEach(type => {
        const schema = JSON.stringify({ type });
        const result = validateToolSchema(schema);
        expect(result.valid).toBe(true);
      });
    });

    it('should reject invalid JSON', () => {
      const result = validateToolSchema('{invalid json}');
      expect(result.valid).toBe(false);
      expect(result.error).toContain('Invalid JSON');
    });

    it('should reject non-object schemas', () => {
      const result = validateToolSchema('"string"');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Schema must be an object');
    });

    it('should require type property', () => {
      const result = validateToolSchema('{"properties": {}}');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Schema must have a "type" property');
      expect(result.suggestions).toContain('Add "type": "object" for object parameters');
    });

    it('should validate type values', () => {
      const result = validateToolSchema('{"type": "invalid-type"}');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Invalid schema type "invalid-type"');
      expect(result.suggestions).toBeDefined();
    });

    it('should validate object properties structure', () => {
      const invalidObjectSchema = JSON.stringify({
        type: 'object',
        properties: {
          invalidProp: 'not-an-object'
        }
      });

      const result = validateToolSchema(invalidObjectSchema);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Property "invalidProp" must be an object with a schema');
    });

    it('should validate object property types', () => {
      const invalidObjectSchema = JSON.stringify({
        type: 'object',
        properties: {
          propWithoutType: { description: 'Missing type' }
        }
      });

      const result = validateToolSchema(invalidObjectSchema);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Property "propWithoutType" must have a type');
    });

    it('should validate array items structure', () => {
      const invalidArraySchema = JSON.stringify({
        type: 'array',
        items: { description: 'Missing type' }
      });

      const result = validateToolSchema(invalidArraySchema);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Array items must have a type specification');
    });
  });
});

describe('Tool Examples Validation', () => {
  describe('validateToolExamples', () => {
    it('should allow empty examples', () => {
      const result = validateToolExamples('');
      expect(result.valid).toBe(true);
    });

    it('should validate correct examples format', () => {
      const examples = JSON.stringify([
        {
          input: { path: '/test.txt' },
          output: 'File contents here'
        },
        {
          input: { path: '/data.json' },
          output: { data: 'parsed json' }
        }
      ]);

      const result = validateToolExamples(examples);
      expect(result.valid).toBe(true);
    });

    it('should reject invalid JSON', () => {
      const result = validateToolExamples('{invalid json}');
      expect(result.valid).toBe(false);
      expect(result.error).toContain('Invalid JSON');
    });

    it('should reject non-array examples', () => {
      const result = validateToolExamples('{"input": {}, "output": {}}');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Examples must be an array');
      expect(result.suggestions).toContain('Wrap examples in square brackets: [{"input": {...}, "output": {...}}]');
    });

    it('should validate example structure', () => {
      const invalidExamples = JSON.stringify([
        'not an object'
      ]);

      const result = validateToolExamples(invalidExamples);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Example 1 must be an object');
      expect(result.suggestions).toContain('Each example should be an object with "input" and "output" properties');
    });

    it('should require input property', () => {
      const invalidExamples = JSON.stringify([
        {
          output: 'result'
        }
      ]);

      const result = validateToolExamples(invalidExamples);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Example 1 must have an "input" property');
      expect(result.suggestions).toContain('Add "input" property with example parameters');
    });

    it('should require output property', () => {
      const invalidExamples = JSON.stringify([
        {
          input: { param: 'value' }
        }
      ]);

      const result = validateToolExamples(invalidExamples);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Example 1 must have an "output" property');
      expect(result.suggestions).toContain('Add "output" property with expected result');
    });

    it('should validate multiple examples', () => {
      const invalidExamples = JSON.stringify([
        {
          input: { param: 'value1' },
          output: 'result1'
        },
        {
          input: { param: 'value2' }
          // Missing output
        }
      ]);

      const result = validateToolExamples(invalidExamples);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Example 2 must have an "output" property');
    });
  });
});

describe('Tool Suggestion Functions', () => {
  describe('suggestToolCategory', () => {
    it('should suggest data category for data-related tools', () => {
      const dataTools = [
        { name: 'Database Query', description: 'Execute SQL queries' },
        { name: 'CSV Parser', description: 'Parse CSV files' },
        { name: 'Analytics Tool', description: 'Analyze data patterns' }
      ];

      dataTools.forEach(tool => {
        const category = suggestToolCategory(tool.name, tool.description);
        expect(category).toBe('data');
      });
    });

    it('should suggest file category for file-related tools', () => {
      const fileTools = [
        { name: 'File Reader', description: 'Read files from disk' },
        { name: 'Document Processor', description: 'Process PDF documents' },
        { name: 'Image Uploader', description: 'Upload images to storage' }
      ];

      fileTools.forEach(tool => {
        const category = suggestToolCategory(tool.name, tool.description);
        expect(category).toBe('file');
      });
    });

    it('should suggest web category for web-related tools', () => {
      const webTools = [
        { name: 'HTTP Client', description: 'Make HTTP requests' },
        { name: 'API Caller', description: 'Call external APIs' },
        { name: 'Web Scraper', description: 'Scrape web pages' }
      ];

      webTools.forEach(tool => {
        const category = suggestToolCategory(tool.name, tool.description);
        expect(category).toBe('web');
      });
    });

    it('should suggest system category for system-related tools', () => {
      const systemTools = [
        { name: 'Process Manager', description: 'Manage system processes' },
        { name: 'Shell Executor', description: 'Execute shell commands' },
        { name: 'OS Info', description: 'Get operating system information' }
      ];

      systemTools.forEach(tool => {
        const category = suggestToolCategory(tool.name, tool.description);
        expect(category).toBe('system');
      });
    });

    it('should suggest ai category for AI-related tools', () => {
      const aiTools = [
        { name: 'GPT Client', description: 'Connect to OpenAI GPT' },
        { name: 'ML Model', description: 'Machine learning predictions' },
        { name: 'AI NLP Processor', description: 'Natural language processing' }
      ];

      aiTools.forEach(tool => {
        const category = suggestToolCategory(tool.name, tool.description);
        expect(category).toBe('ai');
      });
    });

    it('should suggest dev category for development tools', () => {
      const devTools = [
        { name: 'Git Manager', description: 'Manage git repositories' },
        { name: 'NPM Build Tool', description: 'Build and deploy applications' },
        { name: 'Code Test Runner', description: 'Run unit tests' }
      ];

      devTools.forEach(tool => {
        const category = suggestToolCategory(tool.name, tool.description);
        expect(category).toBe('dev');
      });
    });

    it('should default to general category', () => {
      const generalTools = [
        { name: 'Unknown Tool', description: 'Does something' },
        { name: 'Generic Helper', description: 'Helps with stuff' }
      ];

      generalTools.forEach(tool => {
        const category = suggestToolCategory(tool.name, tool.description);
        expect(category).toBe('general');
      });
    });
  });

  describe('suggestImplementationType', () => {
    it('should suggest internal for no endpoint URL', () => {
      expect(suggestImplementationType()).toBe('internal');
      expect(suggestImplementationType('')).toBe('internal');
    });

    it('should suggest webhook for webhook URLs', () => {
      const webhookUrls = [
        'https://api.example.com/webhook',
        'http://localhost:3000/callback'
      ];

      webhookUrls.forEach(url => {
        const type = suggestImplementationType(url);
        expect(type).toBe('webhook');
      });
    });

    it('should suggest external for HTTP URLs', () => {
      const httpUrls = [
        'https://api.example.com/tool',
        'http://service.example.com/api'
      ];

      httpUrls.forEach(url => {
        const type = suggestImplementationType(url);
        expect(type).toBe('external');
      });
    });

    it('should default to internal for non-HTTP URLs', () => {
      const nonHttpUrls = [
        'ftp://files.example.com',
        'file:///local/path'
      ];

      nonHttpUrls.forEach(url => {
        const type = suggestImplementationType(url);
        expect(type).toBe('internal');
      });
    });
  });

  describe('generateToolSchemaTemplate', () => {
    it('should generate general template', () => {
      const template = generateToolSchemaTemplate('general');
      const parsed = JSON.parse(template);

      expect(parsed.type).toBe('object');
      expect(parsed.properties).toBeDefined();
      expect(parsed.properties.input).toBeDefined();
      expect(parsed.required).toContain('input');
    });

    it('should generate data template', () => {
      const template = generateToolSchemaTemplate('data');
      const parsed = JSON.parse(template);

      expect(parsed.type).toBe('object');
      expect(parsed.properties.query).toBeDefined();
      expect(parsed.properties.format).toBeDefined();
      expect(parsed.properties.format.enum).toEqual(['json', 'csv', 'xml']);
    });

    it('should generate file template', () => {
      const template = generateToolSchemaTemplate('file');
      const parsed = JSON.parse(template);

      expect(parsed.properties.path).toBeDefined();
      expect(parsed.properties.operation).toBeDefined();
      expect(parsed.properties.operation.enum).toEqual(['read', 'write', 'delete']);
    });

    it('should generate web template', () => {
      const template = generateToolSchemaTemplate('web');
      const parsed = JSON.parse(template);

      expect(parsed.properties.url).toBeDefined();
      expect(parsed.properties.url.format).toBe('uri');
      expect(parsed.properties.method).toBeDefined();
      expect(parsed.properties.method.enum).toEqual(['GET', 'POST', 'PUT', 'DELETE']);
    });

    it('should default to general template for unknown categories', () => {
      const template = generateToolSchemaTemplate('unknown' as any);
      const parsed = JSON.parse(template);

      expect(parsed.type).toBe('object');
      expect(parsed.properties.input).toBeDefined();
    });

    it('should generate valid JSON', () => {
      const categories = ['general', 'data', 'file', 'web'] as const;

      categories.forEach(category => {
        const template = generateToolSchemaTemplate(category);
        expect(() => JSON.parse(template)).not.toThrow();
      });
    });
  });
});
