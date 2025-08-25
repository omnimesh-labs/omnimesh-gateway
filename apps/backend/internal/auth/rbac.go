package auth

import (
	"strings"

	"mcp-gateway/apps/backend/internal/types"
)

// RBAC implements Role-Based Access Control
type RBAC struct {
	rolePermissions map[string][]string
	roleHierarchy   map[string]int
}

// NewRBAC creates a new RBAC instance with predefined role permissions
func NewRBAC() *RBAC {
	rbac := &RBAC{
		rolePermissions: make(map[string][]string),
		roleHierarchy: map[string]int{
			types.RoleSystemAdmin: 4,
			types.RoleAdmin:       3,
			types.RoleUser:        2,
			types.RoleViewer:      1,
			types.RoleAPIUser:     1, // Same level as viewer
		},
	}

	// Define permissions for each role
	rbac.defineRolePermissions()
	return rbac
}

// defineRolePermissions sets up the permission mapping for each role
func (r *RBAC) defineRolePermissions() {
	// System Admin - Full access to everything
	r.rolePermissions[types.RoleSystemAdmin] = []string{
		// All basic permissions
		types.PermissionRead,
		types.PermissionWrite,
		types.PermissionDelete,
		types.PermissionAdmin,

		// API access
		types.PermissionAPIAccess,
		types.PermissionAPIKeyManage,

		// User management
		types.PermissionUserRead,
		types.PermissionUserWrite,
		types.PermissionUserDelete,
		types.PermissionUserManage,

		// Server management
		types.PermissionServerRead,
		types.PermissionServerWrite,
		types.PermissionServerDelete,
		types.PermissionServerManage,

		// Session management
		types.PermissionSessionRead,
		types.PermissionSessionWrite,
		types.PermissionSessionDelete,
		types.PermissionSessionManage,

		// Virtual server management
		types.PermissionVirtualServerRead,
		types.PermissionVirtualServerWrite,
		types.PermissionVirtualServerDelete,
		types.PermissionVirtualServerManage,

		// Resource management
		types.PermissionResourceRead,
		types.PermissionResourceWrite,
		types.PermissionResourceDelete,
		types.PermissionResourceManage,

		// Prompt management
		types.PermissionPromptRead,
		types.PermissionPromptWrite,
		types.PermissionPromptDelete,
		types.PermissionPromptManage,

		// Tool management
		types.PermissionToolRead,
		types.PermissionToolWrite,
		types.PermissionToolDelete,
		types.PermissionToolManage,
		types.PermissionToolExecute,

		// Audit and logging
		types.PermissionAuditRead,
		types.PermissionLogsRead,
		types.PermissionMetricsRead,
		types.PermissionSystemManage,

		// Organization management
		types.PermissionOrgRead,
		types.PermissionOrgWrite,
		types.PermissionOrgDelete,
		types.PermissionOrgManage,

		// A2A Agent management
		types.PermissionA2AAgentRead,
		types.PermissionA2AAgentWrite,
		types.PermissionA2AAgentDelete,
		types.PermissionA2AAgentManage,
		types.PermissionA2AAgentExecute,

		// Namespace management
		types.PermissionNamespaceRead,
		types.PermissionNamespaceWrite,
		types.PermissionNamespaceDelete,
		types.PermissionNamespaceManage,
		types.PermissionNamespaceExecute,
	}

	// Admin - Organization-level admin permissions
	r.rolePermissions[types.RoleAdmin] = []string{
		// Basic permissions
		types.PermissionRead,
		types.PermissionWrite,
		types.PermissionDelete,

		// API access
		types.PermissionAPIAccess,
		types.PermissionAPIKeyManage,

		// User management (within organization)
		types.PermissionUserRead,
		types.PermissionUserWrite,
		types.PermissionUserDelete,
		types.PermissionUserManage,

		// Server management
		types.PermissionServerRead,
		types.PermissionServerWrite,
		types.PermissionServerDelete,
		types.PermissionServerManage,

		// Session management
		types.PermissionSessionRead,
		types.PermissionSessionWrite,
		types.PermissionSessionDelete,
		types.PermissionSessionManage,

		// Virtual server management
		types.PermissionVirtualServerRead,
		types.PermissionVirtualServerWrite,
		types.PermissionVirtualServerDelete,
		types.PermissionVirtualServerManage,

		// Resource management
		types.PermissionResourceRead,
		types.PermissionResourceWrite,
		types.PermissionResourceDelete,
		types.PermissionResourceManage,

		// Prompt management
		types.PermissionPromptRead,
		types.PermissionPromptWrite,
		types.PermissionPromptDelete,
		types.PermissionPromptManage,

		// Tool management
		types.PermissionToolRead,
		types.PermissionToolWrite,
		types.PermissionToolDelete,
		types.PermissionToolManage,
		types.PermissionToolExecute,

		// Audit and logging (read-only)
		types.PermissionAuditRead,
		types.PermissionLogsRead,
		types.PermissionMetricsRead,

		// Organization (read and update, but not delete)
		types.PermissionOrgRead,
		types.PermissionOrgWrite,

		// A2A Agent management
		types.PermissionA2AAgentRead,
		types.PermissionA2AAgentWrite,
		types.PermissionA2AAgentDelete,
		types.PermissionA2AAgentManage,
		types.PermissionA2AAgentExecute,

		// Namespace management
		types.PermissionNamespaceRead,
		types.PermissionNamespaceWrite,
		types.PermissionNamespaceDelete,
		types.PermissionNamespaceManage,
		types.PermissionNamespaceExecute,
	}

	// User - Regular user permissions
	r.rolePermissions[types.RoleUser] = []string{
		// Basic read/write permissions
		types.PermissionRead,
		types.PermissionWrite,

		// API access
		types.PermissionAPIAccess,

		// Server management (read and create, limited delete)
		types.PermissionServerRead,
		types.PermissionServerWrite,

		// Session management (own sessions)
		types.PermissionSessionRead,
		types.PermissionSessionWrite,
		types.PermissionSessionDelete,

		// Virtual server management (own virtual servers)
		types.PermissionVirtualServerRead,
		types.PermissionVirtualServerWrite,

		// Resource management (limited - no delete)
		types.PermissionResourceRead,
		types.PermissionResourceWrite,

		// Prompt management (limited - no delete)
		types.PermissionPromptRead,
		types.PermissionPromptWrite,

		// Tool management (limited - no delete but can execute)
		types.PermissionToolRead,
		types.PermissionToolWrite,
		types.PermissionToolExecute,

		// Basic metrics (own data)
		types.PermissionMetricsRead,

		// Organization read-only
		types.PermissionOrgRead,

		// A2A Agent management (limited - no delete)
		types.PermissionA2AAgentRead,
		types.PermissionA2AAgentWrite,
		types.PermissionA2AAgentExecute,

		// Namespace management (limited - no delete)
		types.PermissionNamespaceRead,
		types.PermissionNamespaceWrite,
		types.PermissionNamespaceExecute,
	}

	// Viewer - Read-only permissions
	r.rolePermissions[types.RoleViewer] = []string{
		// Basic read permission
		types.PermissionRead,

		// API access (read-only)
		types.PermissionAPIAccess,

		// Server read-only
		types.PermissionServerRead,

		// Session read-only
		types.PermissionSessionRead,

		// Virtual server read-only
		types.PermissionVirtualServerRead,

		// Resource read-only
		types.PermissionResourceRead,

		// Prompt read-only
		types.PermissionPromptRead,

		// Tool read-only
		types.PermissionToolRead,

		// Basic metrics (read-only)
		types.PermissionMetricsRead,

		// Organization read-only
		types.PermissionOrgRead,

		// A2A Agent read-only
		types.PermissionA2AAgentRead,

		// Namespace read-only
		types.PermissionNamespaceRead,
	}

	// API User - Similar to viewer but for API access
	r.rolePermissions[types.RoleAPIUser] = []string{
		// Basic read permission
		types.PermissionRead,

		// API access
		types.PermissionAPIAccess,

		// Server read-only
		types.PermissionServerRead,

		// Session read-only
		types.PermissionSessionRead,

		// Virtual server read-only
		types.PermissionVirtualServerRead,

		// Resource read-only
		types.PermissionResourceRead,

		// Prompt read-only
		types.PermissionPromptRead,

		// Tool read-only (and execute for API users)
		types.PermissionToolRead,
		types.PermissionToolExecute,

		// Organization read-only
		types.PermissionOrgRead,

		// A2A Agent read and execute for API users
		types.PermissionA2AAgentRead,
		types.PermissionA2AAgentExecute,

		// Namespace read and execute for API users
		types.PermissionNamespaceRead,
		types.PermissionNamespaceExecute,
	}
}

// HasPermission checks if a role has a specific permission
func (r *RBAC) HasPermission(role, permission string) bool {
	permissions, exists := r.rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a role has any of the specified permissions
func (r *RBAC) HasAnyPermission(role string, permissions []string) bool {
	for _, permission := range permissions {
		if r.HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a role has all of the specified permissions
func (r *RBAC) HasAllPermissions(role string, permissions []string) bool {
	for _, permission := range permissions {
		if !r.HasPermission(role, permission) {
			return false
		}
	}
	return true
}

// HasRequiredRole checks if user role meets the required role level
func (r *RBAC) HasRequiredRole(userRole, requiredRole string) bool {
	userLevel := r.roleHierarchy[userRole]
	requiredLevel := r.roleHierarchy[requiredRole]
	return userLevel >= requiredLevel
}

// GetRolePermissions returns all permissions for a role
func (r *RBAC) GetRolePermissions(role string) []string {
	permissions, exists := r.rolePermissions[role]
	if !exists {
		return []string{}
	}

	// Return a copy to prevent modification
	result := make([]string, len(permissions))
	copy(result, permissions)
	return result
}

// GetRoleLevel returns the hierarchy level for a role
func (r *RBAC) GetRoleLevel(role string) int {
	level, exists := r.roleHierarchy[role]
	if !exists {
		return 0
	}
	return level
}

// CanAccessResource checks if a role can access a specific resource type with an action
func (r *RBAC) CanAccessResource(role, resource, action string) bool {
	// Build permission string: resource_action (e.g., server_read, user_write)
	permission := strings.ToLower(resource + "_" + action)
	return r.HasPermission(role, permission)
}

// CanManageResource checks if a role has management permissions for a resource
func (r *RBAC) CanManageResource(role, resource string) bool {
	managePermission := strings.ToLower(resource + "_manage")
	return r.HasPermission(role, managePermission)
}

// IsSystemAdmin checks if the role is system admin
func (r *RBAC) IsSystemAdmin(role string) bool {
	return role == types.RoleSystemAdmin
}

// IsAdmin checks if the role is admin or higher
func (r *RBAC) IsAdmin(role string) bool {
	return r.HasRequiredRole(role, types.RoleAdmin)
}

// IsUser checks if the role is user or higher
func (r *RBAC) IsUser(role string) bool {
	return r.HasRequiredRole(role, types.RoleUser)
}

// IsViewer checks if the role is viewer or higher
func (r *RBAC) IsViewer(role string) bool {
	return r.HasRequiredRole(role, types.RoleViewer)
}

// ValidateRole checks if a role is valid
func (r *RBAC) ValidateRole(role string) bool {
	_, exists := r.roleHierarchy[role]
	return exists
}

// GetAllRoles returns all available roles
func (r *RBAC) GetAllRoles() []string {
	roles := make([]string, 0, len(r.roleHierarchy))
	for role := range r.roleHierarchy {
		roles = append(roles, role)
	}
	return roles
}

// CanElevateToRole checks if a user with currentRole can elevate someone to targetRole
func (r *RBAC) CanElevateToRole(currentRole, targetRole string) bool {
	// System admins can elevate to any role
	if r.IsSystemAdmin(currentRole) {
		return true
	}

	// Admins can elevate to user, viewer, or api_user, but not to admin or system_admin
	if r.IsAdmin(currentRole) {
		return targetRole == types.RoleUser ||
			targetRole == types.RoleViewer ||
			targetRole == types.RoleAPIUser
	}

	// Regular users cannot elevate anyone
	return false
}
