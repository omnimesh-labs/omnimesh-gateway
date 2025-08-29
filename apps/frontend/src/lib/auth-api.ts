'use client';

import type {
	User,
	LoginRequest,
	LoginResponse,
	RefreshResponse,
	ApiKey,
	CreateApiKeyRequest,
	CreateApiKeyResponse
} from '@/lib/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

// Token management
const ACCESS_TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';

class AuthAPI {
	private getToken(key: string): string | null {
		if (typeof window === 'undefined') return null;
		return localStorage.getItem(key);
	}

	private setToken(key: string, token: string): void {
		if (typeof window !== 'undefined') {
			localStorage.setItem(key, token);
		}
	}

	private removeToken(key: string): void {
		if (typeof window !== 'undefined') {
			localStorage.removeItem(key);
		}
	}

	private async apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
		const url = `${API_BASE_URL}${endpoint}`;
		const accessToken = this.getToken(ACCESS_TOKEN_KEY);

		const response = await fetch(url, {
			headers: {
				'Content-Type': 'application/json',
				...(accessToken && { 'Authorization': `Bearer ${accessToken}` }),
				...options.headers,
			},
			...options,
		});

		if (!response.ok) {
			const errorText = await response.text();

			// Try to parse JSON error response to get clean error message
			try {
				const errorObj = JSON.parse(errorText);
				if (errorObj.error?.message) {
					throw new Error(errorObj.error.message);
				}
				if (errorObj.message) {
					throw new Error(errorObj.message);
				}
				if (errorObj.error) {
					throw new Error(errorObj.error);
				}
			} catch (parseError) {
				// If JSON parsing fails, use a clean fallback message
				if (response.status === 401) {
					throw new Error('Invalid credentials');
				}
				if (response.status === 403) {
					throw new Error('Access denied');
				}
				if (response.status === 404) {
					throw new Error('Resource not found');
				}
				if (response.status >= 500) {
					throw new Error('Server error occurred');
				}
			}

			// Fallback to original error format if nothing else worked
			throw new Error(`Request failed: ${response.statusText}`);
		}

		// Handle 204 No Content responses (common for DELETE operations)
		if (response.status === 204) {
			return null as unknown as T;
		}

		// Check if response has content before parsing JSON
		const contentType = response.headers.get('content-type');
		if (contentType && contentType.includes('application/json')) {
			return response.json();
		}

		// For non-JSON responses or empty responses, return null
		return null as unknown as T;
	}

	public isAuthenticated(): boolean {
		return !!this.getToken(ACCESS_TOKEN_KEY);
	}

	public clearTokens(): void {
		this.removeToken(ACCESS_TOKEN_KEY);
		this.removeToken(REFRESH_TOKEN_KEY);
	}

	public async login(credentials: LoginRequest): Promise<LoginResponse> {
		const response = await this.apiRequest<LoginResponse>('/api/auth/login', {
			method: 'POST',
			body: JSON.stringify(credentials)
		});

		// Store tokens - tokens are nested under response.data
		const accessToken = response.data?.access_token || response.access_token;
		const refreshToken = response.data?.refresh_token || response.refresh_token;
		this.setToken(ACCESS_TOKEN_KEY, accessToken);
		this.setToken(REFRESH_TOKEN_KEY, refreshToken);

		return response;
	}

	public async logout(): Promise<void> {
		try {
			await this.apiRequest<void>('/api/auth/logout', {
				method: 'POST'
			});
		} finally {
			this.clearTokens();
		}
	}

	public async refresh(): Promise<RefreshResponse> {
		const refreshToken = this.getToken(REFRESH_TOKEN_KEY);
		if (!refreshToken) {
			throw new Error('No refresh token available');
		}

		const response = await this.apiRequest<RefreshResponse>('/api/auth/refresh', {
			method: 'POST',
			headers: {
				'Authorization': `Bearer ${refreshToken}`
			}
		});

		// Update tokens - tokens are nested under response.data
		const newAccessToken = response.data?.access_token || response.access_token;
		const newRefreshToken = response.data?.refresh_token || response.refresh_token;
		this.setToken(ACCESS_TOKEN_KEY, newAccessToken);
		this.setToken(REFRESH_TOKEN_KEY, newRefreshToken);

		return response;
	}

	public async getProfile(): Promise<User> {
		const response = await this.apiRequest<{data: User; success: boolean}>('/api/auth/profile');
		return response.data || response as unknown as User;
	}

	public async updateProfile(data: Partial<User>): Promise<User> {
		const response = await this.apiRequest<{data: User; success: boolean}>('/api/auth/profile', {
			method: 'PUT',
			body: JSON.stringify(data)
		});
		return response.data || response as unknown as User;
	}

	public async getApiKeys(): Promise<ApiKey[]> {
		const response = await this.apiRequest<{data: ApiKey[], success: boolean}>('/api/auth/api-keys');
		return response.data || [];
	}

	public async createApiKey(data: CreateApiKeyRequest): Promise<CreateApiKeyResponse> {
		const response = await this.apiRequest<{data: CreateApiKeyResponse; success: boolean}>('/api/auth/api-keys', {
			method: 'POST',
			body: JSON.stringify(data)
		});
		return response.data || response as unknown as CreateApiKeyResponse;
	}

	public async deleteApiKey(id: string): Promise<void> {
		await this.apiRequest<void>(`/api/auth/api-keys/${id}`, {
			method: 'DELETE'
		});
	}
}

export const authApi = new AuthAPI();
