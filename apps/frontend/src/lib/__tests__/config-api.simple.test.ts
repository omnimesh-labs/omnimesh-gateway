import { configApi } from '../config-api';

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('Configuration API - Simple Tests', () => {
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

  describe('ConfigAPI class', () => {
    it('should be properly instantiated', () => {
      expect(configApi).toBeDefined();
      expect(configApi.exportConfiguration).toBeInstanceOf(Function);
      expect(configApi.importConfiguration).toBeInstanceOf(Function);
    });
  });

  describe('exportConfiguration', () => {
    it('should make POST request to correct endpoint', async () => {
      const mockBlob = new Blob(['test-data'], { type: 'application/json' });
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        blob: jest.fn().mockResolvedValueOnce(mockBlob)
      });

      const options = { includeSecrets: false, format: 'json' };
      const result = await configApi.exportConfiguration(options);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/admin/export',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(options)
        })
      );
      expect(result).toBeInstanceOf(Blob);
      expect(result).toBe(mockBlob);
    });

    it('should handle empty options object', async () => {
      const mockBlob = new Blob(['empty-config'], { type: 'application/json' });
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        blob: jest.fn().mockResolvedValueOnce(mockBlob)
      });

      const result = await configApi.exportConfiguration({});

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/admin/export',
        expect.objectContaining({
          body: JSON.stringify({})
        })
      );
      expect(result).toBe(mockBlob);
    });

    it('should handle complex options object', async () => {
      const mockBlob = new Blob(['complex-config'], { type: 'application/json' });
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        blob: jest.fn().mockResolvedValueOnce(mockBlob)
      });

      const complexOptions = {
        includeSecrets: true,
        format: 'yaml',
        sections: ['servers', 'policies', 'users'],
        compression: 'gzip',
        metadata: {
          exportedBy: 'admin',
          timestamp: '2023-01-01T00:00:00Z'
        }
      };

      const result = await configApi.exportConfiguration(complexOptions);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/admin/export',
        expect.objectContaining({
          body: JSON.stringify(complexOptions)
        })
      );
      expect(result).toBe(mockBlob);
    });

    it('should throw error when export fails with 400 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request'
      });

      await expect(configApi.exportConfiguration({})).rejects.toThrow(
        'Export failed: 400 Bad Request'
      );
    });

    it('should throw error when export fails with 401 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized'
      });

      await expect(configApi.exportConfiguration({})).rejects.toThrow(
        'Export failed: 401 Unauthorized'
      );
    });

    it('should throw error when export fails with 403 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        statusText: 'Forbidden'
      });

      await expect(configApi.exportConfiguration({})).rejects.toThrow(
        'Export failed: 403 Forbidden'
      );
    });

    it('should throw error when export fails with 500 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error'
      });

      await expect(configApi.exportConfiguration({})).rejects.toThrow(
        'Export failed: 500 Internal Server Error'
      );
    });

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(configApi.exportConfiguration({})).rejects.toThrow('Network error');
    });

    it('should handle blob() method throwing error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        blob: jest.fn().mockRejectedValueOnce(new Error('Blob parsing error'))
      });

      await expect(configApi.exportConfiguration({})).rejects.toThrow('Blob parsing error');
    });
  });

  describe('importConfiguration', () => {
    it('should make POST request to correct endpoint with FormData', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200
      });

      const formData = new FormData();
      formData.append('file', new Blob(['config-data']), 'config.json');
      formData.append('options', JSON.stringify({ replaceExisting: true }));

      await configApi.importConfiguration(formData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/admin/import',
        expect.objectContaining({
          method: 'POST',
          body: formData
        })
      );
    });

    it('should handle empty FormData', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200
      });

      const emptyFormData = new FormData();
      await configApi.importConfiguration(emptyFormData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/admin/import',
        expect.objectContaining({
          method: 'POST',
          body: emptyFormData
        })
      );
    });

    it('should handle FormData with multiple files', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200
      });

      const formData = new FormData();
      formData.append('mainConfig', new Blob(['main-config']), 'main.json');
      formData.append('secrets', new Blob(['secret-config']), 'secrets.json');
      formData.append('backup', 'true');

      await configApi.importConfiguration(formData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/admin/import',
        expect.objectContaining({
          method: 'POST',
          body: formData
        })
      );
    });

    it('should not include Content-Type header (let browser set it)', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200
      });

      const formData = new FormData();
      await configApi.importConfiguration(formData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/admin/import',
        expect.objectContaining({
          method: 'POST',
          body: formData
        })
      );

      // Ensure Content-Type header is not explicitly set (browser will set multipart/form-data)
      const callArgs = mockFetch.mock.calls[0][1];
      expect(callArgs.headers).toBeUndefined();
    });

    it('should throw error when import fails with 400 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request'
      });

      const formData = new FormData();
      await expect(configApi.importConfiguration(formData)).rejects.toThrow(
        'Import failed: 400 Bad Request'
      );
    });

    it('should throw error when import fails with 401 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized'
      });

      const formData = new FormData();
      await expect(configApi.importConfiguration(formData)).rejects.toThrow(
        'Import failed: 401 Unauthorized'
      );
    });

    it('should throw error when import fails with 403 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        statusText: 'Forbidden'
      });

      const formData = new FormData();
      await expect(configApi.importConfiguration(formData)).rejects.toThrow(
        'Import failed: 403 Forbidden'
      );
    });

    it('should throw error when import fails with 413 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 413,
        statusText: 'Payload Too Large'
      });

      const formData = new FormData();
      await expect(configApi.importConfiguration(formData)).rejects.toThrow(
        'Import failed: 413 Payload Too Large'
      );
    });

    it('should throw error when import fails with 422 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 422,
        statusText: 'Unprocessable Entity'
      });

      const formData = new FormData();
      await expect(configApi.importConfiguration(formData)).rejects.toThrow(
        'Import failed: 422 Unprocessable Entity'
      );
    });

    it('should throw error when import fails with 500 status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error'
      });

      const formData = new FormData();
      await expect(configApi.importConfiguration(formData)).rejects.toThrow(
        'Import failed: 500 Internal Server Error'
      );
    });

    it('should handle network errors during import', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network connection failed'));

      const formData = new FormData();
      await expect(configApi.importConfiguration(formData)).rejects.toThrow(
        'Network connection failed'
      );
    });

    it('should return void on successful import', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200
      });

      const formData = new FormData();
      const result = await configApi.importConfiguration(formData);

      expect(result).toBeUndefined();
    });

    it('should handle 201 status (Created) as success', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 201
      });

      const formData = new FormData();
      const result = await configApi.importConfiguration(formData);

      expect(result).toBeUndefined();
    });

    it('should handle 204 status (No Content) as success', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204
      });

      const formData = new FormData();
      const result = await configApi.importConfiguration(formData);

      expect(result).toBeUndefined();
    });
  });

  describe('API Base URL configuration', () => {
    it('should use consistent API base URL for all requests', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        blob: jest.fn().mockResolvedValueOnce(new Blob(['test']))
      });

      await configApi.exportConfiguration({});

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/admin/export'),
        expect.any(Object)
      );
    });
  });

  describe('Error handling edge cases', () => {
    it('should handle fetch returning undefined response', async () => {
      mockFetch.mockResolvedValueOnce(undefined);

      await expect(configApi.exportConfiguration({})).rejects.toThrow();
    });

    it('should handle response without ok property', async () => {
      mockFetch.mockResolvedValueOnce({
        status: 200,
        blob: jest.fn().mockResolvedValueOnce(new Blob(['test']))
      });

      await expect(configApi.exportConfiguration({})).rejects.toThrow();
    });

    it('should handle response without status property', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Unknown Error'
      });

      await expect(configApi.exportConfiguration({})).rejects.toThrow(
        'Export failed: undefined Unknown Error'
      );
    });

    it('should handle response without statusText property', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500
      });

      await expect(configApi.exportConfiguration({})).rejects.toThrow(
        'Export failed: 500 undefined'
      );
    });
  });
});
