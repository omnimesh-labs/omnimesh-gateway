'use client';

import type {
	SystemStats,
	MCPServer,
	LogEntry,
	AuditLogEntry,
	LogQueryParams,
	AuditQueryParams,
	Namespace,
	A2AAgent,
	A2AAgentSpec,
	A2AStats,
	Policy,
	CreatePolicyRequest,
	UpdatePolicyRequest,
	ContentFilter,
	CreateContentFilterRequest,
	UpdateContentFilterRequest,
	Endpoint,
	CreateServerRequest,
	MCPDiscoveryResponse,
	MCPPackage,
	Tool,
	CreateToolRequest,
	UpdateToolRequest,
	Prompt,
	CreatePromptRequest,
	UpdatePromptRequest,
	AuthConfiguration,
	AuthConfigurationRequest,
	AuthConfigDefaults
} from '@/lib/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';


async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
	const accessToken = typeof window !== 'undefined' ? localStorage.getItem('access_token') : null;

	const response = await fetch(`${API_BASE_URL}${endpoint}`, {
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
		} catch (_parseError) {
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

// Admin API
class AdminAPI {
	public async getStats(): Promise<SystemStats> {
		return await apiRequest<SystemStats>('/api/admin/stats');
	}

	public async getLogs(params: LogQueryParams = {}): Promise<{ logs: LogEntry[]; total: number }> {
		const query = new URLSearchParams();
		if (params.level) query.set('level', params.level);
		if (params.server_id) query.set('server_id', params.server_id);
		if (params.user_id) query.set('user_id', params.user_id);
		if (params.start_date) query.set('start_date', params.start_date);
		if (params.end_date) query.set('end_date', params.end_date);
		if (params.limit) query.set('limit', String(params.limit));
		if (params.offset) query.set('offset', String(params.offset));

		const queryString = query.toString();
		const endpoint = `/api/admin/logs${queryString ? '?' + queryString : ''}`;
		return await apiRequest<{ logs: LogEntry[]; total: number }>(endpoint);
	}

	public async getAuditLogs(params: AuditQueryParams = {}): Promise<{ logs: AuditLogEntry[]; total: number }> {
		const query = new URLSearchParams();
		if (params.user_id) query.set('user_id', params.user_id);
		if (params.action) query.set('action', params.action);
		if (params.resource_type) query.set('resource_type', params.resource_type);
		if (params.start_date) query.set('start_date', params.start_date);
		if (params.end_date) query.set('end_date', params.end_date);
		if (params.limit) query.set('limit', String(params.limit));
		if (params.offset) query.set('offset', String(params.offset));

		const queryString = query.toString();
		const endpoint = `/api/admin/audit${queryString ? '?' + queryString : ''}`;
		return await apiRequest<{ logs: AuditLogEntry[]; total: number }>(endpoint);
	}
}

// Server API
class ServerAPI {
	public async listServers(): Promise<MCPServer[]> {
		const response = await apiRequest<{data: MCPServer[]; success: boolean}>('/api/gateway/servers');
		return response.data || [];
	}

	public async registerServer(data: CreateServerRequest): Promise<MCPServer> {
		const response = await apiRequest<{data: MCPServer; success: boolean}>('/api/gateway/servers', {
			method: 'POST',
			body: JSON.stringify(data)
		});
		return response.data;
	}

	public async createServer(data: CreateServerRequest): Promise<MCPServer> {
		return this.registerServer(data);
	}

	public async updateServer(id: string, data: Partial<MCPServer>): Promise<MCPServer> {
		const response = await apiRequest<{data: MCPServer; success: boolean}>(`/api/gateway/servers/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
		return response.data;
	}

	public async unregisterServer(id: string): Promise<void> {
		await apiRequest<void>(`/api/gateway/servers/${id}`, {
			method: 'DELETE'
		});
	}

	public async deleteServer(id: string): Promise<void> {
		return this.unregisterServer(id);
	}

	public async getServer(id: string): Promise<MCPServer> {
		const response = await apiRequest<{data: MCPServer; success: boolean}>(`/api/gateway/servers/${id}`);
		return response.data;
	}

	public async testConnection(id: string): Promise<{ success: boolean; error?: string }> {
		return await apiRequest<{ success: boolean; error?: string }>(`/api/gateway/servers/${id}/test`);
	}

	public async restartServer(id: string): Promise<void> {
		await apiRequest<void>(`/api/gateway/servers/${id}/restart`, {
			method: 'POST'
		});
	}

	public async discoverServerTools(id: string): Promise<{ success: boolean; message: string }> {
		return await apiRequest<{ success: boolean; message: string }>(`/api/gateway/servers/${id}/discover-tools`, {
			method: 'POST'
		});
	}
}

// Namespace API
class NamespaceAPI {
	public async listNamespaces(): Promise<Namespace[]> {
		const response = await apiRequest<{ namespaces: Namespace[]; total: number }>('/api/namespaces');
		return response.namespaces || [];
	}

	public async getNamespace(id: string): Promise<Namespace> {
		return await apiRequest<Namespace>(`/api/namespaces/${id}`);
	}

	public async createNamespace(data: Partial<Namespace>): Promise<Namespace> {
		return await apiRequest<Namespace>('/api/namespaces', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	public async updateNamespace(id: string, data: Partial<Namespace>): Promise<Namespace> {
		return await apiRequest<Namespace>(`/api/namespaces/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	public async deleteNamespace(id: string): Promise<void> {
		await apiRequest<void>(`/api/namespaces/${id}`, {
			method: 'DELETE'
		});
	}
}

// A2A Agent API
class A2AApi {
	public async listAgents(params: { is_active?: boolean; tags?: string } = {}): Promise<A2AAgent[]> {
		const query = new URLSearchParams();
		if (params.is_active !== undefined) query.set('is_active', String(params.is_active));
		if (params.tags) query.set('tags', params.tags);

		const queryString = query.toString();
		const endpoint = `/api/a2a${queryString ? '?' + queryString : ''}`;

		const response = await apiRequest<{ agents: A2AAgent[]; total: number }>(endpoint);
		return response.agents || [];
	}

	public async getStats(): Promise<A2AStats> {
		return await apiRequest<A2AStats>('/api/a2a/stats');
	}

	public async createAgent(agentSpec: A2AAgentSpec): Promise<A2AAgent> {
		return await apiRequest<A2AAgent>('/api/a2a', {
			method: 'POST',
			body: JSON.stringify(agentSpec)
		});
	}

	public async updateAgent(id: string, agentData: Partial<A2AAgentSpec>): Promise<A2AAgent> {
		return await apiRequest<A2AAgent>(`/api/a2a/${id}`, {
			method: 'PUT',
			body: JSON.stringify(agentData)
		});
	}

	public async deleteAgent(id: string): Promise<void> {
		await apiRequest<void>(`/api/a2a/${id}`, {
			method: 'DELETE'
		});
	}

	public async toggleAgent(id: string, isActive: boolean): Promise<A2AAgent> {
		return await apiRequest<A2AAgent>(`/api/a2a/${id}/toggle`, {
			method: 'POST',
			body: JSON.stringify({ is_active: isActive })
		});
	}

	public async testAgent(id: string, testData: { message: string; context?: Record<string, unknown> }): Promise<Record<string, unknown>> {
		return await apiRequest<Record<string, unknown>>(`/api/a2a/${id}/test`, {
			method: 'POST',
			body: JSON.stringify(testData)
		});
	}
}

// Policy API
class PolicyAPI {
	public async listPolicies(): Promise<Policy[]> {
		const response = await apiRequest<{ policies: Policy[] }>('/api/admin/policies');
		return response.policies;
	}

	public async getPolicy(id: string): Promise<Policy> {
		const response = await apiRequest<{ policy: Policy }>(`/api/admin/policies/${id}`);
		return response.policy;
	}

	public async createPolicy(data: CreatePolicyRequest): Promise<Policy> {
		const response = await apiRequest<{ policy: Policy }>('/api/admin/policies', {
			method: 'POST',
			body: JSON.stringify(data)
		});
		return response.policy;
	}

	public async updatePolicy(id: string, data: UpdatePolicyRequest): Promise<void> {
		await apiRequest<void>(`/api/admin/policies/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	public async deletePolicy(id: string): Promise<void> {
		await apiRequest<void>(`/api/admin/policies/${id}`, {
			method: 'DELETE'
		});
	}

	public async togglePolicy(id: string, isActive: boolean): Promise<Policy> {
		const response = await apiRequest<{ policy: Policy }>(`/api/admin/policies/${id}`, {
			method: 'PUT',
			body: JSON.stringify({ is_active: isActive })
		});
		return response.policy;
	}
}

// Content Filter API
class ContentFilterAPI {
	public async listFilters(): Promise<ContentFilter[]> {
		const response = await apiRequest<{ filters: ContentFilter[] }>('/api/admin/filters');
		return response.filters;
	}

	public async getFilter(id: string): Promise<ContentFilter> {
		const response = await apiRequest<{ filter: ContentFilter }>(`/api/admin/filters/${id}`);
		return response.filter;
	}

	public async createFilter(data: CreateContentFilterRequest): Promise<ContentFilter> {
		const response = await apiRequest<{ filter: ContentFilter }>('/api/admin/filters', {
			method: 'POST',
			body: JSON.stringify(data)
		});
		return response.filter;
	}

	public async updateFilter(id: string, data: UpdateContentFilterRequest): Promise<void> {
		await apiRequest<void>(`/api/admin/filters/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	public async deleteFilter(id: string): Promise<void> {
		await apiRequest<void>(`/api/admin/filters/${id}`, {
			method: 'DELETE'
		});
	}

	public async getFilterTypes(): Promise<Record<string, unknown>> {
		return await apiRequest<Record<string, unknown>>('/api/admin/filters/types');
	}
}

// Endpoint API
class EndpointAPI {
	public async listEndpoints(): Promise<Endpoint[]> {
		const response = await apiRequest<{ endpoints: Endpoint[]; total: number }>('/api/endpoints');
		return response.endpoints || [];
	}

	public async getEndpoint(id: string): Promise<Endpoint> {
		return await apiRequest<Endpoint>(`/api/endpoints/${id}`);
	}

	public async createEndpoint(data: Partial<Endpoint>): Promise<Endpoint> {
		return await apiRequest<Endpoint>('/api/endpoints', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	public async updateEndpoint(id: string, data: Partial<Endpoint>): Promise<Endpoint> {
		return await apiRequest<Endpoint>(`/api/endpoints/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	public async deleteEndpoint(id: string): Promise<void> {
		await apiRequest<void>(`/api/endpoints/${id}`, {
			method: 'DELETE'
		});
	}
}

// Discovery API - External MCP Package Discovery
class DiscoveryAPI {
	public async searchPackages(query?: string, offset?: number, pageSize?: number): Promise<MCPDiscoveryResponse> {
		const params = new URLSearchParams();
		if (query) params.set('query', query);
		if (offset !== undefined) params.set('offset', String(offset));
		if (pageSize !== undefined) params.set('pageSize', String(pageSize));

		const queryString = params.toString();
		const endpoint = `/api/mcp/search${queryString ? '?' + queryString : ''}`;
		const response = await apiRequest<{ success: boolean; data: MCPDiscoveryResponse; message: string }>(endpoint);
		return response.data;
	}

	public async listPackages(offset?: number, pageSize?: number): Promise<MCPDiscoveryResponse> {
		const params = new URLSearchParams();
		if (offset !== undefined) params.set('offset', String(offset));
		if (pageSize !== undefined) params.set('pageSize', String(pageSize));

		const queryString = params.toString();
		const endpoint = `/api/mcp/packages${queryString ? '?' + queryString : ''}`;
		const response = await apiRequest<{ success: boolean; data: MCPDiscoveryResponse; message: string }>(endpoint);
		return response.data;
	}

	public async getPackageDetails(packageName: string): Promise<MCPPackage> {
		const response = await apiRequest<{ success: boolean; data: MCPPackage; message: string }>(`/api/mcp/packages/${packageName}`);
		return response.data;
	}
}

// Tools API
class ToolsAPI {
	public async listTools(params: {
		active?: boolean;
		category?: string;
		search?: string;
		popular?: boolean;
		include_public?: boolean;
		limit?: number;
		offset?: number;
	} = {}): Promise<Tool[]> {
		const query = new URLSearchParams();
		if (params.active !== undefined) query.set('active', String(params.active));
		if (params.category) query.set('category', params.category);
		if (params.search) query.set('search', params.search);
		if (params.popular) query.set('popular', String(params.popular));
		if (params.include_public) query.set('include_public', String(params.include_public));
		if (params.limit) query.set('limit', String(params.limit));
		if (params.offset) query.set('offset', String(params.offset));

		const queryString = query.toString();
		const endpoint = `/api/gateway/tools${queryString ? '?' + queryString : ''}`;
		const response = await apiRequest<{ success: boolean; data: Tool[]; count: number }>(endpoint);
		return response.data || [];
	}

	public async getTool(id: string): Promise<Tool> {
		const response = await apiRequest<{ success: boolean; data: Tool }>(`/api/gateway/tools/${id}`);
		return response.data;
	}

	public async createTool(data: CreateToolRequest): Promise<Tool> {
		const response = await apiRequest<{ success: boolean; data: Tool }>('/api/gateway/tools', {
			method: 'POST',
			body: JSON.stringify(data)
		});
		return response.data;
	}

	public async updateTool(id: string, data: UpdateToolRequest): Promise<Tool> {
		const response = await apiRequest<{ success: boolean; data: Tool }>(`/api/gateway/tools/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
		return response.data;
	}

	public async deleteTool(id: string): Promise<void> {
		await apiRequest<{ success: boolean; message: string }>(`/api/gateway/tools/${id}`, {
			method: 'DELETE'
		});
	}

	public async executeTool(id: string): Promise<Tool> {
		const response = await apiRequest<{ success: boolean; data: Tool; message: string }>(`/api/gateway/tools/${id}/execute`, {
			method: 'POST'
		});
		return response.data;
	}

	public async getPublicTools(params: { limit?: number; offset?: number } = {}): Promise<Tool[]> {
		const query = new URLSearchParams();
		if (params.limit) query.set('limit', String(params.limit));
		if (params.offset) query.set('offset', String(params.offset));

		const queryString = query.toString();
		const endpoint = `/api/mcp/tools/public${queryString ? '?' + queryString : ''}`;
		const response = await apiRequest<{ success: boolean; data: Tool[]; count: number }>(endpoint);
		return response.data || [];
	}
}

// Prompts API
class PromptsAPI {
	public async listPrompts(params: {
		active?: boolean;
		category?: string;
		search?: string;
		popular?: boolean;
		limit?: number;
		offset?: number;
	} = {}): Promise<Prompt[]> {
		const query = new URLSearchParams();
		if (params.active !== undefined) query.set('active', String(params.active));
		if (params.category) query.set('category', params.category);
		if (params.search) query.set('search', params.search);
		if (params.popular) query.set('popular', String(params.popular));
		if (params.limit) query.set('limit', String(params.limit));
		if (params.offset) query.set('offset', String(params.offset));

		const queryString = query.toString();
		const endpoint = `/api/gateway/prompts${queryString ? '?' + queryString : ''}`;
		const response = await apiRequest<{ success: boolean; data: Prompt[]; count: number }>(endpoint);
		return response.data || [];
	}

	public async getPrompt(id: string): Promise<Prompt> {
		const response = await apiRequest<{ success: boolean; data: Prompt }>(`/api/gateway/prompts/${id}`);
		return response.data;
	}

	public async createPrompt(data: CreatePromptRequest): Promise<Prompt> {
		const response = await apiRequest<{ success: boolean; data: Prompt }>('/api/gateway/prompts', {
			method: 'POST',
			body: JSON.stringify(data)
		});
		return response.data;
	}

	public async updatePrompt(id: string, data: UpdatePromptRequest): Promise<Prompt> {
		const response = await apiRequest<{ success: boolean; data: Prompt }>(`/api/gateway/prompts/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
		return response.data;
	}

	public async deletePrompt(id: string): Promise<void> {
		await apiRequest<{ success: boolean; message: string }>(`/api/gateway/prompts/${id}`, {
			method: 'DELETE'
		});
	}

	public async usePrompt(id: string): Promise<Prompt> {
		const response = await apiRequest<{ success: boolean; data: Prompt; message: string }>(`/api/gateway/prompts/${id}/use`, {
			method: 'POST'
		});
		return response.data;
	}
}

// Auth Configuration API
class AuthConfigAPI {
	public async getAuthConfig(): Promise<AuthConfiguration> {
		const response = await apiRequest<{ success: boolean; data: AuthConfiguration }>('/api/admin/auth-config');
		return response.data;
	}

	public async updateAuthConfig(config: AuthConfigurationRequest): Promise<AuthConfiguration> {
		const response = await apiRequest<{ success: boolean; data: AuthConfiguration; message: string }>('/api/admin/auth-config', {
			method: 'PUT',
			body: JSON.stringify(config)
		});
		return response.data;
	}

	public async getAuthConfigDefaults(): Promise<AuthConfigDefaults> {
		const response = await apiRequest<{ success: boolean; data: AuthConfigDefaults }>('/api/admin/auth-config/defaults');
		return response.data;
	}
}

export const adminApi = new AdminAPI();
export const serverApi = new ServerAPI();
export const namespaceApi = new NamespaceAPI();
export const a2aApi = new A2AApi();
export const policyApi = new PolicyAPI();
export const contentFilterApi = new ContentFilterAPI();
export const endpointApi = new EndpointAPI();
export const discoveryApi = new DiscoveryAPI();
export const toolsApi = new ToolsAPI();
export const promptsApi = new PromptsAPI();
export const authConfigApi = new AuthConfigAPI();
