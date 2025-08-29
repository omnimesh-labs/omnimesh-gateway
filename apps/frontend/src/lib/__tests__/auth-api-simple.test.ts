import { authApi } from '../auth-api';

// Simple test without MSW to test the configuration
describe('AuthAPI - Simple Tests', () => {
  beforeEach(() => {
    localStorage.clear();
    sessionStorage.clear();
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

  describe('token management', () => {
    it('should handle missing tokens gracefully', async () => {
      // Clear all tokens
      localStorage.clear();

      // Should throw error when no refresh token is available
      await expect(authApi.refresh()).rejects.toThrow('No refresh token available');
    });
  });
});
