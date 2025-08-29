import * as z from 'zod';
import {
  serverSchema,
  serverSchemaWithProtocolValidation,
  createServerSchema,
  updateServerSchema,
  validateEnvironmentVariable,
  validateCommandArgument
} from '../server';

describe('Server Validation Schemas', () => {
  describe('serverSchema', () => {
    it('should validate basic server data', () => {
      const validServer = {
        name: 'Test Server',
        description: 'A test MCP server',
        protocol: 'http' as const,
        url: 'https://api.example.com',
        command: 'node server.js',
        args: ['--port', '3000'],
        environment: ['NODE_ENV=production', 'PORT=3000'],
        working_dir: '/app',
        version: '1.0.0',
        timeout_seconds: 300,
        max_retries: 3,
        status: 'active' as const,
        health_check_url: 'https://api.example.com/health',
        is_active: true,
        metadata: '{"region": "us-west-2"}',
        tags: ['production', 'api']
      };

      expect(() => serverSchema.parse(validServer)).not.toThrow();
    });

    it('should apply defaults for optional fields', () => {
      const minimalServer = {
        name: 'Test Server',
        protocol: 'http' as const
      };

      const parsed = serverSchema.parse(minimalServer);
      expect(parsed.timeout_seconds).toBe(300);
      expect(parsed.max_retries).toBe(3);
      expect(parsed.status).toBe('active');
      expect(parsed.is_active).toBe(true);
      expect(parsed.args).toEqual([]);
      expect(parsed.environment).toEqual([]);
      expect(parsed.tags).toEqual([]);
    });

    it('should validate protocol enum', () => {
      const validProtocols = ['http', 'https', 'stdio', 'websocket', 'sse'];

      validProtocols.forEach(protocol => {
        const server = {
          name: 'Test Server',
          protocol: protocol as any
        };

        expect(() => serverSchema.parse(server)).not.toThrow();
      });
    });

    it('should reject invalid protocols', () => {
      const server = {
        name: 'Test Server',
        protocol: 'invalid-protocol'
      };

      expect(() => serverSchema.parse(server)).toThrow();
    });

    it('should validate server status enum', () => {
      const validStatuses = ['active', 'inactive', 'error', 'pending'];

      validStatuses.forEach(status => {
        const server = {
          name: 'Test Server',
          protocol: 'http' as const,
          status: status as any
        };

        expect(() => serverSchema.parse(server)).not.toThrow();
      });
    });

    it('should reject invalid server status', () => {
      const server = {
        name: 'Test Server',
        protocol: 'http' as const,
        status: 'invalid-status'
      };

      expect(() => serverSchema.parse(server)).toThrow();
    });

    it('should validate args array limits', () => {
      const tooManyArgs = Array.from({length: 21}, (_, i) => `arg${i}`);
      const server = {
        name: 'Test Server',
        protocol: 'stdio' as const,
        args: tooManyArgs
      };

      expect(() => serverSchema.parse(server)).toThrow('Maximum 20 arguments allowed');
    });

    it('should validate arg length limits', () => {
      const server = {
        name: 'Test Server',
        protocol: 'stdio' as const,
        args: ['x'.repeat(101)]
      };

      expect(() => serverSchema.parse(server)).toThrow('Each argument must be less than 100 characters');
    });

    it('should validate environment array limits', () => {
      const tooManyEnvVars = Array.from({length: 51}, (_, i) => `VAR${i}=value`);
      const server = {
        name: 'Test Server',
        protocol: 'stdio' as const,
        environment: tooManyEnvVars
      };

      expect(() => serverSchema.parse(server)).toThrow('Maximum 50 environment variables allowed');
    });

    it('should validate environment variable length', () => {
      const server = {
        name: 'Test Server',
        protocol: 'stdio' as const,
        environment: ['VAR=' + 'x'.repeat(197)] // 200 char limit, so 197 + 'VAR=' = 201
      };

      expect(() => serverSchema.parse(server)).toThrow();
    });

    it('should validate timeout and retry limits', () => {
      const server = {
        name: 'Test Server',
        protocol: 'http' as const,
        timeout_seconds: 0,
        max_retries: -1
      };

      expect(() => serverSchema.parse(server)).toThrow();
    });
  });

  describe('serverSchemaWithProtocolValidation', () => {
    describe('HTTP/HTTPS protocol validation', () => {
      it('should require URL for HTTP protocols', () => {
        const httpServer = {
          name: 'HTTP Server',
          protocol: 'http' as const
        };

        expect(() => serverSchemaWithProtocolValidation.parse(httpServer)).toThrow();

        const httpsServer = {
          name: 'HTTPS Server',
          protocol: 'https' as const
        };

        expect(() => serverSchemaWithProtocolValidation.parse(httpsServer)).toThrow();
      });

      it('should accept valid HTTP/HTTPS URLs', () => {
        const httpServer = {
          name: 'HTTP Server',
          protocol: 'http' as const,
          url: 'http://localhost:3000'
        };

        const httpsServer = {
          name: 'HTTPS Server',
          protocol: 'https' as const,
          url: 'https://api.example.com'
        };

        expect(() => serverSchemaWithProtocolValidation.parse(httpServer)).not.toThrow();
        expect(() => serverSchemaWithProtocolValidation.parse(httpsServer)).not.toThrow();
      });
    });

    describe('SSE protocol validation', () => {
      it('should require URL for SSE protocol', () => {
        const sseServer = {
          name: 'SSE Server',
          protocol: 'sse' as const
        };

        expect(() => serverSchemaWithProtocolValidation.parse(sseServer)).toThrow();
      });

      it('should accept valid SSE URL', () => {
        const sseServer = {
          name: 'SSE Server',
          protocol: 'sse' as const,
          url: 'https://api.example.com/events'
        };

        expect(() => serverSchemaWithProtocolValidation.parse(sseServer)).not.toThrow();
      });
    });

    describe('WebSocket protocol validation', () => {
      it('should require WebSocket URL', () => {
        const wsServer = {
          name: 'WebSocket Server',
          protocol: 'websocket' as const
        };

        expect(() => serverSchemaWithProtocolValidation.parse(wsServer)).toThrow();
      });

      it('should require ws:// or wss:// protocol', () => {
        const invalidWsServer = {
          name: 'WebSocket Server',
          protocol: 'websocket' as const,
          url: 'http://localhost:3000'
        };

        expect(() => serverSchemaWithProtocolValidation.parse(invalidWsServer)).toThrow();
      });

      it('should accept valid WebSocket URLs', () => {
        const wsServer = {
          name: 'WebSocket Server',
          protocol: 'websocket' as const,
          url: 'ws://localhost:3000'
        };

        const wssServer = {
          name: 'Secure WebSocket Server',
          protocol: 'websocket' as const,
          url: 'wss://api.example.com/ws'
        };

        expect(() => serverSchemaWithProtocolValidation.parse(wsServer)).not.toThrow();
        expect(() => serverSchemaWithProtocolValidation.parse(wssServer)).not.toThrow();
      });
    });

    describe('STDIO protocol validation', () => {
      it('should require command for STDIO protocol', () => {
        const stdioServer = {
          name: 'STDIO Server',
          protocol: 'stdio' as const
        };

        expect(() => serverSchemaWithProtocolValidation.parse(stdioServer)).toThrow();
      });

      it('should accept valid command for STDIO', () => {
        const stdioServer = {
          name: 'STDIO Server',
          protocol: 'stdio' as const,
          command: 'node server.js'
        };

        expect(() => serverSchemaWithProtocolValidation.parse(stdioServer)).not.toThrow();
      });
    });

    describe('Health check URL validation', () => {
      it('should validate health check URL matches protocol', () => {
        const httpServerWithHttpsHealth = {
          name: 'HTTP Server',
          protocol: 'http' as const,
          url: 'http://localhost:3000',
          health_check_url: 'https://localhost:3000/health'
        };

        expect(() => serverSchemaWithProtocolValidation.parse(httpServerWithHttpsHealth))
          .toThrow();
      });

      it('should accept compatible health check URLs', () => {
        const httpServer = {
          name: 'HTTP Server',
          protocol: 'http' as const,
          url: 'http://localhost:3000',
          health_check_url: 'http://localhost:3000/health'
        };

        const httpsServer = {
          name: 'HTTPS Server',
          protocol: 'https' as const,
          url: 'https://api.example.com',
          health_check_url: 'https://api.example.com/health'
        };

        const wsServer = {
          name: 'WebSocket Server',
          protocol: 'websocket' as const,
          url: 'ws://localhost:3000',
          health_check_url: 'http://localhost:3000/health'
        };

        expect(() => serverSchemaWithProtocolValidation.parse(httpServer)).not.toThrow();
        expect(() => serverSchemaWithProtocolValidation.parse(httpsServer)).not.toThrow();
        expect(() => serverSchemaWithProtocolValidation.parse(wsServer)).not.toThrow();
      });

      it('should allow no health check URL', () => {
        const server = {
          name: 'Test Server',
          protocol: 'http' as const,
          url: 'http://localhost:3000'
        };

        expect(() => serverSchemaWithProtocolValidation.parse(server)).not.toThrow();
      });
    });
  });

  describe('Schema compositions', () => {
    it('should create server with createServerSchema', () => {
      const createServer = {
        name: 'New Server',
        protocol: 'http' as const,
        url: 'http://localhost:3000'
      };

      expect(() => createServerSchema.parse(createServer)).not.toThrow();
    });

    it('should allow partial updates with updateServerSchema', () => {
      const partialUpdate = {
        name: 'Updated Server Name',
        timeout_seconds: 120
      };

      expect(() => updateServerSchema.parse(partialUpdate)).not.toThrow();
    });

    it('should validate partial updates', () => {
      const validPartialUpdate = {
        timeout_seconds: 120
      };

      expect(() => updateServerSchema.parse(validPartialUpdate)).not.toThrow();
    });
  });
});

describe('Environment Variable Validation', () => {
  describe('validateEnvironmentVariable', () => {
    it('should validate correct environment variables', () => {
      const validEnvVars = [
        'NODE_ENV=production',
        'PORT=3000',
        'API_KEY=secret-key-123',
        'DEBUG_MODE=true',
        'DATABASE_URL=postgresql://user:pass@host:5432/db'
      ];

      validEnvVars.forEach(envVar => {
        const result = validateEnvironmentVariable(envVar);
        expect(result.valid).toBe(true);
      });
    });

    it('should reject environment variables without equals sign', () => {
      const result = validateEnvironmentVariable('NODE_ENV_PRODUCTION');
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Environment variable must be in format KEY=VALUE');
    });

    it('should reject empty keys', () => {
      const results = [
        validateEnvironmentVariable('=value'),
        validateEnvironmentVariable(' =value'),
        validateEnvironmentVariable('\t=value')
      ];

      results.forEach(result => {
        expect(result.valid).toBe(false);
        expect(result.error).toBe('Environment variable key cannot be empty');
      });
    });

    it('should validate key format', () => {
      const invalidKeys = [
        '123VAR=value',
        'VAR-NAME=value',
        'var.name=value',
        'var name=value'
      ];

      invalidKeys.forEach(envVar => {
        const result = validateEnvironmentVariable(envVar);
        expect(result.valid).toBe(false);
        expect(result.error).toBe('Environment variable key can only contain letters, numbers, and underscores');
      });
    });

    it('should accept valid key formats', () => {
      const validKeys = [
        'VAR=value',
        'VAR_NAME=value',
        'var_name=value',
        '_PRIVATE_VAR=value',
        'VAR123=value',
        'VAR_123_ABC=value'
      ];

      validKeys.forEach(envVar => {
        const result = validateEnvironmentVariable(envVar);
        expect(result.valid).toBe(true);
      });
    });

    it('should handle values with equals signs', () => {
      const result = validateEnvironmentVariable('DATABASE_URL=postgresql://user:pass=secret@host/db');
      expect(result.valid).toBe(true);
    });

    it('should allow empty values', () => {
      const result = validateEnvironmentVariable('EMPTY_VAR=');
      expect(result.valid).toBe(true);
    });
  });
});

describe('Command Argument Validation', () => {
  describe('validateCommandArgument', () => {
    it('should validate correct command arguments', () => {
      const validArgs = [
        '--port',
        '3000',
        '--env=production',
        '/path/to/file',
        'simple-arg',
        'complex-arg_123'
      ];

      validArgs.forEach(arg => {
        const result = validateCommandArgument(arg);
        expect(result.valid).toBe(true);
      });
    });

    it('should reject empty arguments', () => {
      const emptyArgs = ['', '   ', '\t', '\n'];

      emptyArgs.forEach(arg => {
        const result = validateCommandArgument(arg);
        expect(result.valid).toBe(false);
        expect(result.error).toBe('Command argument cannot be empty');
      });
    });

    it('should reject too long arguments', () => {
      const longArg = 'x'.repeat(101);
      const result = validateCommandArgument(longArg);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Command argument must be less than 100 characters');
    });

    it('should accept arguments at the limit', () => {
      const limitArg = 'x'.repeat(100);
      const result = validateCommandArgument(limitArg);
      expect(result.valid).toBe(true);
    });

    it('should handle special characters in arguments', () => {
      const specialArgs = [
        '--config=/path/to/config.json',
        '"argument with spaces"',
        "'single quoted'",
        '--flag=value with spaces'
      ];

      specialArgs.forEach(arg => {
        const result = validateCommandArgument(arg);
        expect(result.valid).toBe(true);
      });
    });
  });
});

describe('Server Schema Edge Cases', () => {
  describe('Complex validation scenarios', () => {
    it('should handle complete server configuration', () => {
      const complexServer = {
        name: 'Production API Gateway',
        description: 'Main API gateway for production environment with load balancing and health monitoring',
        protocol: 'https' as const,
        url: 'https://api.production.example.com',
        command: 'node',
        args: ['dist/server.js', '--env=production', '--port=443', '--ssl=true'],
        environment: [
          'NODE_ENV=production',
          'PORT=443',
          'SSL_CERT_PATH=/certs/server.crt',
          'SSL_KEY_PATH=/certs/server.key',
          'DATABASE_URL=postgresql://user:password@db.example.com:5432/production',
          'REDIS_URL=redis://cache.example.com:6379',
          'LOG_LEVEL=info',
          'RATE_LIMIT_REQUESTS=1000',
          'RATE_LIMIT_WINDOW=3600'
        ],
        working_dir: '/app',
        version: '2.1.3',
        timeout_seconds: 120,
        max_retries: 5,
        status: 'active' as const,
        health_check_url: 'https://api.production.example.com/health',
        is_active: true,
        metadata: JSON.stringify({
          region: 'us-west-2',
          datacenter: 'aws-west-2a',
          loadBalancer: true,
          monitoring: {
            enabled: true,
            alerting: true,
            metrics: ['response_time', 'error_rate', 'throughput']
          }
        }),
        tags: ['production', 'api', 'gateway', 'https', 'load-balanced']
      };

      expect(() => serverSchemaWithProtocolValidation.parse(complexServer)).not.toThrow();
    });

    it('should handle minimal STDIO server', () => {
      const minimalStdio = {
        name: 'Simple CLI Tool',
        protocol: 'stdio' as const,
        command: 'python script.py'
      };

      expect(() => serverSchemaWithProtocolValidation.parse(minimalStdio)).not.toThrow();
    });

    it('should handle WebSocket server with HTTP health check', () => {
      const wsWithHealth = {
        name: 'Real-time Chat',
        protocol: 'websocket' as const,
        url: 'wss://chat.example.com/ws',
        health_check_url: 'https://chat.example.com/health'
      };

      expect(() => serverSchemaWithProtocolValidation.parse(wsWithHealth)).not.toThrow();
    });

    it('should validate mixed protocol scenarios', () => {
      const mixedScenarios = [
        {
          name: 'HTTP with HTTPS health',
          protocol: 'http' as const,
          url: 'http://localhost:3000',
          health_check_url: 'https://localhost:3001/health'
        },
        {
          name: 'SSE with wrong health protocol',
          protocol: 'sse' as const,
          url: 'http://events.example.com',
          health_check_url: 'ftp://events.example.com/status'
        }
      ];

      expect(() => serverSchemaWithProtocolValidation.parse(mixedScenarios[0])).toThrow();
      expect(() => serverSchemaWithProtocolValidation.parse(mixedScenarios[1])).toThrow();
    });
  });
});
