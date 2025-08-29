import * as z from 'zod';
import { commonValidation, validationUtils, enums } from '../common';

describe('Common Validation', () => {
  describe('UUID validation', () => {
    it('should validate correct UUID format', () => {
      const validUuid = '123e4567-e89b-12d3-a456-426614174000';
      expect(() => commonValidation.uuid.parse(validUuid)).not.toThrow();
      expect(() => commonValidation.requiredUuid.parse(validUuid)).not.toThrow();
    });

    it('should reject invalid UUID formats', () => {
      const invalidUuids = ['invalid-uuid', '123', '', 'not-a-uuid-at-all'];

      invalidUuids.forEach(uuid => {
        expect(() => commonValidation.uuid.parse(uuid)).toThrow('Invalid UUID format');
        expect(() => commonValidation.requiredUuid.parse(uuid)).toThrow('Invalid UUID format');
      });
    });

    it('should handle optional UUID', () => {
      expect(commonValidation.optionalUuid.parse(undefined)).toBeUndefined();
      expect(commonValidation.optionalUuid.parse('123e4567-e89b-12d3-a456-426614174000')).toBe('123e4567-e89b-12d3-a456-426614174000');
    });
  });

  describe('Name validation', () => {
    it('should accept valid names', () => {
      const validNames = ['Test Server', 'api-gateway-v2', 'MyApp_123'];

      validNames.forEach(name => {
        expect(() => commonValidation.name.parse(name)).not.toThrow();
      });
    });

    it('should trim whitespace', () => {
      expect(commonValidation.name.parse('  Test Name  ')).toBe('Test Name');
    });

    it('should reject empty or too long names', () => {
      expect(() => commonValidation.name.parse('')).toThrow();
      expect(() => commonValidation.name.parse('x'.repeat(256))).toThrow();
    });
  });

  describe('Description validation', () => {
    it('should accept valid descriptions', () => {
      const validDescriptions = ['A simple description', '', undefined];

      validDescriptions.forEach(desc => {
        expect(() => commonValidation.description.parse(desc)).not.toThrow();
      });
    });

    it('should trim whitespace', () => {
      expect(commonValidation.description.parse('  Test Description  ')).toBe('Test Description');
    });

    it('should reject too long descriptions', () => {
      expect(() => commonValidation.description.parse('x'.repeat(1001))).toThrow('Description must be less than 1000 characters');
    });

    it('should convert empty string to literal empty', () => {
      expect(commonValidation.description.parse('')).toBe('');
    });
  });

  describe('Email validation', () => {
    it('should validate correct email formats', () => {
      const validEmails = [
        'test@example.com',
        'user.name@domain.co.uk',
        'first+last@test-domain.org'
      ];

      validEmails.forEach(email => {
        expect(() => commonValidation.email.parse(email)).not.toThrow();
      });
    });

    it('should convert to lowercase', () => {
      expect(commonValidation.email.parse('TEST@EXAMPLE.COM')).toBe('test@example.com');
    });

    it('should reject invalid email formats', () => {
      const invalidEmails = ['invalid', 'test@', '@example.com', 'test..test@example.com'];

      invalidEmails.forEach(email => {
        expect(() => commonValidation.email.parse(email)).toThrow('Invalid email address');
      });
    });

    it('should reject too long emails', () => {
      const longEmail = 'x'.repeat(250) + '@example.com';
      expect(() => commonValidation.email.parse(longEmail)).toThrow('Email must be less than 255 characters');
    });
  });

  describe('URL validation', () => {
    it('should validate correct URL formats', () => {
      const validUrls = [
        'https://example.com',
        'http://localhost:3000',
        'https://api.example.com/v1/endpoint'
      ];

      validUrls.forEach(url => {
        expect(() => commonValidation.url.parse(url)).not.toThrow();
        expect(() => commonValidation.optionalUrl.parse(url)).not.toThrow();
      });
    });

    it('should handle optional URLs', () => {
      expect(commonValidation.optionalUrl.parse(undefined)).toBeUndefined();
      expect(commonValidation.optionalUrl.parse('')).toBe('');
    });

    it('should reject invalid URL formats', () => {
      const invalidUrls = ['not-a-url', 'javascript:alert(1)'];

      invalidUrls.forEach(url => {
        expect(() => commonValidation.url.parse(url)).toThrow();
      });
    });
  });

  describe('URI validation', () => {
    it('should validate various URI schemes', () => {
      const validUris = [
        'http://example.com',
        'https://example.com',
        'file:///path/to/file',
        'ftp://ftp.example.com',
        '/relative/path',
        './relative/path'
      ];

      validUris.forEach(uri => {
        expect(() => commonValidation.uri.parse(uri)).not.toThrow();
      });
    });

    it('should reject invalid URI formats', () => {
      const invalidUris = ['', 'not-a-uri'];

      invalidUris.forEach(uri => {
        expect(() => commonValidation.uri.parse(uri)).toThrow();
      });
    });
  });

  describe('Tags validation', () => {
    it('should validate tag arrays', () => {
      const validTags = [['api', 'v1'], ['production', 'stable'], []];

      validTags.forEach(tags => {
        expect(() => commonValidation.tags.parse(tags)).not.toThrow();
      });
    });

    it('should default to empty array', () => {
      expect(commonValidation.tags.parse(undefined)).toEqual([]);
    });

    it('should reject too many tags', () => {
      const tooManyTags = Array.from({length: 21}, (_, i) => `tag${i}`);
      expect(() => commonValidation.tags.parse(tooManyTags)).toThrow('Maximum 20 tags allowed');
    });

    it('should reject invalid tag formats', () => {
      expect(() => commonValidation.tags.parse([''])).toThrow();
      expect(() => commonValidation.tags.parse(['x'.repeat(51)])).toThrow();
    });
  });

  describe('JSON validation', () => {
    it('should validate JSON strings', () => {
      const validJson = ['{"key": "value"}', '[]', 'null', '"string"'];

      validJson.forEach(json => {
        expect(() => commonValidation.jsonString.parse(json)).not.toThrow();
        expect(() => commonValidation.requiredJsonString.parse(json)).not.toThrow();
      });
    });

    it('should accept empty optional JSON', () => {
      expect(commonValidation.jsonString.parse('')).toBe('');
      expect(commonValidation.jsonString.parse(undefined)).toBeUndefined();
    });

    it('should reject invalid JSON', () => {
      const invalidJson = ['{invalid}', '{"unclosed": }', 'undefined'];

      invalidJson.forEach(json => {
        expect(() => commonValidation.jsonString.parse(json)).toThrow('Invalid JSON format');
        expect(() => commonValidation.requiredJsonString.parse(json)).toThrow('Invalid JSON format');
      });
    });

    it('should require non-empty JSON for requiredJsonString', () => {
      expect(() => commonValidation.requiredJsonString.parse('')).toThrow('JSON is required');
    });
  });

  describe('Slug validation', () => {
    it('should validate correct slug formats', () => {
      const validSlugs = ['my-app', 'api_v2', 'test-server-1'];

      validSlugs.forEach(slug => {
        expect(() => commonValidation.slug.parse(slug)).not.toThrow();
      });
    });

    it('should reject invalid slug formats', () => {
      const invalidSlugs = ['My App', 'api-v2!', '-starts-with-dash', 'ends-with-dash-', ''];

      invalidSlugs.forEach(slug => {
        expect(() => commonValidation.slug.parse(slug)).toThrow();
      });
    });
  });

  describe('Function name validation', () => {
    it('should validate correct function names', () => {
      const validNames = ['functionName', 'my_function', '_privateFunction', 'func123'];

      validNames.forEach(name => {
        expect(() => commonValidation.functionName.parse(name)).not.toThrow();
      });
    });

    it('should reject invalid function names', () => {
      const invalidNames = ['123function', 'func-name', 'function name', ''];

      invalidNames.forEach(name => {
        expect(() => commonValidation.functionName.parse(name)).toThrow();
      });
    });
  });

  describe('MIME type validation', () => {
    it('should validate correct MIME types', () => {
      const validTypes = ['application/json', 'text/plain', 'image/jpeg'];

      validTypes.forEach(type => {
        expect(() => commonValidation.mimeType.parse(type)).not.toThrow();
      });
    });

    it('should handle optional MIME types', () => {
      expect(commonValidation.mimeType.parse(undefined)).toBeUndefined();
    });

    it('should reject invalid MIME types', () => {
      const invalidTypes = ['invalid', 'application/', '/json'];

      invalidTypes.forEach(type => {
        expect(() => commonValidation.mimeType.parse(type)).toThrow('Invalid MIME type format');
      });
    });
  });

  describe('Numeric validations', () => {
    it('should validate file sizes', () => {
      expect(() => commonValidation.fileSize.parse(1024)).not.toThrow();
      expect(commonValidation.fileSize.parse(undefined)).toBeUndefined();
    });

    it('should reject invalid file sizes', () => {
      expect(() => commonValidation.fileSize.parse(-1)).toThrow('File size cannot be negative');
      expect(() => commonValidation.fileSize.parse(1024 * 1024 * 1024 + 1)).toThrow('File size cannot exceed 1GB');
    });

    it('should validate timeout values', () => {
      expect(() => commonValidation.timeout.parse(30)).not.toThrow();
      expect(() => commonValidation.timeout.parse(0)).toThrow('Timeout must be at least 1 second');
      expect(() => commonValidation.timeout.parse(301)).toThrow('Timeout cannot exceed 300 seconds');
    });

    it('should validate retry counts', () => {
      expect(() => commonValidation.retries.parse(3)).not.toThrow();
      expect(() => commonValidation.retries.parse(-1)).toThrow('Retry count cannot be negative');
      expect(() => commonValidation.retries.parse(11)).toThrow('Retry count cannot exceed 10');
    });

    it('should validate rate limits', () => {
      expect(() => commonValidation.rateLimit.parse(100)).not.toThrow();
      expect(() => commonValidation.rateLimit.parse(0)).toThrow('Rate limit must be at least 1 request');
    });
  });

  describe('Boolean defaults', () => {
    it('should handle boolean defaults', () => {
      expect(commonValidation.booleanDefault(true).parse(undefined)).toBe(true);
      expect(commonValidation.booleanDefault(false).parse(undefined)).toBe(false);
      expect(commonValidation.isActive.parse(undefined)).toBe(true);
      expect(commonValidation.isPublic.parse(undefined)).toBe(false);
    });
  });
});

describe('Validation Utils', () => {
  describe('slugify function', () => {
    it('should convert strings to valid slugs', () => {
      expect(validationUtils.slugify('My App Name')).toBe('my-app-name');
      expect(validationUtils.slugify('  API v2.0!  ')).toBe('api-v20');
      expect(validationUtils.slugify('Test---Server')).toBe('test-server');
      expect(validationUtils.slugify('-start-end-')).toBe('start-end');
    });
  });

  describe('parseJson function', () => {
    it('should parse valid JSON', () => {
      const result = validationUtils.parseJson('{"key": "value"}');
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data).toEqual({key: 'value'});
      }
    });

    it('should handle invalid JSON', () => {
      const result = validationUtils.parseJson('{invalid}');
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error).toContain('Invalid JSON');
      }
    });
  });

  describe('validateToolSchema function', () => {
    it('should validate correct tool schemas', () => {
      const validSchema = '{"type": "object", "properties": {"name": {"type": "string"}}}';
      const result = validationUtils.validateToolSchema(validSchema);
      expect(result.valid).toBe(true);
    });

    it('should reject invalid tool schemas', () => {
      const invalidSchema = '{"invalid": "schema"}';
      const result = validationUtils.validateToolSchema(invalidSchema);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Schema must have a "type" property');
    });

    it('should handle invalid JSON', () => {
      const result = validationUtils.validateToolSchema('{invalid}');
      expect(result.valid).toBe(false);
    });
  });

  describe('validatePromptParameters function', () => {
    it('should validate correct prompt parameters', () => {
      const validParams = '[{"name": "input", "type": "string"}]';
      const result = validationUtils.validatePromptParameters(validParams);
      expect(result.valid).toBe(true);
    });

    it('should reject non-array parameters', () => {
      const invalidParams = '{"name": "input"}';
      const result = validationUtils.validatePromptParameters(invalidParams);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Parameters must be an array');
    });

    it('should validate parameter structure', () => {
      const invalidParams = '[{"type": "string"}]';
      const result = validationUtils.validatePromptParameters(invalidParams);
      expect(result.valid).toBe(false);
      expect(result.error).toContain('must be an object with a "name" property');
    });
  });

  describe('getValidationError function', () => {
    it('should return first error message', () => {
      const errors = [{message: 'First error'}, {message: 'Second error'}];
      expect(validationUtils.getValidationError(errors)).toBe('First error');
    });

    it('should return empty string for no errors', () => {
      expect(validationUtils.getValidationError([])).toBe('');
    });

    it('should handle errors without message', () => {
      const errors = [{}];
      expect(validationUtils.getValidationError(errors)).toBe('Validation error');
    });
  });
});

describe('Enum definitions', () => {
  describe('planType enum', () => {
    it('should contain expected plan types', () => {
      expect(enums.planType).toEqual(['free', 'pro', 'enterprise']);
    });
  });

  describe('protocol enum', () => {
    it('should contain expected protocols', () => {
      expect(enums.protocol).toEqual(['http', 'https', 'stdio', 'websocket', 'sse']);
    });
  });

  describe('serverStatus enum', () => {
    it('should contain expected statuses', () => {
      expect(enums.serverStatus).toEqual(['active', 'inactive', 'error', 'pending']);
    });
  });

  describe('userRole enum', () => {
    it('should contain expected roles', () => {
      expect(enums.userRole).toEqual(['admin', 'user', 'viewer', 'api_user']);
    });
  });

  describe('resourceType enum', () => {
    it('should contain expected resource types', () => {
      expect(enums.resourceType).toEqual(['file', 'url', 'database', 'api', 'memory', 'custom']);
    });
  });

  describe('toolCategory enum', () => {
    it('should contain expected tool categories', () => {
      expect(enums.toolCategory).toEqual(['general', 'data', 'file', 'web', 'system', 'ai', 'dev', 'custom']);
    });
  });

  describe('adapterType enum', () => {
    it('should contain expected adapter types', () => {
      expect(enums.adapterType).toEqual(['REST', 'GraphQL', 'gRPC', 'SOAP']);
    });
  });

  describe('all other enums', () => {
    it('should have proper enum structures', () => {
      expect(Array.isArray(enums.sessionStatus)).toBe(true);
      expect(Array.isArray(enums.processStatus)).toBe(true);
      expect(Array.isArray(enums.toolImplementationType)).toBe(true);
      expect(Array.isArray(enums.promptCategory)).toBe(true);
      expect(Array.isArray(enums.logLevel)).toBe(true);
      expect(Array.isArray(enums.windowType)).toBe(true);
      expect(Array.isArray(enums.rateLimitScope)).toBe(true);
      expect(Array.isArray(enums.auditAction)).toBe(true);
      expect(Array.isArray(enums.importStatus)).toBe(true);
      expect(Array.isArray(enums.conflictStrategy)).toBe(true);
    });
  });
});
