// API service functions for MCP Gateway - Migrated from old frontend

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
	role: 'admin' | 'user' | 'viewer' | 'api_user';
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
	role: 'admin' | 'user' | 'viewer' | 'api_user';
	is_active: boolean;
	expires_at?: string;
	created_at: string;
	last_used_at?: string;
}

export interface CreateApiKeyRequest {
	name: string;
	role: 'admin' | 'user' | 'viewer' | 'api_user';
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

// Content Management Types - Resource Types
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

// Namespace Management Types
export interface Namespace {
	id: string;
	organization_id: string;
	name: string;
	description?: string;
	servers?: string[];
	server_count?: number;
	created_at: string;
	updated_at: string;
	created_by?: string;
	is_active: boolean;
	metadata?: Record<string, any>;
}

export interface CreateNamespaceRequest {
	name: string;
	description?: string;
	server_ids?: string[];
}

export interface UpdateNamespaceRequest {
	name?: string;
	description?: string;
	server_ids?: string[];
}

export interface NamespaceTool {
	tool_id: string;
	tool_name: string;
	prefixed_name: string;
	server_id: string;
	server_name: string;
	status: 'ACTIVE' | 'INACTIVE';
}

export interface ExecuteNamespaceToolRequest {
	tool_name: string;
	parameters?: Record<string, any>;
}

// Endpoint Management Types
export interface EndpointURLs {
	sse: string;
	http: string;
	websocket: string;
	openapi: string;
	documentation: string;
}

export interface Endpoint {
	id: string;
	organization_id: string;
	namespace_id: string;
	name: string;
	description?: string;
	enable_api_key_auth: boolean;
	enable_oauth: boolean;
	enable_public_access: boolean;
	use_query_param_auth: boolean;
	rate_limit_requests: number;
	rate_limit_window: number;
	allowed_origins: string[];
	allowed_methods: string[];
	created_at: string;
	updated_at: string;
	created_by?: string;
	is_active: boolean;
	metadata?: Record<string, any>;
	urls?: EndpointURLs;
	namespace?: Namespace;
}

export interface CreateEndpointRequest {
	namespace_id: string;
	name: string;
	description?: string;
	enable_api_key_auth?: boolean;
	enable_oauth?: boolean;
	enable_public_access?: boolean;
	use_query_param_auth?: boolean;
	rate_limit_requests?: number;
	rate_limit_window?: number;
	allowed_origins?: string[];
	allowed_methods?: string[];
	metadata?: Record<string, any>;
}

export interface UpdateEndpointRequest {
	description?: string;
	enable_api_key_auth?: boolean;
	enable_oauth?: boolean;
	enable_public_access?: boolean;
	use_query_param_auth?: boolean;
	rate_limit_requests?: number;
	rate_limit_window?: number;
	allowed_origins?: string[];
	allowed_methods?: string[];
	is_active?: boolean;
	metadata?: Record<string, any>;
}

// Configuration Management Types
export interface ConfigurationExport {
	metadata: ExportMetadata;
	servers?: any[];
	virtualServers?: any[];
	tools?: any[];
	prompts?: any[];
	resources?: any[];
	policies?: any[];
	rateLimits?: any[];
}

export interface ExportMetadata {
	exportId: string;
	timestamp: string;
	version: string;
	gateway: string;
	organization: string;
	entityTypes: string[];
	totalEntities: number;
	filters: ExportFilters;
	exportedBy: string;
	tags?: Record<string, string>;
}

export interface ExportFilters {
	includeInactive: boolean;
	includeDependencies: boolean;
	tags?: string[];
	dateRange?: DateRange;
}

export interface DateRange {
	start: string;
	end: string;
}

export interface ExportRequest {
	entityTypes: string[];
	includeInactive: boolean;
	includeDependencies: boolean;
	tags?: string[];
	filters?: ExportFilters;
}

export interface ImportRequest {
	configData: ConfigurationExport;
	conflictStrategy: 'update' | 'skip' | 'rename' | 'fail';
	dryRun: boolean;
	rekeySecret?: string;
	options?: ImportOptions;
}

export interface ImportOptions {
	preserveIds?: boolean;
	updateTimestamps?: boolean;
	validateReferences?: boolean;
	ignoreEntityTypes?: string[];
}

export interface ImportResult {
	importId: string;
	status: 'pending' | 'running' | 'completed' | 'failed' | 'partial' | 'validating';
	summary: ImportSummary;
	details?: ImportItemResult[];
	errors?: ImportError[];
	warnings?: ImportWarning[];
	startedAt: string;
	completedAt?: string;
	duration?: number;
}

export interface ImportSummary {
	totalItems: number;
	processedItems: number;
	createdItems: number;
	updatedItems: number;
	skippedItems: number;
	failedItems: number;
	entityCounts: Record<string, ImportEntityCount>;
}

export interface ImportEntityCount {
	total: number;
	created: number;
	updated: number;
	skipped: number;
	failed: number;
}

export interface ImportItemResult {
	entityType: string;
	entityId?: string;
	entityName: string;
	action: 'create' | 'update' | 'skip' | 'rename';
	status: string;
	message?: string;
	error?: string;
	oldId?: string;
	newId?: string;
}

export interface ImportError {
	code: string;
	message: string;
	entityType?: string;
	entityName?: string;
	details?: Record<string, any>;
}

export interface ImportWarning {
	code: string;
	message: string;
	entityType?: string;
	entityName?: string;
	details?: Record<string, any>;
}

export interface ImportHistory {
	id: string;
	organizationId: string;
	filename: string;
	entityTypes: string[];
	status: 'pending' | 'running' | 'completed' | 'failed' | 'partial' | 'validating';
	conflictStrategy: 'update' | 'skip' | 'rename' | 'fail';
	dryRun: boolean;
	summary: ImportSummary;
	errorCount: number;
	warningCount: number;
	duration?: number;
	importedBy: string;
	metadata: Record<string, any>;
	createdAt: string;
	completedAt?: string;
}

export interface ValidateImportRequest {
	configData: ConfigurationExport;
}

export interface ValidationResult {
	valid: boolean;
	errors: ValidationError[];
	warnings: ValidationWarning[];
	entityCounts: Record<string, number>;
	conflicts: ConflictItem[];
	dependencies: DependencyItem[];
	compatibilityCheck: CompatibilityResult;
}

export interface ValidationError {
	code: string;
	message: string;
	entityType?: string;
	entityName?: string;
	field?: string;
	details?: Record<string, any>;
}

export interface ValidationWarning {
	code: string;
	message: string;
	entityType?: string;
	entityName?: string;
	details?: Record<string, any>;
}

export interface ConflictItem {
	entityType: string;
	entityName: string;
	conflictType: string;
	existingValue: any;
	importValue: any;
	suggestion?: string;
	details?: Record<string, any>;
}

export interface DependencyItem {
	entityType: string;
	entityName: string;
	dependsOnType: string;
	dependsOnName: string;
	dependsOnId: string;
	required: boolean;
	missing: boolean;
}

export interface CompatibilityResult {
	compatible: boolean;
	version: string;
	requiredVersion?: string;
	missingFeatures?: string[];
	unsupportedTypes?: string[];
}

export interface ImportHistoryQuery {
	status?: string;
	entityType?: string;
	importedBy?: string;
	startDate?: string;
	endDate?: string;
	limit?: number;
	offset?: number;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api';

// Get access token from localStorage
function getAccessToken(): string | null {
	if (typeof window !== 'undefined') {
		return localStorage.getItem('access_token');
	}

	return null;
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
async function apiRequest<T>(endpoint: string, options: RequestInit = {}, retryOnAuth = true): Promise<T> {
	try {
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
			...(options.headers as Record<string, string>)
		};

		// Add Authorization header if token exists
		const token = getAccessToken();

		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}

		const response = await fetch(`${API_BASE_URL}${endpoint}`, {
			headers,
			mode: 'cors',
			credentials: 'include',
			...options
		});

		if (!response.ok) {
			const errorData = await response.json().catch(() => ({
				success: false,
				error: {
					code: 'HTTP_ERROR',
					message: `HTTP ${response.status}: ${response.statusText}`
				}
			}));

			// If 401 and we haven't tried refreshing yet, attempt token refresh
			if (
				response.status === 401 &&
				retryOnAuth &&
				!isRefreshing &&
				endpoint !== '/auth/refresh' &&
				endpoint !== '/auth/login'
			) {
				const refreshToken = typeof window !== 'undefined' ? localStorage.getItem('refresh_token') : null;

				if (refreshToken) {
					try {
						isRefreshing = true;
						// Try to refresh the token
						const refreshResponse = await fetch(`${API_BASE_URL}/auth/refresh`, {
							method: 'POST',
							headers: {
								'Content-Type': 'application/json'
							},
							mode: 'cors',
							credentials: 'include',
							body: JSON.stringify({ refresh_token: refreshToken })
						});

						if (refreshResponse.ok) {
							const refreshData = await refreshResponse.json();

							// Update tokens in localStorage synchronously
							if (typeof window !== 'undefined') {
								localStorage.setItem('access_token', refreshData.data.access_token);
								localStorage.setItem('refresh_token', refreshData.data.refresh_token);
							}

							// Retry the original request with the new token
							return apiRequest(endpoint, options, false);
						} else {
							// Refresh failed, clear tokens and redirect to login
							clearTokens();

							if (typeof window !== 'undefined') {
								window.location.href = '/sign-in';
							}

							throw new Error('Session expired. Please log in again.');
						}
					} catch (_refreshError) {
						clearTokens();

						// Redirect to login page when refresh fails
						if (typeof window !== 'undefined') {
							window.location.href = '/sign-in';
						}

						throw new Error('Session expired. Please log in again.');
					} finally {
						isRefreshing = false;
					}
				} else {
					// No refresh token available, redirect to login
					clearTokens();

					if (typeof window !== 'undefined') {
						window.location.href = '/sign-in';
					}

					throw new Error('Session expired. Please log in again.');
				}
			}

			// For non-401 errors or when not retrying auth
			throw new Error(errorData.error?.message || `HTTP ${response.status}: ${response.statusText}`);
		}

		return response.json();
	} catch (error) {
		// Handle CORS and network errors
		if (error instanceof TypeError && error.message.includes('fetch')) {
			throw new Error(
				'Network error: Unable to connect to the server. Please ensure the backend is running and CORS is configured correctly.'
			);
		}

		throw error;
	}
}

// Server Management APIs
export const serverApi = {
	async listServers(): Promise<MCPServer[]> {
		const response = await apiRequest<ApiResponse<MCPServer[]>>('/gateway/servers');
		return response.data;
	},

	async getServer(id: string): Promise<MCPServer> {
		const response = await apiRequest<ApiResponse<MCPServer>>(`/gateway/servers/${id}`);
		return response.data;
	},

	async registerServer(serverData: CreateServerRequest): Promise<MCPServer> {
		const response = await apiRequest<ApiResponse<MCPServer>>('/gateway/servers', {
			method: 'POST',
			body: JSON.stringify(serverData)
		});
		return response.data;
	},

	async updateServer(id: string, serverData: Partial<CreateServerRequest>): Promise<MCPServer> {
		const response = await apiRequest<ApiResponse<MCPServer>>(`/gateway/servers/${id}`, {
			method: 'PUT',
			body: JSON.stringify(serverData)
		});
		return response.data;
	},

	async unregisterServer(id: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/gateway/servers/${id}`, {
			method: 'DELETE'
		});
	},

	async getServerStats(id: string): Promise<any> {
		const response = await apiRequest<ApiResponse<any>>(`/gateway/servers/${id}/stats`);
		return response.data;
	}
};

// MCP Discovery APIs
export const discoveryApi = {
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

	async listPackages(offset = 0, pageSize = 20): Promise<MCPDiscoveryResponse> {
		const params = new URLSearchParams();

		if (offset > 0) params.append('offset', offset.toString());

		if (pageSize !== 20) params.append('pageSize', pageSize.toString());

		const response = await apiRequest<ApiResponse<MCPDiscoveryResponse>>(
			`/mcp/packages${params.toString() ? '?' + params.toString() : ''}`
		);
		return response.data;
	},

	async getPackageDetails(packageName: string): Promise<MCPPackage> {
		const response = await apiRequest<ApiResponse<MCPPackage>>(`/mcp/packages/${packageName}`);
		return response.data;
	}
};

// Session Management APIs
export const sessionApi = {
	async createSession(serverId: string): Promise<any> {
		const response = await apiRequest<ApiResponse<any>>('/gateway/sessions', {
			method: 'POST',
			body: JSON.stringify({ server_id: serverId })
		});
		return response.data;
	},

	async listSessions(): Promise<any[]> {
		const response = await apiRequest<ApiResponse<any[]>>('/gateway/sessions');
		return response.data;
	},

	async closeSession(sessionId: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/gateway/sessions/${sessionId}`, {
			method: 'DELETE'
		});
	}
};

// Admin & Logging APIs
export const adminApi = {
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

	async getAuditLogs(
		params?: AuditQueryParams
	): Promise<{ data: AuditLogEntry[]; pagination: { limit: number; offset: number; total: number } }> {
		const queryParams = new URLSearchParams();

		if (params) {
			Object.entries(params).forEach(([key, value]) => {
				if (value !== undefined) {
					queryParams.append(key, value.toString());
				}
			});
		}

		const response = await apiRequest<
			ApiResponse<{ data: AuditLogEntry[]; pagination: { limit: number; offset: number; total: number } }>
		>(`/admin/audit${queryParams.toString() ? '?' + queryParams.toString() : ''}`);
		return response.data;
	},

	async getStats(): Promise<SystemStats> {
		const response = await apiRequest<ApiResponse<SystemStats>>('/admin/stats');
		return response.data;
	},

	async getMetrics(): Promise<string> {
		const response = await fetch(`${API_BASE_URL}/admin/metrics`, {
			headers: {
				Accept: 'text/plain'
			},
			mode: 'cors',
			credentials: 'include'
		});

		if (!response.ok) {
			throw new Error(`HTTP ${response.status}: ${response.statusText}`);
		}

		return response.text();
	}
};

// Authentication APIs
export const authApi = {
	async login(credentials: LoginRequest): Promise<LoginResponse> {
		const response = await apiRequest<ApiResponse<LoginResponse>>('/auth/login', {
			method: 'POST',
			body: JSON.stringify(credentials)
		});

		// Store tokens synchronously to prevent race conditions
		if (typeof window !== 'undefined') {
			localStorage.setItem('access_token', response.data.access_token);
			localStorage.setItem('refresh_token', response.data.refresh_token);
			// Force a small delay to ensure localStorage write is complete
			await new Promise((resolve) => setTimeout(resolve, 10));
		}

		return response.data;
	},

	async refresh(): Promise<LoginResponse> {
		const refreshToken = typeof window !== 'undefined' ? localStorage.getItem('refresh_token') : null;

		if (!refreshToken) {
			throw new Error('No refresh token available');
		}

		const response = await apiRequest<ApiResponse<LoginResponse>>('/auth/refresh', {
			method: 'POST',
			body: JSON.stringify({ refresh_token: refreshToken })
		});

		// Update tokens in localStorage
		if (typeof window !== 'undefined') {
			localStorage.setItem('access_token', response.data.access_token);
			localStorage.setItem('refresh_token', response.data.refresh_token);
			// Force a small delay to ensure localStorage write is complete
			await new Promise((resolve) => setTimeout(resolve, 10));
		}

		return response.data;
	},

	async logout(): Promise<void> {
		try {
			await apiRequest<ApiResponse<any>>('/auth/logout', {
				method: 'POST'
			});
		} finally {
			clearTokens();
		}
	},

	async getProfile(): Promise<User> {
		const response = await apiRequest<ApiResponse<User>>('/auth/profile');
		return response.data;
	},

	async updateProfile(updates: UpdateProfileRequest): Promise<User> {
		const response = await apiRequest<ApiResponse<User>>('/auth/profile', {
			method: 'PUT',
			body: JSON.stringify(updates)
		});
		return response.data;
	},

	async createApiKey(keyData: CreateApiKeyRequest): Promise<CreateApiKeyResponse> {
		const response = await apiRequest<ApiResponse<CreateApiKeyResponse>>('/auth/api-keys', {
			method: 'POST',
			body: JSON.stringify(keyData)
		});
		return response.data;
	},

	async listApiKeys(): Promise<ApiKey[]> {
		const response = await apiRequest<ApiResponse<ApiKey[]>>('/auth/api-keys');
		return response.data;
	},

	async deleteApiKey(keyId: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/auth/api-keys/${keyId}`, {
			method: 'DELETE'
		});
	},

	isAuthenticated(): boolean {
		if (typeof window === 'undefined') return false;

		return getAccessToken() !== null || localStorage.getItem('refresh_token') !== null;
	},

	clearTokens
};

// Policy Management APIs
export const policyApi = {
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

	async getPolicy(id: string): Promise<{ policy: Policy }> {
		const response = await apiRequest<{ policy: Policy }>(`/admin/policies/${id}`);
		return response;
	},

	async createPolicy(policyData: CreatePolicyRequest): Promise<{ message: string; policy: Policy; user_id: string }> {
		const response = await apiRequest<{ message: string; policy: Policy; user_id: string }>('/admin/policies', {
			method: 'POST',
			body: JSON.stringify(policyData)
		});
		return response;
	},

	async updatePolicy(id: string, policyData: UpdatePolicyRequest): Promise<{ message: string; policy_id: string }> {
		const response = await apiRequest<{ message: string; policy_id: string }>(`/admin/policies/${id}`, {
			method: 'PUT',
			body: JSON.stringify(policyData)
		});
		return response;
	},

	async deletePolicy(id: string): Promise<{ message: string; policy_id: string }> {
		const response = await apiRequest<{ message: string; policy_id: string }>(`/admin/policies/${id}`, {
			method: 'DELETE'
		});
		return response;
	}
};

// Resource Management APIs
export const resourceApi = {
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

	async getResource(id: string): Promise<Resource> {
		const response = await apiRequest<Resource>(`/gateway/resources/${id}`);
		return response;
	},

	async createResource(resourceData: CreateResourceRequest): Promise<Resource> {
		const response = await apiRequest<Resource>('/gateway/resources', {
			method: 'POST',
			body: JSON.stringify(resourceData)
		});
		return response;
	},

	async updateResource(id: string, resourceData: UpdateResourceRequest): Promise<Resource> {
		const response = await apiRequest<Resource>(`/gateway/resources/${id}`, {
			method: 'PUT',
			body: JSON.stringify(resourceData)
		});
		return response;
	},

	async deleteResource(id: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/gateway/resources/${id}`, {
			method: 'DELETE'
		});
	}
};

// Prompt Management APIs
export const promptApi = {
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

	async getPrompt(id: string): Promise<Prompt> {
		const response = await apiRequest<Prompt>(`/gateway/prompts/${id}`);
		return response;
	},

	async createPrompt(promptData: CreatePromptRequest): Promise<Prompt> {
		const response = await apiRequest<Prompt>('/gateway/prompts', {
			method: 'POST',
			body: JSON.stringify(promptData)
		});
		return response;
	},

	async updatePrompt(id: string, promptData: UpdatePromptRequest): Promise<Prompt> {
		const response = await apiRequest<Prompt>(`/gateway/prompts/${id}`, {
			method: 'PUT',
			body: JSON.stringify(promptData)
		});
		return response;
	},

	async deletePrompt(id: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/gateway/prompts/${id}`, {
			method: 'DELETE'
		});
	},

	async usePrompt(id: string, useData: UsePromptRequest): Promise<{ rendered_prompt: string; usage_count: number }> {
		const response = await apiRequest<{ rendered_prompt: string; usage_count: number }>(
			`/gateway/prompts/${id}/use`,
			{
				method: 'POST',
				body: JSON.stringify(useData)
			}
		);
		return response;
	}
};

// Tool Management APIs
export const toolApi = {
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

	async getTool(id: string): Promise<Tool> {
		const response = await apiRequest<Tool>(`/gateway/tools/${id}`);
		return response;
	},

	async getToolByFunction(functionName: string): Promise<Tool> {
		const response = await apiRequest<Tool>(`/gateway/tools/function/${functionName}`);
		return response;
	},

	async createTool(toolData: CreateToolRequest): Promise<Tool> {
		const response = await apiRequest<Tool>('/gateway/tools', {
			method: 'POST',
			body: JSON.stringify(toolData)
		});
		return response;
	},

	async updateTool(id: string, toolData: UpdateToolRequest): Promise<Tool> {
		const response = await apiRequest<Tool>(`/gateway/tools/${id}`, {
			method: 'PUT',
			body: JSON.stringify(toolData)
		});
		return response;
	},

	async deleteTool(id: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/gateway/tools/${id}`, {
			method: 'DELETE'
		});
	},

	async executeTool(id: string, executeData: ExecuteToolRequest): Promise<{ result: any; usage_count: number }> {
		const response = await apiRequest<{ result: any; usage_count: number }>(`/gateway/tools/${id}/execute`, {
			method: 'POST',
			body: JSON.stringify(executeData)
		});
		return response;
	}
};

// Configuration Management APIs
export const configApi = {
	async exportConfiguration(request: ExportRequest): Promise<ConfigurationExport> {
		const response = await apiRequest<ApiResponse<ConfigurationExport>>('/admin/config/export', {
			method: 'POST',
			body: JSON.stringify(request)
		});
		return response.data;
	},

	async importConfiguration(request: ImportRequest): Promise<ImportResult> {
		const response = await apiRequest<ApiResponse<ImportResult>>('/admin/config/import', {
			method: 'POST',
			body: JSON.stringify(request)
		});
		return response.data;
	},

	async validateImport(request: ValidateImportRequest): Promise<ValidationResult> {
		const response = await apiRequest<ApiResponse<ValidationResult>>('/admin/config/validate', {
			method: 'POST',
			body: JSON.stringify(request)
		});
		return response.data;
	},

	async getImportHistory(
		params?: ImportHistoryQuery
	): Promise<{ data: ImportHistory[]; pagination: { limit: number; offset: number; total: number } }> {
		const queryParams = new URLSearchParams();

		if (params) {
			Object.entries(params).forEach(([key, value]) => {
				if (value !== undefined) {
					queryParams.append(key, value.toString());
				}
			});
		}

		const response = await apiRequest<
			ApiResponse<{ data: ImportHistory[]; pagination: { limit: number; offset: number; total: number } }>
		>(`/admin/config/imports${queryParams.toString() ? '?' + queryParams.toString() : ''}`);
		return response.data;
	},

	async parseConfigurationFile(file: File): Promise<ConfigurationExport> {
		return new Promise((resolve, reject) => {
			const reader = new FileReader();
			reader.onload = (e) => {
				try {
					const content = e.target?.result as string;
					const config = JSON.parse(content) as ConfigurationExport;
					resolve(config);
				} catch (error) {
					reject(new Error('Invalid JSON file format'));
				}
			};
			reader.onerror = () => reject(new Error('Failed to read file'));
			reader.readAsText(file);
		});
	},

	downloadConfiguration(config: ConfigurationExport, filename?: string): void {
		const blob = new Blob([JSON.stringify(config, null, 2)], {
			type: 'application/json'
		});
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = filename || `mcp-gateway-config-${Date.now()}.json`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}
};

// Namespace Management APIs
export const namespaceApi = {
	async listNamespaces(): Promise<Namespace[]> {
		const response = await apiRequest<{ namespaces: Namespace[]; total: number }>('/namespaces');
		return response.namespaces;
	},

	async getNamespace(id: string): Promise<Namespace> {
		const response = await apiRequest<Namespace>(`/namespaces/${id}`);
		return response;
	},

	async createNamespace(namespaceData: CreateNamespaceRequest): Promise<Namespace> {
		const response = await apiRequest<Namespace>('/namespaces', {
			method: 'POST',
			body: JSON.stringify(namespaceData)
		});
		return response;
	},

	async updateNamespace(id: string, namespaceData: UpdateNamespaceRequest): Promise<Namespace> {
		const response = await apiRequest<Namespace>(`/namespaces/${id}`, {
			method: 'PUT',
			body: JSON.stringify(namespaceData)
		});
		return response;
	},

	async deleteNamespace(id: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/namespaces/${id}`, {
			method: 'DELETE'
		});
	},

	async addServerToNamespace(namespaceId: string, serverId: string, priority?: number): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/namespaces/${namespaceId}/servers`, {
			method: 'POST',
			body: JSON.stringify({ server_id: serverId, priority })
		});
	},

	async removeServerFromNamespace(namespaceId: string, serverId: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/namespaces/${namespaceId}/servers/${serverId}`, {
			method: 'DELETE'
		});
	},

	async updateServerStatus(namespaceId: string, serverId: string, status: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/namespaces/${namespaceId}/servers/${serverId}/status`, {
			method: 'PUT',
			body: JSON.stringify({ status })
		});
	},

	async getNamespaceTools(namespaceId: string): Promise<NamespaceTool[]> {
		const response = await apiRequest<ApiResponse<NamespaceTool[]>>(`/namespaces/${namespaceId}/tools`);
		return response.data;
	},

	async updateToolStatus(namespaceId: string, toolId: string, status: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/namespaces/${namespaceId}/tools/${toolId}/status`, {
			method: 'PUT',
			body: JSON.stringify({ status })
		});
	},

	async executeNamespaceTool(namespaceId: string, request: ExecuteNamespaceToolRequest): Promise<any> {
		const response = await apiRequest<ApiResponse<any>>(`/namespaces/${namespaceId}/execute`, {
			method: 'POST',
			body: JSON.stringify(request)
		});
		return response.data;
	}
};

// Endpoint Management APIs
export const endpointApi = {
	async listEndpoints(): Promise<Endpoint[]> {
		const response = await apiRequest<{ endpoints: Endpoint[]; total: number }>('/endpoints');
		return response.endpoints || [];
	},

	async listPublicEndpoints(): Promise<Endpoint[]> {
		const response = await apiRequest<{ endpoints: Endpoint[]; total: number }>('/api/public/endpoints');
		return response.endpoints || [];
	},

	async getEndpoint(id: string): Promise<Endpoint> {
		const response = await apiRequest<Endpoint>(`/endpoints/${id}`);
		return response;
	},

	async createEndpoint(endpointData: CreateEndpointRequest): Promise<Endpoint> {
		const response = await apiRequest<Endpoint>('/endpoints', {
			method: 'POST',
			body: JSON.stringify(endpointData)
		});
		return response;
	},

	async updateEndpoint(id: string, endpointData: UpdateEndpointRequest): Promise<Endpoint> {
		const response = await apiRequest<Endpoint>(`/endpoints/${id}`, {
			method: 'PUT',
			body: JSON.stringify(endpointData)
		});
		return response;
	},

	async deleteEndpoint(id: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/endpoints/${id}`, {
			method: 'DELETE'
		});
	}
};

// A2A Agent Management Types
export interface A2AAgent {
	id: string;
	organization_id: string;
	name: string;
	description?: string;
	agent_type: string;
	is_active: boolean;
	capabilities: string[];
	allowed_namespaces?: string[];
	api_key_id?: string;
	configuration?: Record<string, any>;
	rate_limit?: number;
	rate_window?: string;
	tags?: string[];
	last_used_at?: string;
	created_at: string;
	updated_at: string;
	metrics?: {
		request_count: number;
		error_count: number;
		avg_response_time: number;
	};
}

export interface A2AAgentSpec {
	name: string;
	description?: string;
	agent_type: string;
	capabilities: string[];
	allowed_namespaces?: string[];
	configuration?: Record<string, any>;
	rate_limit?: number;
	rate_window?: string;
	tags?: string[];
	is_active?: boolean;
}

export interface A2AHealthCheck {
	agent_id: string;
	status: 'healthy' | 'unhealthy';
	last_check: string;
	response_time_ms: number;
	error?: string;
}

export interface A2AStats {
	total: number;
	active: number;
	inactive: number;
	by_type: Record<string, number>;
}

export interface A2ATestRequest {
	message: string;
	context?: string;
	parameters?: Record<string, any>;
}

export interface A2ATestResponse {
	success: boolean;
	content?: string;
	error?: string;
	execution_time_ms: number;
	tokens_used?: {
		prompt: number;
		completion: number;
		total: number;
	};
	metadata?: Record<string, any>;
}

// A2A Agent Management APIs
export const a2aApi = {
	// List all A2A agents with optional filtering
	async listAgents(filters?: {
		agent_type?: string;
		is_active?: boolean;
		health_status?: string;
		tags?: string;
	}): Promise<A2AAgent[]> {
		const queryParams = new URLSearchParams();

		if (filters) {
			if (filters.agent_type) queryParams.append('agent_type', filters.agent_type);

			if (filters.is_active !== undefined) queryParams.append('is_active', filters.is_active.toString());

			if (filters.health_status) queryParams.append('health_status', filters.health_status);

			if (filters.tags) queryParams.append('tags', filters.tags);
		}

		const response = await apiRequest<ApiResponse<A2AAgent[]>>(
			`/a2a${queryParams.toString() ? '?' + queryParams.toString() : ''}`
		);
		return response.data;
	},

	// Get specific A2A agent details
	async getAgent(id: string): Promise<A2AAgent> {
		const response = await apiRequest<ApiResponse<A2AAgent>>(`/a2a/${id}`);
		return response.data;
	},

	// Create new A2A agent
	async createAgent(agentData: A2AAgentSpec): Promise<A2AAgent> {
		const response = await apiRequest<ApiResponse<A2AAgent>>('/a2a', {
			method: 'POST',
			body: JSON.stringify(agentData)
		});
		return response.data;
	},

	// Update existing A2A agent
	async updateAgent(id: string, agentData: Partial<A2AAgentSpec>): Promise<A2AAgent> {
		const response = await apiRequest<ApiResponse<A2AAgent>>(`/a2a/${id}`, {
			method: 'PUT',
			body: JSON.stringify(agentData)
		});
		return response.data;
	},

	// Delete A2A agent
	async deleteAgent(id: string): Promise<void> {
		await apiRequest<ApiResponse<any>>(`/a2a/${id}`, {
			method: 'DELETE'
		});
	},

	// Toggle A2A agent active status
	async toggleAgent(id: string, active: boolean): Promise<A2AAgent> {
		const response = await apiRequest<ApiResponse<A2AAgent>>(`/a2a/${id}/toggle`, {
			method: 'POST',
			body: JSON.stringify({ active })
		});
		return response.data;
	},

	// Test A2A agent with a simple chat request
	async testAgent(id: string, testData: A2ATestRequest): Promise<A2ATestResponse> {
		const response = await apiRequest<ApiResponse<A2ATestResponse>>(`/a2a/${id}/test`, {
			method: 'POST',
			body: JSON.stringify(testData)
		});
		return response.data;
	},

	// Get A2A agent health check
	async checkAgentHealth(id: string): Promise<A2AHealthCheck> {
		const response = await apiRequest<ApiResponse<A2AHealthCheck>>(`/a2a/${id}/health`, {
			method: 'POST'
		});
		return response.data;
	},

	// Get A2A statistics for the organization
	async getStats(): Promise<A2AStats> {
		const response = await apiRequest<ApiResponse<A2AStats>>('/a2a/stats');
		return response.data;
	},

	// Get available agent types and their default configurations
	async getAgentTypes(): Promise<Record<string, any>> {
		const response = await apiRequest<ApiResponse<Record<string, any>>>('/a2a/types');
		return response.data;
	}
};

// Health check API
export const healthApi = {
	async checkHealth(): Promise<{ status: string }> {
		const response = await fetch(`${API_BASE_URL.replace('/api', '')}/health`, {
			mode: 'cors',
			credentials: 'include'
		});

		if (!response.ok) {
			throw new Error(`Health check failed: ${response.status}`);
		}

		return response.json();
	}
};
