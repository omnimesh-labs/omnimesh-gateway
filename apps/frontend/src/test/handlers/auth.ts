import { http, HttpResponse } from 'msw';

interface LoginRequest { email: string; password: string; }
interface RefreshRequest { refresh_token: string; }
interface UpdateProfileRequest { name?: string; }
interface CreateApiKeyRequest { name: string; }

const API_BASE_URL = 'http://localhost:8080';

export const authHandlers = [
  // Login endpoint
  http.post(`${API_BASE_URL}/api/auth/login`, async ({ request }) => {
    const { email, password } = await request.json() as LoginRequest;

    if (email === 'test@example.com' && password === 'testpassword123') {
      // Success response with wrapped structure
      return HttpResponse.json({
        success: true,
        data: {
          access_token: 'mock-access-token-12345',
          refresh_token: 'mock-refresh-token-67890',
          token_type: 'Bearer',
          expires_in: 900, // 15 minutes
          user: {
            id: '00000000-0000-0000-0000-000000000001',
            email: 'test@example.com',
            name: 'Test User',
            organization_id: '00000000-0000-0000-0000-000000000002',
            role: 'admin',
            is_active: true,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z'
          }
        }
      }, { status: 200 });
    }

    // Error response
    return HttpResponse.json({
      success: false,
      error: {
        code: 'UNAUTHORIZED',
        message: 'Invalid credentials'
      }
    }, { status: 401 });
  }),

  // Refresh token endpoint
  http.post(`${API_BASE_URL}/api/auth/refresh`, async ({ request }) => {
    const { refresh_token } = await request.json() as RefreshRequest;

    if (refresh_token === 'mock-refresh-token-67890') {
      return HttpResponse.json({
        success: true,
        data: {
          access_token: 'new-mock-access-token-12345',
          refresh_token: 'new-mock-refresh-token-67890',
          token_type: 'Bearer',
          expires_in: 900,
          user: {
            id: '00000000-0000-0000-0000-000000000001',
            email: 'test@example.com',
            name: 'Test User',
            organization_id: '00000000-0000-0000-0000-000000000002',
            role: 'admin',
            is_active: true,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z'
          }
        }
      }, { status: 200 });
    }

    return HttpResponse.json({
      success: false,
      error: {
        code: 'UNAUTHORIZED',
        message: 'Invalid refresh token'
      }
    }, { status: 401 });
  }),

  // Get profile endpoint
  http.get(`${API_BASE_URL}/api/auth/profile`, ({ request }) => {
    const authHeader = request.headers.get('Authorization');

    if (authHeader?.includes('mock-access-token') || authHeader?.includes('new-mock-access-token')) {
      return HttpResponse.json({
        success: true,
        data: {
          id: '00000000-0000-0000-0000-000000000001',
          email: 'test@example.com',
          name: 'Test User',
          organization_id: '00000000-0000-0000-0000-000000000002',
          role: 'admin',
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z'
        }
      }, { status: 200 });
    }

    return HttpResponse.json({
      success: false,
      error: {
        code: 'UNAUTHORIZED',
        message: 'Invalid token'
      }
    }, { status: 401 });
  }),

  // Update profile endpoint
  http.put(`${API_BASE_URL}/api/auth/profile`, async ({ request }) => {
    const authHeader = request.headers.get('Authorization');

    if (authHeader?.includes('mock-access-token') || authHeader?.includes('new-mock-access-token')) {
      const updateData = await request.json() as UpdateProfileRequest;

      return HttpResponse.json({
        success: true,
        data: {
          id: '00000000-0000-0000-0000-000000000001',
          email: 'test@example.com',
          name: updateData.name || 'Test User',
          organization_id: '00000000-0000-0000-0000-000000000002',
          role: 'admin',
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z'
        }
      }, { status: 200 });
    }

    return HttpResponse.json({
      success: false,
      error: {
        code: 'UNAUTHORIZED',
        message: 'Invalid token'
      }
    }, { status: 401 });
  }),

  // Create API key endpoint
  http.post(`${API_BASE_URL}/api/auth/api-keys`, async ({ request }) => {
    const authHeader = request.headers.get('Authorization');

    if (authHeader?.includes('mock-access-token') || authHeader?.includes('new-mock-access-token')) {
      const { name } = await request.json() as CreateApiKeyRequest;

      return HttpResponse.json({
        success: true,
        data: {
          api_key: {
            id: '00000000-0000-0000-0000-000000000003',
            name: name,
            prefix: 'mcp_',
            key_type: 'user',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
            expires_at: null,
            is_active: true
          },
          key: 'mcp_1234567890abcdef1234567890abcdef12345678'
        }
      }, { status: 201 });
    }

    return HttpResponse.json({
      success: false,
      error: {
        code: 'UNAUTHORIZED',
        message: 'Invalid token'
      }
    }, { status: 401 });
  }),

  // Get API keys endpoint
  http.get(`${API_BASE_URL}/api/auth/api-keys`, ({ request }) => {
    const authHeader = request.headers.get('Authorization');

    if (authHeader?.includes('mock-access-token') || authHeader?.includes('new-mock-access-token')) {
      return HttpResponse.json({
        success: true,
        data: [
          {
            id: '00000000-0000-0000-0000-000000000003',
            name: 'Test API Key',
            prefix: 'mcp_',
            key_type: 'user',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
            expires_at: null,
            is_active: true
          }
        ]
      }, { status: 200 });
    }

    return HttpResponse.json({
      success: false,
      error: {
        code: 'UNAUTHORIZED',
        message: 'Invalid token'
      }
    }, { status: 401 });
  }),
];
