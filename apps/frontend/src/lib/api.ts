// API service functions for MCP Gateway

// Types for API responses
export interface MCPServer {
    id: string;
    organization_id: string;
    name: string;
    description: string;
    url: string;
    protocol: string;
    version: string;
    status: 'active' | 'inactive' | 'unhealthy' | 'maintenance';
    weight: number;
    metadata: Record<string, string>;
    health_check_url: string;
    timeout: number;
    max_retries: number;
    is_active: boolean;
    created_at: string;
    updated_at: string;
    command?: string;
    args?: string[];
    environment?: string[];
    working_dir?: string;
}

export interface MCPPackage {
    name: string;
    description: string;
    githubUrl: string;
    github_stars: number;
    package_registry: string;
    package_name: string;
    package_download_count: number;
    command: string;
    args: string[];
    envs: string[];
}

export interface MCPDiscoveryResponse {
    results: Record<string, MCPPackage>;
    total: number;
    offset: number;
    pageSize: number;
    hasMore: boolean;
}

export interface ApiResponse<T> {
    success: boolean;
    data: T;
    message?: string;
}

export interface ErrorResponse {
    success: false;
    error: {
        code: string;
        message: string;
        details?: any;
    };
}

export interface CreateServerRequest {
    name: string;
    description: string;
    protocol: string;
    command?: string;
    args?: string[];
    environment?: string[];
    url?: string;
    version?: string;
    weight?: number;
    timeout?: number;
    max_retries?: number;
    health_check_url?: string;
    working_dir?: string;
    metadata?: Record<string, string>;
}

const API_BASE_URL = 'http://localhost:8080/api';

// Generic fetch wrapper with error handling
async function apiRequest<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<T> {
    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            mode: 'cors', // Explicitly set CORS mode
            credentials: 'include', // Include credentials for CORS
            ...options,
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({
                success: false,
                error: {
                    code: 'HTTP_ERROR',
                    message: `HTTP ${response.status}: ${response.statusText}`,
                },
            }));
            throw new Error(errorData.error?.message || `HTTP ${response.status}: ${response.statusText}`);
        }

        return response.json();
    } catch (error) {
        // Handle CORS and network errors
        if (error instanceof TypeError && error.message.includes('fetch')) {
            throw new Error('Network error: Unable to connect to the server. Please ensure the backend is running and CORS is configured correctly.');
        }
        throw error;
    }
}

// Server Management APIs
export const serverApi = {
    // List all registered servers
    async listServers(): Promise<MCPServer[]> {
        const response = await apiRequest<ApiResponse<MCPServer[]>>('/gateway/servers');
        return response.data;
    },

    // Get specific server details
    async getServer(id: string): Promise<MCPServer> {
        const response = await apiRequest<ApiResponse<MCPServer>>(`/gateway/servers/${id}`);
        return response.data;
    },

    // Register a new server
    async registerServer(serverData: CreateServerRequest): Promise<MCPServer> {
        const response = await apiRequest<ApiResponse<MCPServer>>('/gateway/servers', {
            method: 'POST',
            body: JSON.stringify(serverData),
        });
        return response.data;
    },

    // Update an existing server
    async updateServer(id: string, serverData: Partial<CreateServerRequest>): Promise<MCPServer> {
        const response = await apiRequest<ApiResponse<MCPServer>>(`/gateway/servers/${id}`, {
            method: 'PUT',
            body: JSON.stringify(serverData),
        });
        return response.data;
    },

    // Unregister a server
    async unregisterServer(id: string): Promise<void> {
        await apiRequest<ApiResponse<any>>(`/gateway/servers/${id}`, {
            method: 'DELETE',
        });
    },

    // Get server statistics
    async getServerStats(id: string): Promise<any> {
        const response = await apiRequest<ApiResponse<any>>(`/gateway/servers/${id}/stats`);
        return response.data;
    },
};

// MCP Discovery APIs
export const discoveryApi = {
    // Search for available MCP packages
    async searchPackages(query = '', offset = 0, pageSize = 20): Promise<MCPDiscoveryResponse> {
        const params = new URLSearchParams();
        if (query) params.append('query', query);
        if (offset > 0) params.append('offset', offset.toString());
        if (pageSize !== 20) params.append('pageSize', pageSize.toString());

        const response = await apiRequest<ApiResponse<MCPDiscoveryResponse>>(
            `/mcp/search${params.toString() ? '?' + params.toString() : ''}`
        );
        return response.data;
    },

    // List all available packages
    async listPackages(offset = 0, pageSize = 20): Promise<MCPDiscoveryResponse> {
        const params = new URLSearchParams();
        if (offset > 0) params.append('offset', offset.toString());
        if (pageSize !== 20) params.append('pageSize', pageSize.toString());

        const response = await apiRequest<ApiResponse<MCPDiscoveryResponse>>(
            `/mcp/packages${params.toString() ? '?' + params.toString() : ''}`
        );
        return response.data;
    },

    // Get specific package details
    async getPackageDetails(packageName: string): Promise<MCPPackage> {
        const response = await apiRequest<ApiResponse<MCPPackage>>(`/mcp/packages/${packageName}`);
        return response.data;
    },
};

// Session Management APIs
export const sessionApi = {
    // Create a new MCP session
    async createSession(serverId: string): Promise<any> {
        const response = await apiRequest<ApiResponse<any>>('/gateway/sessions', {
            method: 'POST',
            body: JSON.stringify({ server_id: serverId }),
        });
        return response.data;
    },

    // List all sessions
    async listSessions(): Promise<any[]> {
        const response = await apiRequest<ApiResponse<any[]>>('/gateway/sessions');
        return response.data;
    },

    // Close a session
    async closeSession(sessionId: string): Promise<void> {
        await apiRequest<ApiResponse<any>>(`/gateway/sessions/${sessionId}`, {
            method: 'DELETE',
        });
    },
};
