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
	Endpoint,
	CreateServerRequest,
	MCPDiscoveryResponse
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
		const error = await response.text();
		throw new Error(`API request failed: ${response.status} ${response.statusText} - ${error}`);
	}

	return response.json();
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

	public async createServer(data: CreateServerRequest): Promise<MCPServer> {
		return await apiRequest<MCPServer>('/api/gateway/servers', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	public async updateServer(id: string, data: Partial<MCPServer>): Promise<MCPServer> {
		return await apiRequest<MCPServer>(`/api/gateway/servers/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	public async deleteServer(id: string): Promise<void> {
		await apiRequest<void>(`/api/gateway/servers/${id}`, {
			method: 'DELETE'
		});
	}

	public async getServer(id: string): Promise<MCPServer> {
		return await apiRequest<MCPServer>(`/api/gateway/servers/${id}`);
	}

	public async testConnection(id: string): Promise<{ success: boolean; error?: string }> {
		return await apiRequest<{ success: boolean; error?: string }>(`/api/gateway/servers/${id}/test`);
	}

	public async restartServer(id: string): Promise<void> {
		await apiRequest<void>(`/api/gateway/servers/${id}/restart`, {
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
		return await apiRequest<Policy[]>('/admin/policies');
	}

	public async createPolicy(data: Partial<Policy>): Promise<Policy> {
		return await apiRequest<Policy>('/admin/policies', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	public async updatePolicy(id: string, data: Partial<Policy>): Promise<Policy> {
		return await apiRequest<Policy>(`/admin/policies/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	public async deletePolicy(id: string): Promise<void> {
		await apiRequest<void>(`/admin/policies/${id}`, {
			method: 'DELETE'
		});
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

// Discovery API
class DiscoveryAPI {
	public async discoverServer(id: string): Promise<MCPDiscoveryResponse> {
		return await apiRequest<MCPDiscoveryResponse>(`/api/discovery/servers/${id}`);
	}

	public async discoverByUrl(url: string): Promise<MCPDiscoveryResponse> {
		return await apiRequest<MCPDiscoveryResponse>('/api/discovery/url', {
			method: 'POST',
			body: JSON.stringify({ url })
		});
	}

	public async discoverByCommand(command: string, args?: string[]): Promise<MCPDiscoveryResponse> {
		return await apiRequest<MCPDiscoveryResponse>('/api/discovery/command', {
			method: 'POST',
			body: JSON.stringify({ command, args })
		});
	}
}

export const adminApi = new AdminAPI();
export const serverApi = new ServerAPI();
export const namespaceApi = new NamespaceAPI();
export const a2aApi = new A2AApi();
export const policyApi = new PolicyAPI();
export const endpointApi = new EndpointAPI();
export const discoveryApi = new DiscoveryAPI();
