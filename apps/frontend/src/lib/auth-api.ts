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
			const error = await response.text();
			throw new Error(`API request failed: ${response.status} ${response.statusText} - ${error}`);
		}

		return response.json();
	}

	public isAuthenticated(): boolean {
		return !!this.getToken(ACCESS_TOKEN_KEY);
	}

	public clearTokens(): void {
		this.removeToken(ACCESS_TOKEN_KEY);
		this.removeToken(REFRESH_TOKEN_KEY);
	}

	public async login(credentials: LoginRequest): Promise<LoginResponse> {
		const response = await this.apiRequest<LoginResponse>('/auth/login', {
			method: 'POST',
			body: JSON.stringify(credentials)
		});

		// Store tokens
		this.setToken(ACCESS_TOKEN_KEY, response.access_token);
		this.setToken(REFRESH_TOKEN_KEY, response.refresh_token);

		return response;
	}

	public async logout(): Promise<void> {
		try {
			await this.apiRequest<void>('/auth/logout', {
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

		const response = await this.apiRequest<RefreshResponse>('/auth/refresh', {
			method: 'POST',
			headers: {
				'Authorization': `Bearer ${refreshToken}`
			}
		});

		// Update tokens
		this.setToken(ACCESS_TOKEN_KEY, response.access_token);
		this.setToken(REFRESH_TOKEN_KEY, response.refresh_token);

		return response;
	}

	public async getProfile(): Promise<User> {
		return await this.apiRequest<User>('/auth/profile');
	}

	public async updateProfile(data: Partial<User>): Promise<User> {
		return await this.apiRequest<User>('/auth/profile', {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	public async getApiKeys(): Promise<ApiKey[]> {
		return await this.apiRequest<ApiKey[]>('/auth/api-keys');
	}

	public async createApiKey(data: CreateApiKeyRequest): Promise<CreateApiKeyResponse> {
		return await this.apiRequest<CreateApiKeyResponse>('/auth/api-keys', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	public async deleteApiKey(id: string): Promise<void> {
		await this.apiRequest<void>(`/auth/api-keys/${id}`, {
			method: 'DELETE'
		});
	}
}

export const authApi = new AuthAPI();
