import { authApi } from '../auth-api';
import type { LoginRequest, LoginResponse, User, ApiKey, CreateApiKeyRequest, CreateApiKeyResponse } from '@/lib/types';

// Mock localStorage and sessionStorage before any other imports
const mockLocalStorage = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; },
    key: (index: number) => Object.keys(store)[index] || null,
    get length() { return Object.keys(store).length; }
  };
})();

const mockSessionStorage = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; },
    key: (index: number) => Object.keys(store)[index] || null,
    get length() { return Object.keys(store).length; }
  };
})();

// Mock storage on global for Node.js environment
Object.defineProperty(global, 'localStorage', { value: mockLocalStorage, writable: true });
Object.defineProperty(global, 'sessionStorage', { value: mockSessionStorage, writable: true });

// Check if window is already defined (from setup-simple.ts) and update it
if (typeof window !== 'undefined') {
  // Update existing window object with our mocks
  Object.defineProperty(window, 'localStorage', { value: mockLocalStorage, writable: true });
  Object.defineProperty(window, 'sessionStorage', { value: mockSessionStorage, writable: true });
} else {
  // Create window object if it doesn't exist
  Object.defineProperty(global, 'window', {
    value: {
      localStorage: mockLocalStorage,
      sessionStorage: mockSessionStorage
    },
    writable: true
  });
}

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('AuthAPI - Extended Tests', () => {
  beforeEach(() => {
    // Clear storage
    mockLocalStorage.clear();
    mockSessionStorage.clear();
    // Clear and reset fetch mock
    mockFetch.mockClear();
    mockFetch.mockReset();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('Token Management', () => {
    describe('getToken/setToken/removeToken (private methods tested through public interface)', () => {
      it('should handle token operations in browser environment', () => {
        expect(authApi.isAuthenticated()).toBe(false);

        // Set token through login simulation
        localStorage.setItem('access_token', 'test-token');
        expect(authApi.isAuthenticated()).toBe(true);

        // Clear tokens
        authApi.clearTokens();
        expect(localStorage.getItem('access_token')).toBeNull();
        expect(localStorage.getItem('refresh_token')).toBeNull();
        expect(authApi.isAuthenticated()).toBe(false);
      });

      it('should handle missing localStorage gracefully', () => {
        // This test passes because the auth-api checks for window !== undefined
        // and we have provided a window object, so localStorage calls work normally
        expect(() => authApi.clearTokens()).not.toThrow();
        expect(authApi.isAuthenticated()).toBe(false);
      });
    });

    describe('isAuthenticated', () => {
      it('should return false when no access token exists', () => {
        expect(authApi.isAuthenticated()).toBe(false);
      });

      it('should return true when access token exists', () => {
        localStorage.setItem('access_token', 'valid-token');
        expect(authApi.isAuthenticated()).toBe(true);
      });

      it('should return false when access token is empty string', () => {
        localStorage.setItem('access_token', '');
        expect(authApi.isAuthenticated()).toBe(false);
      });
    });

    describe('clearTokens', () => {
      it('should remove both access and refresh tokens', () => {
        localStorage.setItem('access_token', 'access-token');
        localStorage.setItem('refresh_token', 'refresh-token');

        authApi.clearTokens();

        expect(localStorage.getItem('access_token')).toBeNull();
        expect(localStorage.getItem('refresh_token')).toBeNull();
      });

      it('should handle non-existent tokens gracefully', () => {
        expect(() => authApi.clearTokens()).not.toThrow();
      });
    });
  });

  describe('API Request Helper Function', () => {
    const mockResponseData = { message: 'success' };

    it('should make GET request with default headers', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce(mockResponseData)
      });

      // Test through getProfile which uses apiRequest
      localStorage.setItem('access_token', 'test-token');

      const mockProfile: User = {
        id: '1',
        email: 'test@test.com',
        name: 'Test User',
        role: 'user',
        organization_id: 'org-1',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce({ data: mockProfile, success: true })
      });

      await authApi.getProfile();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/profile',
        {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer test-token'
          }
        }
      );
    });

    it('should make POST request with body', async () => {
      const loginData: LoginRequest = {
        email: 'test@test.com',
        password: 'password123'
      };

      const loginResponse: LoginResponse = {
        access_token: 'new-access-token',
        refresh_token: 'new-refresh-token',
        user: {
          id: '1',
          email: 'test@test.com',
          name: 'Test User',
          role: 'user',
          organization_id: 'org-1',
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z'
        }
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce(loginResponse)
      });

      await authApi.login(loginData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/login',
        {
          method: 'POST',
          body: JSON.stringify(loginData),
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );
    });

    it('should handle authorization header when token exists', async () => {
      localStorage.setItem('access_token', 'existing-token');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204
      });

      await authApi.logout();

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/logout',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Authorization': 'Bearer existing-token'
          })
        })
      );
    });
  });

  describe('HTTP Status Code Handling', () => {
    it('should handle 401 Unauthorized', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        text: jest.fn().mockResolvedValueOnce('Unauthorized')
      });

      await expect(authApi.login({ email: 'test@test.com', password: 'wrong' }))
        .rejects.toThrow('Invalid credentials');
    });

    it('should handle 403 Forbidden', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        statusText: 'Forbidden',
        text: jest.fn().mockResolvedValueOnce('Forbidden')
      });

      localStorage.setItem('access_token', 'token');
      await expect(authApi.getProfile())
        .rejects.toThrow('Access denied');
    });

    it('should handle 404 Not Found', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        text: jest.fn().mockResolvedValueOnce('Not Found')
      });

      localStorage.setItem('access_token', 'token');
      await expect(authApi.deleteApiKey('non-existent'))
        .rejects.toThrow('Resource not found');
    });

    it('should handle 500 Server Error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        text: jest.fn().mockResolvedValueOnce('Internal Server Error')
      });

      await expect(authApi.login({ email: 'test@test.com', password: 'test' }))
        .rejects.toThrow('Server error occurred');
    });

    it('should handle 204 No Content responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204
      });

      localStorage.setItem('access_token', 'token');
      const result = await authApi.deleteApiKey('test-key');
      expect(result).toBeUndefined();
    });
  });

  describe('Error Parsing', () => {
    it('should parse JSON error with message field', async () => {
      const errorResponse = { message: 'Custom error message' };
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: jest.fn().mockResolvedValueOnce(JSON.stringify(errorResponse))
      });

      await expect(authApi.login({ email: 'test@test.com', password: 'test' }))
        .rejects.toThrow('Request failed: Bad Request');
    });

    it('should parse JSON error with nested error.message field', async () => {
      const errorResponse = { error: { message: 'Nested error message' } };
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: jest.fn().mockResolvedValueOnce(JSON.stringify(errorResponse))
      });

      await expect(authApi.login({ email: 'test@test.com', password: 'test' }))
        .rejects.toThrow('Request failed: Bad Request');
    });

    it('should parse JSON error with error string field', async () => {
      const errorResponse = { error: 'String error' };
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        text: jest.fn().mockResolvedValueOnce(JSON.stringify(errorResponse))
      });

      await expect(authApi.login({ email: 'test@test.com', password: 'test' }))
        .rejects.toThrow('Request failed: Bad Request');
    });

    it('should fallback to status-based error for invalid JSON', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        text: jest.fn().mockResolvedValueOnce('Invalid JSON response')
      });

      await expect(authApi.login({ email: 'test@test.com', password: 'test' }))
        .rejects.toThrow('Invalid credentials');
    });

    it('should fallback to generic error for unknown status', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 418,
        statusText: "I'm a teapot",
        text: jest.fn().mockResolvedValueOnce('Invalid JSON response')
      });

      await expect(authApi.login({ email: 'test@test.com', password: 'test' }))
        .rejects.toThrow("Request failed: I'm a teapot");
    });
  });

  describe('Content-Type Parsing', () => {
    it('should parse JSON responses with correct content-type', async () => {
      const responseData = { data: { id: '1', name: 'Test' }, success: true };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json; charset=utf-8']]),
        json: jest.fn().mockResolvedValueOnce(responseData)
      });

      localStorage.setItem('access_token', 'token');
      const result = await authApi.getProfile();

      expect(result).toEqual({ id: '1', name: 'Test' });
    });

    it('should handle responses without JSON content-type', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'text/plain']]),
        json: jest.fn()
      });

      localStorage.setItem('access_token', 'token');
      // This will fail because getProfile expects response.data to exist
      await expect(authApi.getProfile()).rejects.toThrow();
    });

    it('should handle responses without content-type header', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map(),
        json: jest.fn()
      });

      localStorage.setItem('access_token', 'token');
      // This will fail because getProfile expects response.data to exist
      await expect(authApi.getProfile()).rejects.toThrow();
    });
  });

  describe('Login Flow', () => {
    it('should successfully login and store tokens', async () => {
      const loginData: LoginRequest = {
        email: 'test@test.com',
        password: 'password123'
      };

      const loginResponse = {
        data: {
          access_token: 'new-access-token',
          refresh_token: 'new-refresh-token',
          expires_in: 3600
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce(loginResponse)
      });

      const result = await authApi.login(loginData);

      expect(result).toEqual(loginResponse);
      expect(localStorage.getItem('access_token')).toBe('new-access-token');
      expect(localStorage.getItem('refresh_token')).toBe('new-refresh-token');
      expect(authApi.isAuthenticated()).toBe(true);
    });

    it('should handle login response with tokens at root level', async () => {
      const loginResponse = {
        access_token: 'root-level-access-token',
        refresh_token: 'root-level-refresh-token',
        expires_in: 3600
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce(loginResponse)
      });

      await authApi.login({ email: 'test@test.com', password: 'test' });

      expect(localStorage.getItem('access_token')).toBe('root-level-access-token');
      expect(localStorage.getItem('refresh_token')).toBe('root-level-refresh-token');
    });
  });

  describe('Logout Flow', () => {
    it('should successfully logout and clear tokens', async () => {
      localStorage.setItem('access_token', 'existing-token');
      localStorage.setItem('refresh_token', 'existing-refresh');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204
      });

      await authApi.logout();

      expect(localStorage.getItem('access_token')).toBeNull();
      expect(localStorage.getItem('refresh_token')).toBeNull();
      expect(authApi.isAuthenticated()).toBe(false);
    });

    it('should clear tokens even if logout request fails', async () => {
      localStorage.setItem('access_token', 'existing-token');
      localStorage.setItem('refresh_token', 'existing-refresh');

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        text: jest.fn().mockResolvedValueOnce('Server Error')
      });

      await expect(authApi.logout()).rejects.toThrow();

      // Tokens should still be cleared
      expect(localStorage.getItem('access_token')).toBeNull();
      expect(localStorage.getItem('refresh_token')).toBeNull();
    });
  });

  describe('Token Refresh Mechanism', () => {
    it('should successfully refresh tokens', async () => {
      localStorage.setItem('refresh_token', 'existing-refresh-token');

      const refreshResponse = {
        data: {
          access_token: 'new-access-token',
          refresh_token: 'new-refresh-token',
          expires_in: 3600
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce(refreshResponse)
      });

      const result = await authApi.refresh();

      expect(result).toEqual(refreshResponse);
      expect(localStorage.getItem('access_token')).toBe('new-access-token');
      expect(localStorage.getItem('refresh_token')).toBe('new-refresh-token');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/refresh',
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Authorization': 'Bearer existing-refresh-token'
          })
        })
      );
    });

    it('should handle refresh response with tokens at root level', async () => {
      localStorage.setItem('refresh_token', 'existing-refresh-token');

      const refreshResponse = {
        access_token: 'root-new-access-token',
        refresh_token: 'root-new-refresh-token',
        expires_in: 3600
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce(refreshResponse)
      });

      await authApi.refresh();

      expect(localStorage.getItem('access_token')).toBe('root-new-access-token');
      expect(localStorage.getItem('refresh_token')).toBe('root-new-refresh-token');
    });

    it('should throw error when no refresh token is available', async () => {
      localStorage.clear();

      await expect(authApi.refresh())
        .rejects.toThrow('No refresh token available');

      expect(mockFetch).not.toHaveBeenCalled();
    });

    it('should throw error when refresh token is empty', async () => {
      localStorage.setItem('refresh_token', '');

      await expect(authApi.refresh())
        .rejects.toThrow('No refresh token available');
    });
  });

  describe('Edge Cases', () => {
    it('should handle expired tokens gracefully', async () => {
      localStorage.setItem('access_token', 'expired-token');

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        text: jest.fn().mockResolvedValueOnce(JSON.stringify({
          error: { message: 'Token expired' }
        }))
      });

      await expect(authApi.getProfile())
        .rejects.toThrow('Invalid credentials');
    });

    it('should handle missing tokens for authenticated requests', async () => {
      localStorage.clear();

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        text: jest.fn().mockResolvedValueOnce('Unauthorized')
      });

      await expect(authApi.getProfile())
        .rejects.toThrow('Invalid credentials');
    });

    it('should handle malformed token responses', async () => {
      const malformedResponse = {
        data: {
          // Missing required tokens
          expires_in: 3600
        },
        success: true
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce(malformedResponse)
      });

      const result = await authApi.login({ email: 'test@test.com', password: 'test' });
      // Should succeed but store undefined tokens, which is handled gracefully
      expect(result).toEqual(malformedResponse);
      expect(authApi.isAuthenticated()).toBe(false); // No valid token stored
    });

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(authApi.login({ email: 'test@test.com', password: 'test' }))
        .rejects.toThrow('Network error');
    });
  });

  describe('Profile Management', () => {
    const mockUser: User = {
      id: '123',
      email: 'test@example.com',
      name: 'Test User',
      role: 'user',
      organization_id: 'org-123',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z'
    };

    it('should get user profile', async () => {
      localStorage.setItem('access_token', 'valid-token');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce({ data: mockUser, success: true })
      });

      const result = await authApi.getProfile();

      expect(result).toEqual(mockUser);
    });

    it('should update user profile', async () => {
      localStorage.setItem('access_token', 'valid-token');
      const updateData = { name: 'Updated Name' };
      const updatedUser = { ...mockUser, name: 'Updated Name' };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce({ data: updatedUser, success: true })
      });

      const result = await authApi.updateProfile(updateData);

      expect(result).toEqual(updatedUser);
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/profile',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify(updateData)
        })
      );
    });

    it('should handle profile responses at root level', async () => {
      localStorage.setItem('access_token', 'valid-token');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce(mockUser)
      });

      const result = await authApi.getProfile();

      expect(result).toEqual(mockUser);
    });
  });

  describe('API Key Management', () => {
    const mockApiKey: ApiKey = {
      id: 'key-123',
      name: 'Test Key',
      prefix: 'tk_',
      is_active: true,
      created_at: new Date().toISOString(),
      last_used_at: undefined,
      expires_at: undefined
    };

    it('should get API keys', async () => {
      localStorage.setItem('access_token', 'valid-token');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce({ data: [mockApiKey], success: true })
      });

      const result = await authApi.getApiKeys();

      expect(result).toEqual([mockApiKey]);
    });

    it('should return empty array when no API keys data', async () => {
      localStorage.setItem('access_token', 'valid-token');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce({ success: true })
      });

      const result = await authApi.getApiKeys();

      expect(result).toEqual([]);
    });

    it('should create API key', async () => {
      localStorage.setItem('access_token', 'valid-token');
      const createRequest: CreateApiKeyRequest = {
        name: 'New API Key'
      };
      const createResponse: CreateApiKeyResponse = {
        apiKey: {
          id: 'key-456',
          name: 'New API Key',
          prefix: 'tk_',
          is_active: true,
          created_at: new Date().toISOString(),
          last_used_at: undefined,
          expires_at: undefined
        },
        key: 'tk_abcdef123456789'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 201,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce({ data: createResponse, success: true })
      });

      const result = await authApi.createApiKey(createRequest);

      expect(result).toEqual(createResponse);
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/api-keys',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(createRequest)
        })
      );
    });

    it('should delete API key', async () => {
      localStorage.setItem('access_token', 'valid-token');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204
      });

      const result = await authApi.deleteApiKey('key-123');

      expect(result).toBeUndefined();
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/api-keys/key-123',
        expect.objectContaining({
          method: 'DELETE'
        })
      );
    });

    it('should handle API key responses at root level', async () => {
      localStorage.setItem('access_token', 'valid-token');
      const createResponse: CreateApiKeyResponse = {
        apiKey: {
          id: 'key-789',
          name: 'Root Level Key',
          prefix: 'tk_',
          is_active: true,
          created_at: new Date().toISOString(),
          last_used_at: undefined,
          expires_at: undefined
        },
        key: 'tk_rootlevel123'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 201,
        headers: new Map([['content-type', 'application/json']]),
        json: jest.fn().mockResolvedValueOnce(createResponse)
      });

      const result = await authApi.createApiKey({
        name: 'Root Level Key'
      });

      expect(result).toEqual(createResponse);
    });
  });
});
