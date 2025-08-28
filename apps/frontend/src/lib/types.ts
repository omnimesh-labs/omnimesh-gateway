// Proper TypeScript types to replace 'any' usages

// Parameter types for prompts and tools
export interface PromptParameter {
	name: string;
	type: 'string' | 'number' | 'boolean' | 'object' | 'array';
	description?: string;
	required?: boolean;
	default?: unknown;
	enum?: string[];
	pattern?: string;
	min?: number;
	max?: number;
	minLength?: number;
	maxLength?: number;
}

// JSON Schema for tools
export interface JSONSchema {
	type?: 'string' | 'number' | 'integer' | 'boolean' | 'object' | 'array' | 'null';
	properties?: Record<string, JSONSchema>;
	required?: string[];
	items?: JSONSchema;
	enum?: unknown[];
	const?: unknown;
	title?: string;
	description?: string;
	default?: unknown;
	examples?: unknown[];
	minimum?: number;
	maximum?: number;
	exclusiveMinimum?: number;
	exclusiveMaximum?: number;
	multipleOf?: number;
	minLength?: number;
	maxLength?: number;
	pattern?: string;
	format?: string;
	minItems?: number;
	maxItems?: number;
	uniqueItems?: boolean;
	minProperties?: number;
	maxProperties?: number;
	additionalProperties?: boolean | JSONSchema;
	allOf?: JSONSchema[];
	anyOf?: JSONSchema[];
	oneOf?: JSONSchema[];
	not?: JSONSchema;
	$ref?: string;
	$schema?: string;
	$id?: string;
	$defs?: Record<string, JSONSchema>;
}

// Tool example structure
export interface ToolExample {
	name: string;
	description?: string;
	input: Record<string, unknown>;
	output?: unknown;
	explanation?: string;
}

// Policy rule conditions and actions
export interface PolicyCondition {
	field: string;
	operator: 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'in' | 'nin' | 'contains' | 'regex' | 'exists';
	value: unknown;
	caseSensitive?: boolean;
}

export interface PolicyAction {
	type: 'allow' | 'deny' | 'require_auth' | 'rate_limit' | 'log' | 'redirect' | 'transform' | 'notify';
	parameters?: Record<string, unknown>;
	message?: string;
	severity?: 'info' | 'warn' | 'error' | 'critical';
}

export interface PolicyRules {
	conditions: PolicyCondition[];
	actions: PolicyAction[];
	matchType?: 'all' | 'any';
	priority?: number;
}

// Resource access permissions
export interface AccessPermissions {
	read?: string[];
	write?: string[];
	execute?: string[];
	delete?: string[];
	share?: string[];
	public?: boolean;
	roles?: string[];
	users?: string[];
	groups?: string[];
}

// Metadata structures
export interface ResourceMetadata {
	version?: string;
	author?: string;
	license?: string;
	source?: string;
	lastModified?: string;
	contentHash?: string;
	encoding?: string;
	compression?: string;
	custom?: Record<string, unknown>;
}

export interface PromptMetadata {
	version?: string;
	author?: string;
	model?: string;
	temperature?: number;
	maxTokens?: number;
	topP?: number;
	frequencyPenalty?: number;
	presencePenalty?: number;
	stopSequences?: string[];
	custom?: Record<string, unknown>;
}

export interface ToolMetadata {
	version?: string;
	author?: string;
	documentation?: string;
	apiVersion?: string;
	deprecated?: boolean;
	deprecationMessage?: string;
	replacement?: string;
	custom?: Record<string, unknown>;
}

// API Error details
export interface ApiErrorDetails {
	field?: string;
	code?: string;
	constraint?: string;
	value?: unknown;
	expected?: unknown;
	actual?: unknown;
	suggestion?: string;
	documentation?: string;
	stackTrace?: string[];
	innerError?: ApiErrorDetails;
}

// Audit log value types
export interface AuditLogValues {
	[key: string]: string | number | boolean | null | AuditLogValues | AuditLogValues[];
}

// Export data structures
export interface ExportedServer {
	name: string;
	description?: string;
	protocol: string;
	url?: string;
	command?: string;
	args?: string[];
	environment?: string[];
	working_dir?: string;
	version?: string;
	timeout?: number;
	max_retries?: number;
	health_check_url?: string;
	metadata?: Record<string, string>;
}

export interface ExportedVirtualServer {
	name: string;
	description?: string;
	adapter_type: string;
	tools: JSONSchema[];
	metadata?: Record<string, unknown>;
}

export interface ExportedTool {
	name: string;
	description?: string;
	function_name: string;
	schema: JSONSchema;
	category: string;
	implementation_type?: string;
	endpoint_url?: string;
	timeout_seconds?: number;
	max_retries?: number;
	access_permissions?: AccessPermissions;
	is_public?: boolean;
	metadata?: ToolMetadata;
	tags?: string[];
	examples?: ToolExample[];
	documentation?: string;
}

export interface ExportedPrompt {
	name: string;
	description?: string;
	prompt_template: string;
	parameters?: PromptParameter[];
	category: string;
	metadata?: PromptMetadata;
	tags?: string[];
}

export interface ExportedResource {
	name: string;
	description?: string;
	type: string;
	uri?: string;
	content?: string;
	mime_type?: string;
	size_bytes?: number;
	access_permissions?: AccessPermissions;
	metadata?: ResourceMetadata;
	tags?: string[];
}

export interface ExportedPolicy {
	name: string;
	description?: string;
	type: string;
	scope: string;
	priority: number;
	conditions: PolicyCondition[];
	actions: PolicyAction[];
}

export interface ExportedRateLimit {
	name: string;
	description?: string;
	type: string;
	limit: number;
	window: string;
	scope: string;
	conditions?: PolicyCondition[];
}

// Import/Export metadata
export interface ImportExportMetadata {
	version: string;
	format: string;
	timestamp: string;
	source?: string;
	description?: string;
	author?: string;
	organization?: string;
	tags?: string[];
}

// Conflict resolution types
export interface ConflictResolution {
	strategy: 'skip' | 'overwrite' | 'rename' | 'merge';
	renamePattern?: string;
	mergeFields?: string[];
}

// Server stats
export interface ServerStats {
	totalRequests: number;
	successfulRequests: number;
	failedRequests: number;
	averageResponseTime: number;
	uptime: number;
	lastHealthCheck?: string;
	healthStatus: 'healthy' | 'unhealthy' | 'unknown';
	activeConnections: number;
	errorRate: number;
	throughput: number;
	latencyP50?: number;
	latencyP95?: number;
	latencyP99?: number;
}

// Session data
export interface Session {
	id: string;
	server_id: string;
	server_name?: string;
	protocol: string;
	status: 'initializing' | 'active' | 'idle' | 'closing' | 'closed' | 'error';
	client_id?: string;
	connection_id?: string;
	started_at: string;
	last_activity?: string;
	ended_at?: string;
	metadata?: Record<string, unknown>;
	error?: string;
}

// Generic function types for utilities
export type AnyFunction = (...args: unknown[]) => unknown;
export type AnyAsyncFunction = (...args: unknown[]) => Promise<unknown>;

// Cache entry type
export interface CacheEntry<T> {
	data: T;
	timestamp: number;
	ttl?: number;
	tags?: string[];
}

// A2A Agent types
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
	configuration?: Record<string, unknown>;
	created_at: string;
	updated_at: string;
}

export interface A2AAgentSpec {
	name: string;
	description?: string;
	agent_type: string;
	capabilities: string[];
	allowed_namespaces?: string[];
	configuration?: Record<string, unknown>;
}

export interface A2AStats {
	total_agents: number;
	active_agents: number;
	total_requests: number;
	successful_requests: number;
	failed_requests: number;
	average_response_time: number;
}

// Endpoint types
export interface Endpoint {
	id: string;
	organization_id: string;
	namespace_id?: string;
	namespace?: Namespace;
	name: string;
	description?: string;
	enable_api_key_auth: boolean;
	enable_oauth: boolean;
	enable_public_access: boolean;
	rate_limit_requests: number;
	rate_limit_window: number;
	is_active: boolean;
	urls?: Record<string, string>;
	created_at: string;
	updated_at: string;
}

export interface CreateEndpointRequest {
	name: string;
	namespace_id: string;
	description?: string;
	enable_api_key_auth: boolean;
	enable_oauth: boolean;
	enable_public_access: boolean;
	rate_limit_requests: number;
	rate_limit_window: number;
}

export interface UpdateEndpointRequest {
	name?: string;
	description?: string;
	enable_api_key_auth?: boolean;
	enable_oauth?: boolean;
	enable_public_access?: boolean;
	rate_limit_requests?: number;
	rate_limit_window?: number;
	is_active?: boolean;
}

export interface Namespace {
	id: string;
	organization_id: string;
	name: string;
	slug: string;
	description?: string;
	is_active: boolean;
	metadata?: Record<string, unknown>;
	created_at: string;
	updated_at: string;
}

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

// Auth types
export interface User {
	id: string;
	email: string;
	name: string;
	organization_id: string;
	role: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface LoginRequest {
	email: string;
	password: string;
}

export interface LoginResponse {
	user: User;
	access_token: string;
	refresh_token: string;
}

export interface RefreshResponse {
	user: User;
	access_token: string;
	refresh_token: string;
}

export interface ApiKey {
	id: string;
	name: string;
	prefix: string;
	expires_at?: string;
	last_used_at?: string;
	is_active: boolean;
	created_at: string;
}

export interface CreateApiKeyRequest {
	name: string;
	expires_at?: string;
}

export interface CreateApiKeyResponse {
	apiKey: ApiKey;
	key: string;
}

// Dashboard types
export interface SystemStats {
	active_servers: number;
	total_requests: number;
	uptime: string;
	memory_usage: number;
	cpu_usage: number;
}

// Logging types
export interface LogEntry {
	id: string;
	timestamp: string;
	level: string;
	message: string;
	server_id?: string;
	user_id?: string;
	metadata?: Record<string, unknown>;
}

export interface AuditLogEntry {
	id: string;
	timestamp: string;
	user_id: string;
	action: string;
	resource_type: string;
	resource_id?: string;
	details?: Record<string, unknown>;
}

export interface LogQueryParams {
	level?: string;
	server_id?: string;
	user_id?: string;
	start_date?: string;
	end_date?: string;
	limit?: number;
	offset?: number;
}

export interface AuditQueryParams {
	user_id?: string;
	action?: string;
	resource_type?: string;
	start_date?: string;
	end_date?: string;
	limit?: number;
	offset?: number;
}

// Policy types (placeholder for now)
export interface Policy {
	id: string;
	name: string;
	description?: string;
	type: string;
	scope: string;
	priority: number;
	conditions: PolicyCondition[];
	actions: PolicyAction[];
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

// Server management types
export interface CreateServerRequest {
	name: string;
	namespace_id: string;
	description?: string;
	url?: string;
	command?: string;
	args?: string[];
	environment?: string[];
	working_dir?: string;
	timeout_seconds?: number;
	max_retries?: number;
}

export interface MCPDiscoveryResponse {
	tools: {
		name: string;
		description?: string;
		schema: JSONSchema;
	}[];
	resources: {
		uri: string;
		name?: string;
		description?: string;
		mimeType?: string;
	}[];
	prompts: {
		name: string;
		description?: string;
		arguments?: {
			name: string;
			description?: string;
			required?: boolean;
		}[];
	}[];
}
