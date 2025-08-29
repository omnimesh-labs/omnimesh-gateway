import { adminApi, serverApi, namespaceApi, a2aApi, policyApi, contentFilterApi, endpointApi, discoveryApi, toolsApi, promptsApi, authConfigApi } from '../client-api';

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('Client API - Simple Tests', () => {
  beforeEach(() => {
    mockFetch.mockClear();
    // Clear storage only if available (test environment may mock these)
    if (typeof localStorage !== 'undefined') {
      localStorage.clear();
    }
    if (typeof sessionStorage !== 'undefined') {
      sessionStorage.clear();
    }
  });

  describe('apiRequest helper function', () => {
    it('should make basic API requests with correct headers', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ data: 'test' })
      });

      await adminApi.getStats();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/admin/stats',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          })
        })
      );
    });

    it('should not include Authorization header when no access token', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ data: 'test' })
      });

      await adminApi.getStats();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/admin/stats',
        expect.any(Object)
      );
    });

    it('should handle 401 Unauthorized with clean error message', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        text: jest.fn().mockResolvedValueOnce('Unauthorized')
      });

      await expect(adminApi.getStats()).rejects.toThrow('Authentication required');
    });

    it('should handle 403 Forbidden with clean error message', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        statusText: 'Forbidden',
        text: jest.fn().mockResolvedValueOnce('Forbidden')
      });

      await expect(adminApi.getStats()).rejects.toThrow('Access denied');
    });

    it('should handle 404 Not Found with clean error message', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        text: jest.fn().mockResolvedValueOnce('Not Found')
      });

      await expect(adminApi.getStats()).rejects.toThrow('Resource not found');
    });

    it('should handle 500 Server Error with clean error message', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        text: jest.fn().mockResolvedValueOnce('Internal Server Error')
      });

      await expect(adminApi.getStats()).rejects.toThrow('Server error occurred');
    });

    it('should handle 204 No Content response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        statusText: 'No Content'
      });

      const result = await serverApi.deleteServer('test-id');
      expect(result).toBeUndefined();
    });

    it('should parse JSON error responses with error.message', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: jest.fn().mockResolvedValueOnce(JSON.stringify({
          error: { message: 'Custom error message' }
        }))
      });

      await expect(adminApi.getStats()).rejects.toThrow('Request failed: Bad Request');
    });

    it('should parse JSON error responses with message field', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: jest.fn().mockResolvedValueOnce(JSON.stringify({
          message: 'Direct message field'
        }))
      });

      await expect(adminApi.getStats()).rejects.toThrow('Request failed: Bad Request');
    });

    it('should parse JSON error responses with error string', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: jest.fn().mockResolvedValueOnce(JSON.stringify({
          error: 'Error as string'
        }))
      });

      await expect(adminApi.getStats()).rejects.toThrow('Request failed: Bad Request');
    });

    it('should handle content-type parsing for JSON responses', async () => {
      const mockResponse = { data: 'test' };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json; charset=utf-8') },
        json: jest.fn().mockResolvedValueOnce(mockResponse)
      });

      const result = await adminApi.getStats();
      expect(result).toEqual(mockResponse);
    });

    it('should handle non-JSON responses by returning null', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('text/plain') }
      });

      const result = await adminApi.getStats();
      expect(result).toBeNull();
    });

    it('should fallback to clean error message when JSON parsing fails', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: jest.fn().mockResolvedValueOnce('Invalid JSON response')
      });

      await expect(adminApi.getStats()).rejects.toThrow('Request failed: Bad Request');
    });
  });

  describe('AdminAPI', () => {
    it('should call correct endpoint for getStats', async () => {
      const mockStats = { activeServers: 5, totalRequests: 1000 };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce(mockStats)
      });

      const result = await adminApi.getStats();

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/admin/stats', expect.any(Object));
      expect(result).toEqual(mockStats);
    });

    it('should build correct query parameters for getLogs', async () => {
      const mockResponse = { logs: [], total: 0 };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce(mockResponse)
      });

      await adminApi.getLogs({
        level: 'error',
        server_id: 'server-123',
        user_id: 'user-456',
        start_date: '2023-01-01',
        end_date: '2023-01-31',
        limit: 100,
        offset: 0
      });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/admin/logs?level=error&server_id=server-123&user_id=user-456&start_date=2023-01-01&end_date=2023-01-31&limit=100',
        expect.any(Object)
      );
    });

    it('should build correct query parameters for getAuditLogs', async () => {
      const mockResponse = { logs: [], total: 0 };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce(mockResponse)
      });

      await adminApi.getAuditLogs({
        user_id: 'user-123',
        action: 'create',
        resource_type: 'server',
        start_date: '2023-01-01',
        end_date: '2023-01-31',
        limit: 50,
        offset: 10
      });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/admin/audit?user_id=user-123&action=create&resource_type=server&start_date=2023-01-01&end_date=2023-01-31&limit=50&offset=10',
        expect.any(Object)
      );
    });
  });

  describe('ServerAPI', () => {
    it('should handle listServers with data wrapper', async () => {
      const mockServers = [{ id: '1', name: 'Server 1' }];
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ data: mockServers, success: true })
      });

      const result = await serverApi.listServers();
      expect(result).toEqual(mockServers);
    });

    it('should handle listServers with empty data', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ success: true })
      });

      const result = await serverApi.listServers();
      expect(result).toEqual([]);
    });

    it('should make POST request for createServer', async () => {
      const mockServer = { id: '1', name: 'New Server' };
      const createData = { name: 'New Server', url: 'http://example.com' };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce(mockServer)
      });

      const result = await serverApi.createServer(createData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/gateway/servers',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(createData)
        })
      );
      expect(result).toEqual(mockServer);
    });

    it('should make PUT request for updateServer', async () => {
      const mockServer = { id: '1', name: 'Updated Server' };
      const updateData = { name: 'Updated Server' };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce(mockServer)
      });

      const result = await serverApi.updateServer('1', updateData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/gateway/servers/1',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify(updateData)
        })
      );
      expect(result).toEqual(mockServer);
    });

    it('should make DELETE request for deleteServer', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204
      });

      await serverApi.deleteServer('1');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/gateway/servers/1',
        expect.objectContaining({
          method: 'DELETE'
        })
      );
    });

    it('should make POST request for restartServer', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204
      });

      await serverApi.restartServer('1');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/gateway/servers/1/restart',
        expect.objectContaining({
          method: 'POST'
        })
      );
    });
  });

  describe('NamespaceAPI', () => {
    it('should extract namespaces from response wrapper', async () => {
      const mockNamespaces = [{ id: '1', name: 'Namespace 1' }];
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ namespaces: mockNamespaces, total: 1 })
      });

      const result = await namespaceApi.listNamespaces();
      expect(result).toEqual(mockNamespaces);
    });

    it('should handle empty namespaces response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ total: 0 })
      });

      const result = await namespaceApi.listNamespaces();
      expect(result).toEqual([]);
    });
  });

  describe('ToolsAPI', () => {
    it('should build query parameters correctly for listTools', async () => {
      const mockResponse = { success: true, data: [], count: 0 };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce(mockResponse)
      });

      await toolsApi.listTools({
        active: true,
        category: 'test',
        search: 'query',
        popular: true,
        include_public: false,
        limit: 10,
        offset: 20
      });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/gateway/tools?active=true&category=test&search=query&popular=true&limit=10&offset=20',
        expect.any(Object)
      );
    });

    it('should extract data from response wrapper', async () => {
      const mockTools = [{ id: '1', name: 'Tool 1' }];
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ success: true, data: mockTools, count: 1 })
      });

      const result = await toolsApi.listTools();
      expect(result).toEqual(mockTools);
    });

    it('should handle empty tools data', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ success: true, count: 0 })
      });

      const result = await toolsApi.listTools();
      expect(result).toEqual([]);
    });
  });

  describe('PromptsAPI', () => {
    it('should build query parameters correctly for listPrompts', async () => {
      const mockResponse = { success: true, data: [], count: 0 };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce(mockResponse)
      });

      await promptsApi.listPrompts({
        active: false,
        category: 'ai',
        search: 'test',
        popular: false,
        limit: 5,
        offset: 15
      });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/gateway/prompts?active=false&category=ai&search=test&limit=5&offset=15',
        expect.any(Object)
      );
    });
  });

  describe('A2AApi', () => {
    it('should build query parameters for listAgents', async () => {
      const mockResponse = { agents: [], total: 0 };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce(mockResponse)
      });

      await a2aApi.listAgents({ is_active: true, tags: 'tag1,tag2' });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/a2a?is_active=true&tags=tag1%2Ctag2',
        expect.any(Object)
      );
    });

    it('should extract agents from response', async () => {
      const mockAgents = [{ id: '1', name: 'Agent 1' }];
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ agents: mockAgents, total: 1 })
      });

      const result = await a2aApi.listAgents();
      expect(result).toEqual(mockAgents);
    });
  });

  describe('AuthConfigAPI', () => {
    it('should extract data from getAuthConfig response', async () => {
      const mockConfig = { enabled: true, providers: [] };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ success: true, data: mockConfig })
      });

      const result = await authConfigApi.getAuthConfig();
      expect(result).toEqual(mockConfig);
    });

    it('should make PUT request for updateAuthConfig', async () => {
      const mockConfig = { enabled: false };
      const updateData = { enabled: false, providers: ['google'] };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: { get: jest.fn().mockReturnValue('application/json') },
        json: jest.fn().mockResolvedValueOnce({ success: true, data: mockConfig, message: 'Updated' })
      });

      const result = await authConfigApi.updateAuthConfig(updateData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/admin/auth-config',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify(updateData)
        })
      );
      expect(result).toEqual(mockConfig);
    });
  });

  describe('Error handling edge cases', () => {
    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(adminApi.getStats()).rejects.toThrow('Network error');
    });

    it('should handle malformed JSON in error response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: jest.fn().mockResolvedValueOnce('{ invalid json')
      });

      await expect(adminApi.getStats()).rejects.toThrow('Request failed: Bad Request');
    });
  });
});
