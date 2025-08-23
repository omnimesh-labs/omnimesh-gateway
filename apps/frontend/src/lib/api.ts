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
    timeout?: number;
    max_retries?: number;
    health_check_url?: string;
    working_dir?: string;
    metadata?: Record<string, string>;
}

// Logging and Audit Types
export interface LogEntry {
    id: string;
    organization_id: string;
    server_id?: string;
    session_id?: string;
    rpc_method?: string;
    level: 'debug' | 'info' | 'warn' | 'error';
    started_at: string;
    duration_ms?: number;
    status_code?: number;
    error_flag: boolean;
    storage_provider: string;
    object_uri: string;
    byte_offset?: number;
    user_id: string;
    remote_ip?: string;
    client_id?: string;
    connection_id?: string;
    created_at: string;
}

export interface AuditLogEntry {
    id: string;
    organization_id: string;
    action: string;
    resource_type: string;
    resource_id?: string;
    actor_id: string;
    actor_ip?: string;
    old_values?: Record<string, any>;
    new_values?: Record<string, any>;
    metadata?: Record<string, any>;
    created_at: string;
}

export interface LogQueryParams {
    start_time?: string;
    end_time?: string;
    level?: string;
    type?: string;
    user_id?: string;
    method?: string;
    path?: string;
    search?: string;
    limit?: number;
    offset?: number;
}

export interface AuditQueryParams {
    resource_type?: string;
    action?: string;
    actor_id?: string;
    limit?: number;
    offset?: number;
}

export interface SystemStats {
    users: {
        total: number;
        active: number;
    };
    servers: {
        total: number;
        healthy: number;
    };
    requests: {
        total: number;
        successful: number;
        failed: number;
    };
    rate_limits: {
        total_limits: number;
        blocked: number;
    };
}

// Authentication Types
export interface User {
    id: string;
    organization_id: string;
    email: string;
    role: 'admin' | 'user' | 'viewer';
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface LoginResponse {
    access_token: string;
    refresh_token: string;
    expires_in: number;
    user: User;
}

export interface RefreshRequest {
    refresh_token: string;
}

export interface ApiKey {
    id: string;
    name: string;
    key_hash: string;
    role: 'admin' | 'user' | 'viewer';
    is_active: boolean;
    expires_at?: string;
    created_at: string;
    last_used_at?: string;
}

export interface CreateApiKeyRequest {
    name: string;
    role: 'admin' | 'user' | 'viewer';
    expires_at?: string;
}

export interface CreateApiKeyResponse {
    api_key: ApiKey;
    key: string; // The actual key value (only returned once)
}

export interface UpdateProfileRequest {
    email?: string;
    current_password?: string;
    new_password?: string;
}

// Policy Management Types
export interface Policy {
    id: string;
    organization_id: string;
    name: string;
    description: string;
    type: 'access' | 'rate_limit' | 'security';
    priority: number;
    conditions: Record<string, any>;
    actions: Record<string, any>;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface CreatePolicyRequest {
    name: string;
    description: string;
    type: 'access' | 'rate_limit' | 'security';
    priority: number;
    conditions: Record<string, any>;
    actions: Record<string, any>;
}

export interface UpdatePolicyRequest {
    name?: string;
    description?: string;
    priority?: number;
    conditions?: Record<string, any>;
    actions?: Record<string, any>;
    is_active?: boolean;
}

export interface PolicyListResponse {
    policies: Policy[];
    total: number;
    limit: number;
    offset: number;
}

const API_BASE_URL = 'http://localhost:8080/api';

// Get access token from localStorage
function getAccessToken(): string | null {
    if (typeof window !== 'undefined') {
        return localStorage.getItem('access_token');
    }
    return null;
}

// Set access token in localStorage
function setAccessToken(token: string): void {
    if (typeof window !== 'undefined') {
        localStorage.setItem('access_token', token);
    }
}

// Remove tokens from localStorage
function clearTokens(): void {
    if (typeof window !== 'undefined') {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
    }
}

// Track if we're currently refreshing to avoid infinite loops
let isRefreshing = false;

// Generic fetch wrapper with error handling and automatic token refresh
async function apiRequest<T>(
    endpoint: string,
    options: RequestInit = {},
    retryOnAuth = true
): Promise<T> {
    try {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
            ...options.headers as Record<string, string>,
        };

        // Add Authorization header if token exists
        const token = getAccessToken();
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            headers,
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

            // If 401 and we haven't tried refreshing yet, attempt token refresh
            if (response.status === 401 && retryOnAuth && !isRefreshing && endpoint !== '/auth/refresh' && endpoint !== '/auth/login') {
                const refreshToken = typeof window !== 'undefined' ? localStorage.getItem('refresh_token') : null;
                if (refreshToken) {
                    try {
                        isRefreshing = true;
                        console.log('401 detected, attempting token refresh...');
                        
                        // Try to refresh the token
                        const refreshResponse = await fetch(`${API_BASE_URL}/auth/refresh`, {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            mode: 'cors',
                            credentials: 'include',
                            body: JSON.stringify({ refresh_token: refreshToken }),
                        });

                        if (refreshResponse.ok) {
                            const refreshData = await refreshResponse.json();
                            
                            // Update tokens in localStorage
                            if (typeof window !== 'undefined') {
                                localStorage.setItem('access_token', refreshData.data.access_token);
                                localStorage.setItem('refresh_token', refreshData.data.refresh_token);
                            }

                            console.log('Token refreshed successfully, retrying original request...');
                            
                            // Retry the original request with the new token
                            return apiRequest(endpoint, options, false);
                        } else {
                            // Refresh failed, clear tokens
                            clearTokens();
                            throw new Error('Session expired. Please log in again.');
                        }
                    } catch (refreshError) {
                        clearTokens();
                        throw new Error('Session expired. Please log in again.');
                    } finally {
                        isRefreshing = false;
                    }
                }
            }

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

// Admin & Logging APIs
export const adminApi = {
    // Get system logs with filtering
    async getLogs(params?: LogQueryParams): Promise<LogEntry[]> {
        const queryParams = new URLSearchParams();
        if (params) {
            Object.entries(params).forEach(([key, value]) => {
                if (value !== undefined) {
                    queryParams.append(key, value.toString());
                }
            });
        }

        const response = await apiRequest<ApiResponse<LogEntry[]>>(
            `/admin/logs${queryParams.toString() ? '?' + queryParams.toString() : ''}`
        );
        return response.data;
    },

    // Get audit trail logs
    async getAuditLogs(params?: AuditQueryParams): Promise<{ data: AuditLogEntry[], pagination: { limit: number, offset: number, total: number } }> {
        const queryParams = new URLSearchParams();
        if (params) {
            Object.entries(params).forEach(([key, value]) => {
                if (value !== undefined) {
                    queryParams.append(key, value.toString());
                }
            });
        }

        const response = await apiRequest<ApiResponse<{ data: AuditLogEntry[], pagination: { limit: number, offset: number, total: number } }>>(
            `/admin/audit${queryParams.toString() ? '?' + queryParams.toString() : ''}`
        );
        return response.data;
    },

    // Get system statistics
    async getStats(): Promise<SystemStats> {
        const response = await apiRequest<ApiResponse<SystemStats>>('/admin/stats');
        return response.data;
    },

    // Get Prometheus metrics (raw text)
    async getMetrics(): Promise<string> {
        const response = await fetch(`${API_BASE_URL}/admin/metrics`, {
            headers: {
                'Accept': 'text/plain',
            },
            mode: 'cors',
            credentials: 'include',
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        return response.text();
    },
};

// Authentication APIs
export const authApi = {
    // Login user
    async login(credentials: LoginRequest): Promise<LoginResponse> {
        const response = await apiRequest<ApiResponse<LoginResponse>>('/auth/login', {
            method: 'POST',
            body: JSON.stringify(credentials),
        });

        if (typeof window !== 'undefined') {
            localStorage.setItem('access_token', response.data.access_token);
            localStorage.setItem('refresh_token', response.data.refresh_token);
        }

        return response.data;
    },

    // Refresh access token
    async refresh(): Promise<LoginResponse> {
        const refreshToken = typeof window !== 'undefined' ? localStorage.getItem('refresh_token') : null;
        if (!refreshToken) {
            throw new Error('No refresh token available');
        }

        const response = await apiRequest<ApiResponse<LoginResponse>>('/auth/refresh', {
            method: 'POST',
            body: JSON.stringify({ refresh_token: refreshToken }),
        });

        // Update tokens in localStorage
        if (typeof window !== 'undefined') {
            localStorage.setItem('access_token', response.data.access_token);
            localStorage.setItem('refresh_token', response.data.refresh_token);
        }

        return response.data;
    },

    // Logout user
    async logout(): Promise<void> {
        try {
            await apiRequest<ApiResponse<any>>('/auth/logout', {
                method: 'POST',
            });
        } finally {
            clearTokens();
        }
    },

    // Get user profile
    async getProfile(): Promise<User> {
        const response = await apiRequest<ApiResponse<User>>('/auth/profile');
        return response.data;
    },

    // Update user profile
    async updateProfile(updates: UpdateProfileRequest): Promise<User> {
        const response = await apiRequest<ApiResponse<User>>('/auth/profile', {
            method: 'PUT',
            body: JSON.stringify(updates),
        });
        return response.data;
    },

    // Create API key
    async createApiKey(keyData: CreateApiKeyRequest): Promise<CreateApiKeyResponse> {
        const response = await apiRequest<ApiResponse<CreateApiKeyResponse>>('/auth/api-keys', {
            method: 'POST',
            body: JSON.stringify(keyData),
        });
        return response.data;
    },

    // List API keys
    async listApiKeys(): Promise<ApiKey[]> {
        const response = await apiRequest<ApiResponse<ApiKey[]>>('/auth/api-keys');
        return response.data;
    },

    // Delete API key
    async deleteApiKey(keyId: string): Promise<void> {
        await apiRequest<ApiResponse<any>>(`/auth/api-keys/${keyId}`, {
            method: 'DELETE',
        });
    },

    // Check if user is authenticated (has either access token or refresh token)
    isAuthenticated(): boolean {
        if (typeof window === 'undefined') return false;
        return getAccessToken() !== null || localStorage.getItem('refresh_token') !== null;
    },

    // Clear authentication tokens
    clearTokens,
};

// Policy Management APIs
export const policyApi = {
    // List all policies with optional filtering
    async listPolicies(params?: {
        type?: string;
        is_active?: boolean;
        limit?: number;
        offset?: number;
    }): Promise<PolicyListResponse> {
        const queryParams = new URLSearchParams();
        if (params?.type) queryParams.append('type', params.type);
        if (params?.is_active !== undefined) queryParams.append('is_active', params.is_active.toString());
        if (params?.limit) queryParams.append('limit', params.limit.toString());
        if (params?.offset) queryParams.append('offset', params.offset.toString());

        const response = await apiRequest<PolicyListResponse>(
            `/admin/policies${queryParams.toString() ? '?' + queryParams.toString() : ''}`
        );
        return response;
    },

    // Get specific policy details
    async getPolicy(id: string): Promise<{ policy: Policy }> {
        const response = await apiRequest<{ policy: Policy }>(`/admin/policies/${id}`);
        return response;
    },

    // Create a new policy
    async createPolicy(policyData: CreatePolicyRequest): Promise<{ message: string; policy: Policy; user_id: string }> {
        const response = await apiRequest<{ message: string; policy: Policy; user_id: string }>('/admin/policies', {
            method: 'POST',
            body: JSON.stringify(policyData),
        });
        return response;
    },

    // Update an existing policy
    async updatePolicy(id: string, policyData: UpdatePolicyRequest): Promise<{ message: string; policy_id: string }> {
        const response = await apiRequest<{ message: string; policy_id: string }>(`/admin/policies/${id}`, {
            method: 'PUT',
            body: JSON.stringify(policyData),
        });
        return response;
    },

    // Delete a policy
    async deletePolicy(id: string): Promise<{ message: string; policy_id: string }> {
        const response = await apiRequest<{ message: string; policy_id: string }>(`/admin/policies/${id}`, {
            method: 'DELETE',
        });
        return response;
    },
};

// Content Management Types

// Resource Types
export interface Resource {
    id: string;
    organization_id: string;
    name: string;
    description?: string;
    resource_type: 'file' | 'url' | 'database' | 'api' | 'memory' | 'custom';
    uri: string;
    mime_type?: string;
    size_bytes?: number;
    access_permissions?: Record<string, any>;
    is_active: boolean;
    metadata?: Record<string, any>;
    tags?: string[];
    created_at: string;
    updated_at: string;
    created_by?: string;
}

export interface CreateResourceRequest {
    name: string;
    description?: string;
    resource_type: string;
    uri: string;
    mime_type?: string;
    size_bytes?: number;
    access_permissions?: Record<string, any>;
    metadata?: Record<string, any>;
    tags?: string[];
}

export interface UpdateResourceRequest {
    name?: string;
    description?: string;
    resource_type?: string;
    uri?: string;
    mime_type?: string;
    size_bytes?: number;
    access_permissions?: Record<string, any>;
    is_active?: boolean;
    metadata?: Record<string, any>;
    tags?: string[];
}

// Prompt Types
export interface Prompt {
    id: string;
    organization_id: string;
    name: string;
    description?: string;
    prompt_template: string;
    parameters?: any[];
    category: 'general' | 'coding' | 'analysis' | 'creative' | 'educational' | 'business' | 'custom';
    usage_count: number;
    is_active: boolean;
    metadata?: Record<string, any>;
    tags?: string[];
    created_at: string;
    updated_at: string;
    created_by?: string;
}

export interface CreatePromptRequest {
    name: string;
    description?: string;
    prompt_template: string;
    parameters?: any[];
    category: string;
    metadata?: Record<string, any>;
    tags?: string[];
}

export interface UpdatePromptRequest {
    name?: string;
    description?: string;
    prompt_template?: string;
    parameters?: any[];
    category?: string;
    is_active?: boolean;
    metadata?: Record<string, any>;
    tags?: string[];
}

export interface UsePromptRequest {
    parameters?: Record<string, any>;
}

// Tool Types  
export interface Tool {
    id: string;
    organization_id: string;
    name: string;
    description?: string;
    function_name: string;
    schema?: Record<string, any>;
    category: 'general' | 'data' | 'file' | 'web' | 'system' | 'ai' | 'dev' | 'custom';
    implementation_type: 'internal' | 'external' | 'webhook' | 'script';
    endpoint_url?: string;
    timeout_seconds: number;
    max_retries: number;
    usage_count: number;
    access_permissions?: Record<string, any>;
    is_active: boolean;
    is_public: boolean;
    metadata?: Record<string, any>;
    tags?: string[];
    examples?: any[];
    documentation?: string;
    created_at: string;
    updated_at: string;
    created_by?: string;
}

export interface CreateToolRequest {
    name: string;
    description?: string;
    function_name: string;
    schema: Record<string, any>;
    category: string;
    implementation_type?: string;
    endpoint_url?: string;
    timeout_seconds?: number;
    max_retries?: number;
    access_permissions?: Record<string, any>;
    is_public?: boolean;
    metadata?: Record<string, any>;
    tags?: string[];
    examples?: any[];
    documentation?: string;
}

export interface UpdateToolRequest {
    name?: string;
    description?: string;
    function_name?: string;
    schema?: Record<string, any>;
    category?: string;
    implementation_type?: string;
    endpoint_url?: string;
    timeout_seconds?: number;
    max_retries?: number;
    access_permissions?: Record<string, any>;
    is_active?: boolean;
    is_public?: boolean;
    metadata?: Record<string, any>;
    tags?: string[];
    examples?: any[];
    documentation?: string;
}

export interface ExecuteToolRequest {
    parameters?: Record<string, any>;
}

// List Response Types
export interface ContentListResponse<T> {
    data: T[];
    pagination: {
        total: number;
        limit: number;
        offset: number;
        has_more: boolean;
    };
}

export interface ContentQueryParams {
    search?: string;
    limit?: number;
    offset?: number;
    sort?: string;
    order?: 'asc' | 'desc';
    type?: string;
    category?: string;
    is_active?: boolean;
    is_public?: boolean;
}

// Resource Management APIs
export const resourceApi = {
    // List resources with filtering
    async listResources(params?: ContentQueryParams): Promise<ContentListResponse<Resource>> {
        const queryParams = new URLSearchParams();
        if (params) {
            Object.entries(params).forEach(([key, value]) => {
                if (value !== undefined) {
                    queryParams.append(key, value.toString());
                }
            });
        }

        const response = await apiRequest<ContentListResponse<Resource>>(
            `/gateway/resources${queryParams.toString() ? '?' + queryParams.toString() : ''}`
        );
        return response;
    },

    // Get specific resource
    async getResource(id: string): Promise<Resource> {
        const response = await apiRequest<Resource>(`/gateway/resources/${id}`);
        return response;
    },

    // Create new resource
    async createResource(resourceData: CreateResourceRequest): Promise<Resource> {
        const response = await apiRequest<Resource>('/gateway/resources', {
            method: 'POST',
            body: JSON.stringify(resourceData),
        });
        return response;
    },

    // Update resource
    async updateResource(id: string, resourceData: UpdateResourceRequest): Promise<Resource> {
        const response = await apiRequest<Resource>(`/gateway/resources/${id}`, {
            method: 'PUT',
            body: JSON.stringify(resourceData),
        });
        return response;
    },

    // Delete resource
    async deleteResource(id: string): Promise<void> {
        await apiRequest<ApiResponse<any>>(`/gateway/resources/${id}`, {
            method: 'DELETE',
        });
    },
};

// Prompt Management APIs
export const promptApi = {
    // List prompts with filtering
    async listPrompts(params?: ContentQueryParams): Promise<ContentListResponse<Prompt>> {
        const queryParams = new URLSearchParams();
        if (params) {
            Object.entries(params).forEach(([key, value]) => {
                if (value !== undefined) {
                    queryParams.append(key, value.toString());
                }
            });
        }

        const response = await apiRequest<ContentListResponse<Prompt>>(
            `/gateway/prompts${queryParams.toString() ? '?' + queryParams.toString() : ''}`
        );
        return response;
    },

    // Get specific prompt
    async getPrompt(id: string): Promise<Prompt> {
        const response = await apiRequest<Prompt>(`/gateway/prompts/${id}`);
        return response;
    },

    // Create new prompt
    async createPrompt(promptData: CreatePromptRequest): Promise<Prompt> {
        const response = await apiRequest<Prompt>('/gateway/prompts', {
            method: 'POST',
            body: JSON.stringify(promptData),
        });
        return response;
    },

    // Update prompt
    async updatePrompt(id: string, promptData: UpdatePromptRequest): Promise<Prompt> {
        const response = await apiRequest<Prompt>(`/gateway/prompts/${id}`, {
            method: 'PUT',
            body: JSON.stringify(promptData),
        });
        return response;
    },

    // Delete prompt
    async deletePrompt(id: string): Promise<void> {
        await apiRequest<ApiResponse<any>>(`/gateway/prompts/${id}`, {
            method: 'DELETE',
        });
    },

    // Use prompt (increment usage count)
    async usePrompt(id: string, useData: UsePromptRequest): Promise<{ rendered_prompt: string; usage_count: number }> {
        const response = await apiRequest<{ rendered_prompt: string; usage_count: number }>(`/gateway/prompts/${id}/use`, {
            method: 'POST',
            body: JSON.stringify(useData),
        });
        return response;
    },
};

// Tool Management APIs
export const toolApi = {
    // List tools with filtering
    async listTools(params?: ContentQueryParams): Promise<ContentListResponse<Tool>> {
        const queryParams = new URLSearchParams();
        if (params) {
            Object.entries(params).forEach(([key, value]) => {
                if (value !== undefined) {
                    queryParams.append(key, value.toString());
                }
            });
        }

        const response = await apiRequest<ContentListResponse<Tool>>(
            `/gateway/tools${queryParams.toString() ? '?' + queryParams.toString() : ''}`
        );
        return response;
    },

    // List public tools
    async listPublicTools(params?: ContentQueryParams): Promise<ContentListResponse<Tool>> {
        const queryParams = new URLSearchParams();
        if (params) {
            Object.entries(params).forEach(([key, value]) => {
                if (value !== undefined) {
                    queryParams.append(key, value.toString());
                }
            });
        }

        const response = await apiRequest<ContentListResponse<Tool>>(
            `/mcp/tools/public${queryParams.toString() ? '?' + queryParams.toString() : ''}`
        );
        return response;
    },

    // Get specific tool
    async getTool(id: string): Promise<Tool> {
        const response = await apiRequest<Tool>(`/gateway/tools/${id}`);
        return response;
    },

    // Get tool by function name
    async getToolByFunction(functionName: string): Promise<Tool> {
        const response = await apiRequest<Tool>(`/gateway/tools/function/${functionName}`);
        return response;
    },

    // Create new tool
    async createTool(toolData: CreateToolRequest): Promise<Tool> {
        const response = await apiRequest<Tool>('/gateway/tools', {
            method: 'POST',
            body: JSON.stringify(toolData),
        });
        return response;
    },

    // Update tool
    async updateTool(id: string, toolData: UpdateToolRequest): Promise<Tool> {
        const response = await apiRequest<Tool>(`/gateway/tools/${id}`, {
            method: 'PUT',
            body: JSON.stringify(toolData),
        });
        return response;
    },

    // Delete tool
    async deleteTool(id: string): Promise<void> {
        await apiRequest<ApiResponse<any>>(`/gateway/tools/${id}`, {
            method: 'DELETE',
        });
    },

    // Execute tool
    async executeTool(id: string, executeData: ExecuteToolRequest): Promise<{ result: any; usage_count: number }> {
        const response = await apiRequest<{ result: any; usage_count: number }>(`/gateway/tools/${id}/execute`, {
            method: 'POST',
            body: JSON.stringify(executeData),
        });
        return response;
    },
};
