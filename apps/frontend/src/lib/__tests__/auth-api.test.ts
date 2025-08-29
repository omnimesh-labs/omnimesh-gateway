import { authApi } from '../auth-api';
import { server } from '../../test/server';
import { http, HttpResponse } from 'msw';

// Mock localStorage is already set up in test setup
describe('AuthAPI', () => {
  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear();
    sessionStorage.clear();
  });

  describe('login', () => {
    it('should extract tokens from wrapped response structure', async () => {
      const credentials = { email: 'test@example.com', password: 'testpassword123' };

      const response = await authApi.login(credentials);

      // Verify response structure
      expect(response).toHaveProperty('data');
      expect(response.data).toHaveProperty('access_token');
      expect(response.data).toHaveProperty('refresh_token');
      expect(response.data).toHaveProperty('user');

      // Verify tokens are stored in localStorage
      expect(localStorage.getItem('access_token')).toBe('mock-access-token-12345');
      expect(localStorage.getItem('refresh_token')).toBe('mock-refresh-token-67890');
    });

    it('should handle login with invalid credentials', async () => {
      const credentials = { email: 'test@example.com', password: 'wrongpassword' };

      await expect(authApi.login(credentials)).rejects.toThrow('Invalid credentials');

      // Verify no tokens are stored
      expect(localStorage.getItem('access_token')).toBeNull();
      expect(localStorage.getItem('refresh_token')).toBeNull();
    });

    it('should handle backend response structure changes gracefully', async () => {
      // Mock a response without the data wrapper (fallback case)
      server.use(
        http.post('http://localhost:8080/api/auth/login', () => {
          return HttpResponse.json({
            access_token: 'direct-access-token',
            refresh_token: 'direct-refresh-token',
            token_type: 'Bearer',
            expires_in: 900,
            user: {
              id: '1',
              email: 'test@example.com',
              name: 'Test User'
            }
          }, { status: 200 });
        })
      );

      const credentials = { email: 'test@example.com', password: 'testpassword123' };

      const response = await authApi.login(credentials);

      // Verify tokens are extracted from root level (fallback)
      expect(localStorage.getItem('access_token')).toBe('direct-access-token');
      expect(localStorage.getItem('refresh_token')).toBe('direct-refresh-token');
    });

    it('should handle empty response gracefully', async () => {
      // Mock an empty successful response
      server.use(
        http.post('http://localhost:8080/api/auth/login', () => {
          return HttpResponse.json({}, { status: 200 });
        })
      );

      const credentials = { email: 'test@example.com', password: 'testpassword123' };

      // Should not throw, but tokens should be undefined
      const response = await authApi.login(credentials);

      // Verify tokens are not set when not present in response
      expect(localStorage.getItem('access_token')).toBeNull();
      expect(localStorage.getItem('refresh_token')).toBeNull();
    });
  });

  describe('refresh', () => {
    beforeEach(() => {
      // Set up initial tokens
      localStorage.setItem('refresh_token', 'mock-refresh-token-67890');
    });

    it('should extract tokens from wrapped refresh response', async () => {
      const response = await authApi.refresh();

      // Verify response structure
      expect(response).toHaveProperty('data');
      expect(response.data).toHaveProperty('access_token');
      expect(response.data).toHaveProperty('refresh_token');
      expect(response.data).toHaveProperty('user');

      // Verify tokens are updated in localStorage
      expect(localStorage.getItem('access_token')).toBe('new-mock-access-token-12345');
      expect(localStorage.getItem('refresh_token')).toBe('new-mock-refresh-token-67890');
    });

    it('should handle refresh with invalid token', async () => {
      localStorage.setItem('refresh_token', 'invalid-token');

      await expect(authApi.refresh()).rejects.toThrow('Invalid refresh token');
    });

    it('should handle missing refresh token', async () => {
      localStorage.removeItem('refresh_token');

      await expect(authApi.refresh()).rejects.toThrow('No refresh token available');
    });

    it('should handle fallback to root-level tokens', async () => {
      // Mock response without data wrapper
      server.use(
        http.post('http://localhost:8080/api/auth/refresh', () => {
          return HttpResponse.json({
            access_token: 'fallback-access-token',
            refresh_token: 'fallback-refresh-token',
            user: { id: '1', email: 'test@example.com' }
          }, { status: 200 });
        })
      );

      const response = await authApi.refresh();

      // Verify fallback token extraction works
      expect(localStorage.getItem('access_token')).toBe('fallback-access-token');
      expect(localStorage.getItem('refresh_token')).toBe('fallback-refresh-token');
    });
  });

  describe('getProfile', () => {
    beforeEach(() => {
      localStorage.setItem('access_token', 'mock-access-token-12345');
    });

    it('should extract user data from wrapped response', async () => {
      const user = await authApi.getProfile();

      // Verify user data structure
      expect(user).toHaveProperty('id');
      expect(user).toHaveProperty('email');
      expect(user).toHaveProperty('name');
      expect(user.email).toBe('test@example.com');
    });

    it('should handle fallback to root-level user data', async () => {
      // Mock response without data wrapper
      server.use(
        http.get('http://localhost:8080/api/auth/profile', () => {
          return HttpResponse.json({
            id: '1',
            email: 'fallback@example.com',
            name: 'Fallback User'
          }, { status: 200 });
        })
      );

      const user = await authApi.getProfile();

      // Verify fallback works
      expect(user.email).toBe('fallback@example.com');
      expect(user.name).toBe('Fallback User');
    });
  });

  describe('updateProfile', () => {
    beforeEach(() => {
      localStorage.setItem('access_token', 'mock-access-token-12345');
    });

    it('should extract updated user data from wrapped response', async () => {
      const updateData = { name: 'Updated Name' };

      const user = await authApi.updateProfile(updateData);

      // Verify updated user data
      expect(user).toHaveProperty('id');
      expect(user).toHaveProperty('email');
      expect(user).toHaveProperty('name');
      expect(user.name).toBe('Updated Name');
    });
  });

  describe('createApiKey', () => {
    beforeEach(() => {
      localStorage.setItem('access_token', 'mock-access-token-12345');
    });

    it('should extract API key data from wrapped response', async () => {
      const createData = { name: 'Test API Key', role: 'user' };

      const response = await authApi.createApiKey(createData);

      // Verify API key response structure
      expect(response).toHaveProperty('api_key');
      expect(response).toHaveProperty('key');
      expect(response.api_key).toHaveProperty('name');
      expect(response.api_key.name).toBe('Test API Key');
      expect(response.key).toBe('mcp_1234567890abcdef1234567890abcdef12345678');
    });

    it('should handle fallback to root-level API key data', async () => {
      // Mock response without data wrapper
      server.use(
        http.post('http://localhost:8080/api/auth/api-keys', () => {
          return HttpResponse.json({
            api_key: {
              id: '1',
              name: 'Fallback API Key'
            },
            key: 'fallback-key-12345'
          }, { status: 201 });
        })
      );

      const createData = { name: 'Test API Key', role: 'user' };
      const response = await authApi.createApiKey(createData);

      // Verify fallback works
      expect(response.api_key.name).toBe('Fallback API Key');
      expect(response.key).toBe('fallback-key-12345');
    });
  });

  describe('error handling', () => {
    it('should parse JSON error responses correctly', async () => {
      // Mock a detailed JSON error response
      server.use(
        http.post('http://localhost:8080/api/auth/login', () => {
          return HttpResponse.json({
            error: {
              message: 'Detailed error message',
              code: 'INVALID_CREDENTIALS'
            }
          }, { status: 401 });
        })
      );

      const credentials = { email: 'test@example.com', password: 'wrongpassword' };

      await expect(authApi.login(credentials)).rejects.toThrow('Detailed error message');
    });

    it('should handle non-JSON error responses', async () => {
      // Mock a plain text error response
      server.use(
        http.post('http://localhost:8080/api/auth/login', () => {
          return new HttpResponse('Internal Server Error', { status: 500 });
        })
      );

      const credentials = { email: 'test@example.com', password: 'testpassword123' };

      await expect(authApi.login(credentials)).rejects.toThrow('Server error occurred');
    });

    it('should handle network errors gracefully', async () => {
      // Mock a network error
      server.use(
        http.post('http://localhost:8080/api/auth/login', () => {
          return HttpResponse.error();
        })
      );

      const credentials = { email: 'test@example.com', password: 'testpassword123' };

      await expect(authApi.login(credentials)).rejects.toThrow();
    });
  });

  describe('authentication state', () => {
    it('should correctly report authentication status', () => {
      expect(authApi.isAuthenticated()).toBe(false);

      localStorage.setItem('access_token', 'some-token');
      expect(authApi.isAuthenticated()).toBe(true);

      authApi.clearTokens();
      expect(authApi.isAuthenticated()).toBe(false);
    });

    it('should clear both tokens when clearing', () => {
      localStorage.setItem('access_token', 'access-token');
      localStorage.setItem('refresh_token', 'refresh-token');

      authApi.clearTokens();

      expect(localStorage.getItem('access_token')).toBeNull();
      expect(localStorage.getItem('refresh_token')).toBeNull();
    });
  });

  describe('204 No Content handling', () => {
    it('should handle 204 responses correctly', async () => {
      localStorage.setItem('access_token', 'mock-access-token-12345');

      // Mock a 204 response
      server.use(
        http.post('http://localhost:8080/api/auth/logout', () => {
          return new HttpResponse(null, { status: 204 });
        })
      );

      // Should not throw
      await expect(authApi.logout()).resolves.toBeUndefined();
    });
  });
});
