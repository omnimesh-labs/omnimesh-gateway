package auth

import (
	"testing"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestRBAC_RoleHierarchy(t *testing.T) {
	rbac := NewRBAC()

	// Test role hierarchy levels
	assert.Equal(t, 4, rbac.GetRoleLevel(types.RoleSystemAdmin))
	assert.Equal(t, 3, rbac.GetRoleLevel(types.RoleAdmin))
	assert.Equal(t, 2, rbac.GetRoleLevel(types.RoleUser))
	assert.Equal(t, 1, rbac.GetRoleLevel(types.RoleViewer))
	assert.Equal(t, 1, rbac.GetRoleLevel(types.RoleAPIUser))
	assert.Equal(t, 0, rbac.GetRoleLevel("invalid_role"))
}

func TestRBAC_HasRequiredRole(t *testing.T) {
	rbac := NewRBAC()

	// System admin can access everything
	assert.True(t, rbac.HasRequiredRole(types.RoleSystemAdmin, types.RoleAdmin))
	assert.True(t, rbac.HasRequiredRole(types.RoleSystemAdmin, types.RoleUser))
	assert.True(t, rbac.HasRequiredRole(types.RoleSystemAdmin, types.RoleViewer))

	// Admin can access user and viewer
	assert.True(t, rbac.HasRequiredRole(types.RoleAdmin, types.RoleUser))
	assert.True(t, rbac.HasRequiredRole(types.RoleAdmin, types.RoleViewer))
	assert.False(t, rbac.HasRequiredRole(types.RoleAdmin, types.RoleSystemAdmin))

	// User can access viewer
	assert.True(t, rbac.HasRequiredRole(types.RoleUser, types.RoleViewer))
	assert.False(t, rbac.HasRequiredRole(types.RoleUser, types.RoleAdmin))
	assert.False(t, rbac.HasRequiredRole(types.RoleUser, types.RoleSystemAdmin))

	// Viewer cannot access higher roles
	assert.False(t, rbac.HasRequiredRole(types.RoleViewer, types.RoleUser))
	assert.False(t, rbac.HasRequiredRole(types.RoleViewer, types.RoleAdmin))
	assert.False(t, rbac.HasRequiredRole(types.RoleViewer, types.RoleSystemAdmin))

	// Same level access
	assert.True(t, rbac.HasRequiredRole(types.RoleViewer, types.RoleViewer))
	assert.True(t, rbac.HasRequiredRole(types.RoleViewer, types.RoleAPIUser))
	assert.True(t, rbac.HasRequiredRole(types.RoleAPIUser, types.RoleViewer))
}

func TestRBAC_SystemAdminPermissions(t *testing.T) {
	rbac := NewRBAC()

	// System admin should have all permissions
	systemAdminPerms := rbac.GetRolePermissions(types.RoleSystemAdmin)

	expectedPermissions := []string{
		types.PermissionRead,
		types.PermissionWrite,
		types.PermissionDelete,
		types.PermissionAdmin,
		types.PermissionAPIAccess,
		types.PermissionAPIKeyManage,
		types.PermissionUserManage,
		types.PermissionServerManage,
		types.PermissionSessionManage,
		types.PermissionVirtualServerManage,
		types.PermissionAuditRead,
		types.PermissionLogsRead,
		types.PermissionMetricsRead,
		types.PermissionSystemManage,
		types.PermissionOrgManage,
	}

	for _, perm := range expectedPermissions {
		assert.True(t, rbac.HasPermission(types.RoleSystemAdmin, perm), 
			"System admin should have permission: %s", perm)
	}

	assert.True(t, len(systemAdminPerms) > 10, "System admin should have many permissions")
}

func TestRBAC_AdminPermissions(t *testing.T) {
	rbac := NewRBAC()

	// Admin should have organization-level permissions but not system-level
	assert.True(t, rbac.HasPermission(types.RoleAdmin, types.PermissionUserManage))
	assert.True(t, rbac.HasPermission(types.RoleAdmin, types.PermissionServerManage))
	assert.True(t, rbac.HasPermission(types.RoleAdmin, types.PermissionVirtualServerManage))
	assert.True(t, rbac.HasPermission(types.RoleAdmin, types.PermissionAuditRead))
	assert.True(t, rbac.HasPermission(types.RoleAdmin, types.PermissionOrgWrite))

	// Admin should NOT have system-level permissions
	assert.False(t, rbac.HasPermission(types.RoleAdmin, types.PermissionSystemManage))
	assert.False(t, rbac.HasPermission(types.RoleAdmin, types.PermissionOrgDelete))
}

func TestRBAC_UserPermissions(t *testing.T) {
	rbac := NewRBAC()

	// User should have basic operational permissions
	assert.True(t, rbac.HasPermission(types.RoleUser, types.PermissionRead))
	assert.True(t, rbac.HasPermission(types.RoleUser, types.PermissionWrite))
	assert.True(t, rbac.HasPermission(types.RoleUser, types.PermissionAPIAccess))
	assert.True(t, rbac.HasPermission(types.RoleUser, types.PermissionServerRead))
	assert.True(t, rbac.HasPermission(types.RoleUser, types.PermissionServerWrite))
	assert.True(t, rbac.HasPermission(types.RoleUser, types.PermissionSessionRead))
	assert.True(t, rbac.HasPermission(types.RoleUser, types.PermissionSessionWrite))

	// User should NOT have admin or delete permissions
	assert.False(t, rbac.HasPermission(types.RoleUser, types.PermissionDelete))
	assert.False(t, rbac.HasPermission(types.RoleUser, types.PermissionUserManage))
	assert.False(t, rbac.HasPermission(types.RoleUser, types.PermissionServerDelete))
	assert.False(t, rbac.HasPermission(types.RoleUser, types.PermissionAuditRead))
}

func TestRBAC_ViewerPermissions(t *testing.T) {
	rbac := NewRBAC()

	// Viewer should only have read permissions
	assert.True(t, rbac.HasPermission(types.RoleViewer, types.PermissionRead))
	assert.True(t, rbac.HasPermission(types.RoleViewer, types.PermissionAPIAccess))
	assert.True(t, rbac.HasPermission(types.RoleViewer, types.PermissionServerRead))
	assert.True(t, rbac.HasPermission(types.RoleViewer, types.PermissionSessionRead))
	assert.True(t, rbac.HasPermission(types.RoleViewer, types.PermissionVirtualServerRead))
	assert.True(t, rbac.HasPermission(types.RoleViewer, types.PermissionMetricsRead))
	assert.True(t, rbac.HasPermission(types.RoleViewer, types.PermissionOrgRead))

	// Viewer should NOT have write or delete permissions
	assert.False(t, rbac.HasPermission(types.RoleViewer, types.PermissionWrite))
	assert.False(t, rbac.HasPermission(types.RoleViewer, types.PermissionDelete))
	assert.False(t, rbac.HasPermission(types.RoleViewer, types.PermissionServerWrite))
	assert.False(t, rbac.HasPermission(types.RoleViewer, types.PermissionSessionWrite))
	assert.False(t, rbac.HasPermission(types.RoleViewer, types.PermissionUserManage))
}

func TestRBAC_CanAccessResource(t *testing.T) {
	rbac := NewRBAC()

	// Test server resource access
	assert.True(t, rbac.CanAccessResource(types.RoleSystemAdmin, "server", "read"))
	assert.True(t, rbac.CanAccessResource(types.RoleSystemAdmin, "server", "write"))
	assert.True(t, rbac.CanAccessResource(types.RoleSystemAdmin, "server", "delete"))

	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "server", "read"))
	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "server", "write"))
	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "server", "delete"))

	assert.True(t, rbac.CanAccessResource(types.RoleUser, "server", "read"))
	assert.True(t, rbac.CanAccessResource(types.RoleUser, "server", "write"))
	assert.False(t, rbac.CanAccessResource(types.RoleUser, "server", "delete"))

	assert.True(t, rbac.CanAccessResource(types.RoleViewer, "server", "read"))
	assert.False(t, rbac.CanAccessResource(types.RoleViewer, "server", "write"))
	assert.False(t, rbac.CanAccessResource(types.RoleViewer, "server", "delete"))

	// Test user resource access
	assert.True(t, rbac.CanAccessResource(types.RoleSystemAdmin, "user", "manage"))
	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "user", "manage"))
	assert.False(t, rbac.CanAccessResource(types.RoleUser, "user", "manage"))
	assert.False(t, rbac.CanAccessResource(types.RoleViewer, "user", "manage"))

	// Test A2A agent resource access
	assert.True(t, rbac.CanAccessResource(types.RoleSystemAdmin, "a2a_agent", "read"))
	assert.True(t, rbac.CanAccessResource(types.RoleSystemAdmin, "a2a_agent", "write"))
	assert.True(t, rbac.CanAccessResource(types.RoleSystemAdmin, "a2a_agent", "delete"))
	assert.True(t, rbac.CanAccessResource(types.RoleSystemAdmin, "a2a_agent", "execute"))

	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "a2a_agent", "read"))
	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "a2a_agent", "write"))
	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "a2a_agent", "delete"))
	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "a2a_agent", "execute"))

	assert.True(t, rbac.CanAccessResource(types.RoleUser, "a2a_agent", "read"))
	assert.True(t, rbac.CanAccessResource(types.RoleUser, "a2a_agent", "write"))
	assert.False(t, rbac.CanAccessResource(types.RoleUser, "a2a_agent", "delete"))
	assert.True(t, rbac.CanAccessResource(types.RoleUser, "a2a_agent", "execute"))

	assert.True(t, rbac.CanAccessResource(types.RoleViewer, "a2a_agent", "read"))
	assert.False(t, rbac.CanAccessResource(types.RoleViewer, "a2a_agent", "write"))
	assert.False(t, rbac.CanAccessResource(types.RoleViewer, "a2a_agent", "delete"))
	assert.False(t, rbac.CanAccessResource(types.RoleViewer, "a2a_agent", "execute"))
}

func TestRBAC_HasAnyPermission(t *testing.T) {
	rbac := NewRBAC()

	adminPerms := []string{types.PermissionUserManage, types.PermissionSystemManage}
	userPerms := []string{types.PermissionRead, types.PermissionWrite}
	systemPerms := []string{types.PermissionSystemManage, types.PermissionOrgDelete}

	// System admin should have any admin permission
	assert.True(t, rbac.HasAnyPermission(types.RoleSystemAdmin, adminPerms))
	assert.True(t, rbac.HasAnyPermission(types.RoleSystemAdmin, systemPerms))

	// Admin should have some admin permissions but not system permissions
	assert.True(t, rbac.HasAnyPermission(types.RoleAdmin, adminPerms))
	assert.False(t, rbac.HasAnyPermission(types.RoleAdmin, systemPerms))

	// User should have basic permissions
	assert.True(t, rbac.HasAnyPermission(types.RoleUser, userPerms))
	assert.False(t, rbac.HasAnyPermission(types.RoleUser, adminPerms))

	// Viewer should have read permission
	assert.True(t, rbac.HasAnyPermission(types.RoleViewer, userPerms))
	assert.False(t, rbac.HasAnyPermission(types.RoleViewer, adminPerms))
}

func TestRBAC_HasAllPermissions(t *testing.T) {
	rbac := NewRBAC()

	readWritePerms := []string{types.PermissionRead, types.PermissionWrite}
	allBasicPerms := []string{types.PermissionRead, types.PermissionWrite, types.PermissionDelete}
	adminPerms := []string{types.PermissionUserManage, types.PermissionServerManage}

	// System admin should have all permissions
	assert.True(t, rbac.HasAllPermissions(types.RoleSystemAdmin, readWritePerms))
	assert.True(t, rbac.HasAllPermissions(types.RoleSystemAdmin, allBasicPerms))
	assert.True(t, rbac.HasAllPermissions(types.RoleSystemAdmin, adminPerms))

	// Admin should have admin permissions but not all basic permissions (missing delete)
	assert.True(t, rbac.HasAllPermissions(types.RoleAdmin, adminPerms))
	assert.True(t, rbac.HasAllPermissions(types.RoleAdmin, readWritePerms))
	assert.True(t, rbac.HasAllPermissions(types.RoleAdmin, allBasicPerms))

	// User should have read/write but not admin permissions
	assert.True(t, rbac.HasAllPermissions(types.RoleUser, readWritePerms))
	assert.False(t, rbac.HasAllPermissions(types.RoleUser, allBasicPerms)) // missing delete
	assert.False(t, rbac.HasAllPermissions(types.RoleUser, adminPerms))

	// Viewer should only have read permission
	assert.False(t, rbac.HasAllPermissions(types.RoleViewer, readWritePerms)) // missing write
	assert.False(t, rbac.HasAllPermissions(types.RoleViewer, adminPerms))
}

func TestRBAC_RoleValidation(t *testing.T) {
	rbac := NewRBAC()

	// Valid roles
	assert.True(t, rbac.ValidateRole(types.RoleSystemAdmin))
	assert.True(t, rbac.ValidateRole(types.RoleAdmin))
	assert.True(t, rbac.ValidateRole(types.RoleUser))
	assert.True(t, rbac.ValidateRole(types.RoleViewer))
	assert.True(t, rbac.ValidateRole(types.RoleAPIUser))

	// Invalid roles
	assert.False(t, rbac.ValidateRole("invalid_role"))
	assert.False(t, rbac.ValidateRole(""))
	assert.False(t, rbac.ValidateRole("super_admin"))
}

func TestRBAC_RoleElevation(t *testing.T) {
	rbac := NewRBAC()

	// System admin can elevate to any role
	assert.True(t, rbac.CanElevateToRole(types.RoleSystemAdmin, types.RoleAdmin))
	assert.True(t, rbac.CanElevateToRole(types.RoleSystemAdmin, types.RoleUser))
	assert.True(t, rbac.CanElevateToRole(types.RoleSystemAdmin, types.RoleViewer))
	assert.True(t, rbac.CanElevateToRole(types.RoleSystemAdmin, types.RoleAPIUser))

	// Admin can elevate to user roles but not admin roles
	assert.False(t, rbac.CanElevateToRole(types.RoleAdmin, types.RoleSystemAdmin))
	assert.False(t, rbac.CanElevateToRole(types.RoleAdmin, types.RoleAdmin))
	assert.True(t, rbac.CanElevateToRole(types.RoleAdmin, types.RoleUser))
	assert.True(t, rbac.CanElevateToRole(types.RoleAdmin, types.RoleViewer))
	assert.True(t, rbac.CanElevateToRole(types.RoleAdmin, types.RoleAPIUser))

	// User cannot elevate anyone
	assert.False(t, rbac.CanElevateToRole(types.RoleUser, types.RoleSystemAdmin))
	assert.False(t, rbac.CanElevateToRole(types.RoleUser, types.RoleAdmin))
	assert.False(t, rbac.CanElevateToRole(types.RoleUser, types.RoleUser))
	assert.False(t, rbac.CanElevateToRole(types.RoleUser, types.RoleViewer))

	// Viewer cannot elevate anyone
	assert.False(t, rbac.CanElevateToRole(types.RoleViewer, types.RoleUser))
	assert.False(t, rbac.CanElevateToRole(types.RoleViewer, types.RoleViewer))
}

func TestRBAC_RoleCheckers(t *testing.T) {
	rbac := NewRBAC()

	// IsSystemAdmin
	assert.True(t, rbac.IsSystemAdmin(types.RoleSystemAdmin))
	assert.False(t, rbac.IsSystemAdmin(types.RoleAdmin))
	assert.False(t, rbac.IsSystemAdmin(types.RoleUser))
	assert.False(t, rbac.IsSystemAdmin(types.RoleViewer))

	// IsAdmin (admin or higher)
	assert.True(t, rbac.IsAdmin(types.RoleSystemAdmin))
	assert.True(t, rbac.IsAdmin(types.RoleAdmin))
	assert.False(t, rbac.IsAdmin(types.RoleUser))
	assert.False(t, rbac.IsAdmin(types.RoleViewer))

	// IsUser (user or higher)
	assert.True(t, rbac.IsUser(types.RoleSystemAdmin))
	assert.True(t, rbac.IsUser(types.RoleAdmin))
	assert.True(t, rbac.IsUser(types.RoleUser))
	assert.False(t, rbac.IsUser(types.RoleViewer))
	assert.False(t, rbac.IsUser(types.RoleAPIUser))

	// IsViewer (viewer or higher)
	assert.True(t, rbac.IsViewer(types.RoleSystemAdmin))
	assert.True(t, rbac.IsViewer(types.RoleAdmin))
	assert.True(t, rbac.IsViewer(types.RoleUser))
	assert.True(t, rbac.IsViewer(types.RoleViewer))
	assert.True(t, rbac.IsViewer(types.RoleAPIUser))
}

func TestRBAC_GetAllRoles(t *testing.T) {
	rbac := NewRBAC()

	roles := rbac.GetAllRoles()
	assert.Len(t, roles, 5)

	expectedRoles := []string{
		types.RoleSystemAdmin,
		types.RoleAdmin,
		types.RoleUser,
		types.RoleViewer,
		types.RoleAPIUser,
	}

	for _, expectedRole := range expectedRoles {
		assert.Contains(t, roles, expectedRole)
	}
}