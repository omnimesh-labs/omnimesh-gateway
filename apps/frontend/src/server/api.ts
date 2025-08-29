import 'server-only';
import type {
	MCPServer,
	Endpoint,
	Namespace,
	A2AAgent,
	A2AAgentSpec,
	A2AStats,
	CreateEndpointRequest,
	UpdateEndpointRequest,
	User,
	LoginRequest,
	LoginResponse,
	RefreshResponse,
	ApiKey,
	CreateApiKeyRequest,
	CreateApiKeyResponse,
	SystemStats
} from '@/lib/types';

// Server-side API functions that should never run in the browser
// These make direct HTTP calls to the backend API

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
	const url = `${API_BASE_URL}${endpoint}`;
	const response = await fetch(url, {
		headers: {
			'Content-Type': 'application/json',
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
			// If JSON parsing fails, provide clean error messages based on status code
			if (response.status === 401) {
				throw new Error('Authentication required');
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

		// Fallback to clean error message
		throw new Error(`Request failed: ${response.statusText}`);
	}

	// Handle 204 No Content responses (common for DELETE operations)
	if (response.status === 204) {
		return null as T;
	}

	// Check if response has content before parsing JSON
	const contentType = response.headers.get('content-type');
	if (contentType && contentType.includes('application/json')) {
		return response.json();
	}

	// For non-JSON responses or empty responses, return null
	return null as T;
}

// Endpoint API functions
export async function listEndpoints(): Promise<Endpoint[]> {
	const response = await apiRequest<{ endpoints: Endpoint[]; total: number }>('/api/endpoints');
	return response.endpoints || [];
}

export async function getEndpoint(id: string): Promise<Endpoint> {
	return await apiRequest<Endpoint>(`/api/endpoints/${id}`);
}

export async function createEndpoint(endpointData: CreateEndpointRequest): Promise<Endpoint> {
	return await apiRequest<Endpoint>('/api/endpoints', {
		method: 'POST',
		body: JSON.stringify(endpointData)
	});
}

export async function updateEndpoint(id: string, endpointData: Partial<UpdateEndpointRequest>): Promise<Endpoint> {
	return await apiRequest<Endpoint>(`/api/endpoints/${id}`, {
		method: 'PUT',
		body: JSON.stringify(endpointData)
	});
}

export async function deleteEndpoint(id: string): Promise<void> {
	await apiRequest<void>(`/api/endpoints/${id}`, {
		method: 'DELETE'
	});
}

// Namespace API functions
export async function listNamespaces(): Promise<Namespace[]> {
	const response = await apiRequest<{ namespaces: Namespace[]; total: number }>('/api/namespaces');
	return response.namespaces || [];
}

// A2A Agent API functions
export async function listAgents(params?: { is_active?: boolean; tags?: string }): Promise<A2AAgent[]> {
	const query = new URLSearchParams();
	if (params?.is_active !== undefined) query.set('is_active', String(params.is_active));
	if (params?.tags) query.set('tags', params.tags);

	const queryString = query.toString();
	const endpoint = `/api/a2a${queryString ? '?' + queryString : ''}`;

	const response = await apiRequest<{ agents: A2AAgent[]; total: number }>(endpoint);
	return response.agents || [];
}

export async function getA2AStats(): Promise<A2AStats> {
	return await apiRequest<A2AStats>('/api/a2a/stats');
}

export async function createAgent(agentSpec: A2AAgentSpec): Promise<A2AAgent> {
	return await apiRequest<A2AAgent>('/api/a2a', {
		method: 'POST',
		body: JSON.stringify(agentSpec)
	});
}

export async function updateAgent(id: string, agentData: Partial<A2AAgentSpec>): Promise<A2AAgent> {
	return await apiRequest<A2AAgent>(`/api/a2a/${id}`, {
		method: 'PUT',
		body: JSON.stringify(agentData)
	});
}

export async function deleteAgent(id: string): Promise<void> {
	await apiRequest<void>(`/api/a2a/${id}`, {
		method: 'DELETE'
	});
}

export async function toggleAgent(id: string, isActive: boolean): Promise<A2AAgent> {
	return await apiRequest<A2AAgent>(`/api/a2a/${id}/toggle`, {
		method: 'POST',
		body: JSON.stringify({ is_active: isActive })
	});
}

export async function testAgent(id: string, testData: { message: string; context?: Record<string, unknown> }): Promise<Record<string, unknown>> {
	return await apiRequest<Record<string, unknown>>(`/api/a2a/${id}/test`, {
		method: 'POST',
		body: JSON.stringify(testData)
	});
}

// Auth functions
export async function login(credentials: LoginRequest): Promise<LoginResponse> {
	return await apiRequest<LoginResponse>('/api/auth/login', {
		method: 'POST',
		body: JSON.stringify(credentials)
	});
}

export async function logout(): Promise<void> {
	await apiRequest<void>('/api/auth/logout', {
		method: 'POST'
	});
}

export async function refreshToken(): Promise<RefreshResponse> {
	return await apiRequest<RefreshResponse>('/api/auth/refresh', {
		method: 'POST'
	});
}

export async function getProfile(): Promise<User> {
	return await apiRequest<User>('/api/auth/profile');
}

export async function updateProfile(data: Partial<User>): Promise<User> {
	return await apiRequest<User>('/api/auth/profile', {
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function getApiKeys(): Promise<ApiKey[]> {
	return await apiRequest<ApiKey[]>('/api/auth/api-keys');
}

export async function createApiKey(data: CreateApiKeyRequest): Promise<CreateApiKeyResponse> {
	return await apiRequest<CreateApiKeyResponse>('/api/auth/api-keys', {
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function deleteApiKey(id: string): Promise<void> {
	await apiRequest<void>(`/api/auth/api-keys/${id}`, {
		method: 'DELETE'
	});
}

// Config functions
export async function exportConfiguration(options: Record<string, unknown>): Promise<Blob> {
	const response = await fetch(`${API_BASE_URL}/api/admin/config/export`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(options)
	});

	if (!response.ok) {
		throw new Error(`Export failed: ${response.status} ${response.statusText}`);
	}

	return response.blob();
}

export async function importConfiguration(data: FormData): Promise<void> {
	const response = await fetch(`${API_BASE_URL}/api/admin/config/import`, {
		method: 'POST',
		body: data
	});

	if (!response.ok) {
		throw new Error(`Import failed: ${response.status} ${response.statusText}`);
	}
}

// Admin functions
export async function getStats(): Promise<SystemStats> {
	return await apiRequest<SystemStats>('/api/admin/stats');
}

// Server functions
export async function listServers(): Promise<MCPServer[]> {
	return await apiRequest<MCPServer[]>('/api/gateway/servers');
}
