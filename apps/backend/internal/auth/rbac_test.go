package auth

import (
	"testing"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestRBAC_RoleHierarchy(t *testing.T) {
	rbac := NewRBAC()

	// Test role hierarchy levels
	assert.Equal(t, 3, rbac.GetRoleLevel(types.RoleAdmin))
	assert.Equal(t, 2, rbac.GetRoleLevel(types.RoleUser))
	assert.Equal(t, 1, rbac.GetRoleLevel(types.RoleViewer))
	assert.Equal(t, 1, rbac.GetRoleLevel(types.RoleAPIUser))
	assert.Equal(t, 0, rbac.GetRoleLevel("invalid_role"))
}

func TestRBAC_HasRequiredRole(t *testing.T) {
	rbac := NewRBAC()

	// Admin can access everything (superuser)
	assert.True(t, rbac.HasRequiredRole(types.RoleAdmin, types.RoleAdmin))
	assert.True(t, rbac.HasRequiredRole(types.RoleAdmin, types.RoleUser))
	assert.True(t, rbac.HasRequiredRole(types.RoleAdmin, types.RoleViewer))

	// User can access viewer
	assert.True(t, rbac.HasRequiredRole(types.RoleUser, types.RoleViewer))
	assert.False(t, rbac.HasRequiredRole(types.RoleUser, types.RoleAdmin))

	// Viewer cannot access higher roles
	assert.False(t, rbac.HasRequiredRole(types.RoleViewer, types.RoleUser))
	assert.False(t, rbac.HasRequiredRole(types.RoleViewer, types.RoleAdmin))

	// Same level access
	assert.True(t, rbac.HasRequiredRole(types.RoleViewer, types.RoleViewer))
	assert.True(t, rbac.HasRequiredRole(types.RoleViewer, types.RoleAPIUser))
	assert.True(t, rbac.HasRequiredRole(types.RoleAPIUser, types.RoleViewer))
}

func TestRBAC_AdminPermissions(t *testing.T) {
	rbac := NewRBAC()

	// Admin should have all permissions (superuser)
	adminPerms := rbac.GetRolePermissions(types.RoleAdmin)

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
		assert.True(t, rbac.HasPermission(types.RoleAdmin, perm),
			"Admin should have permission: %s", perm)
	}

	assert.True(t, len(adminPerms) > 10, "Admin should have many permissions")

	// Admin is now superuser, so all permissions should return true
	assert.True(t, rbac.HasPermission(types.RoleAdmin, types.PermissionSystemManage))
	assert.True(t, rbac.HasPermission(types.RoleAdmin, types.PermissionOrgDelete))
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
	// Admin has superuser bypass, so can access everything
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
	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "user", "manage"))
	assert.False(t, rbac.CanAccessResource(types.RoleUser, "user", "manage"))
	assert.False(t, rbac.CanAccessResource(types.RoleViewer, "user", "manage"))

	// Test A2A agent resource access
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

	// Admin should have all permissions (superuser bypass)
	assert.True(t, rbac.HasAnyPermission(types.RoleAdmin, adminPerms))
	assert.True(t, rbac.HasAnyPermission(types.RoleAdmin, systemPerms))

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

	// Admin should have all permissions (superuser bypass)
	assert.True(t, rbac.HasAllPermissions(types.RoleAdmin, readWritePerms))
	assert.True(t, rbac.HasAllPermissions(types.RoleAdmin, allBasicPerms))
	assert.True(t, rbac.HasAllPermissions(types.RoleAdmin, adminPerms))

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
	assert.True(t, rbac.ValidateRole(types.RoleAdmin))
	assert.True(t, rbac.ValidateRole(types.RoleUser))
	assert.True(t, rbac.ValidateRole(types.RoleViewer))
	assert.True(t, rbac.ValidateRole(types.RoleAPIUser))

	// Invalid roles
	assert.False(t, rbac.ValidateRole("invalid_role"))
	assert.False(t, rbac.ValidateRole(""))
	assert.False(t, rbac.ValidateRole("super_admin"))
	assert.False(t, rbac.ValidateRole("system_admin"))
}

func TestRBAC_RoleElevation(t *testing.T) {
	rbac := NewRBAC()

	// Admin can elevate to any role (superuser)
	assert.True(t, rbac.CanElevateToRole(types.RoleAdmin, types.RoleAdmin))
	assert.True(t, rbac.CanElevateToRole(types.RoleAdmin, types.RoleUser))
	assert.True(t, rbac.CanElevateToRole(types.RoleAdmin, types.RoleViewer))
	assert.True(t, rbac.CanElevateToRole(types.RoleAdmin, types.RoleAPIUser))

	// User cannot elevate anyone
	assert.False(t, rbac.CanElevateToRole(types.RoleUser, types.RoleAdmin))
	assert.False(t, rbac.CanElevateToRole(types.RoleUser, types.RoleUser))
	assert.False(t, rbac.CanElevateToRole(types.RoleUser, types.RoleViewer))

	// Viewer cannot elevate anyone
	assert.False(t, rbac.CanElevateToRole(types.RoleViewer, types.RoleUser))
	assert.False(t, rbac.CanElevateToRole(types.RoleViewer, types.RoleViewer))
}

func TestRBAC_RoleCheckers(t *testing.T) {
	rbac := NewRBAC()

	// IsAdmin (admin is superuser)
	assert.True(t, rbac.IsAdmin(types.RoleAdmin))
	assert.False(t, rbac.IsAdmin(types.RoleUser))
	assert.False(t, rbac.IsAdmin(types.RoleViewer))

	// IsUser (user or higher)
	assert.True(t, rbac.IsUser(types.RoleAdmin))
	assert.True(t, rbac.IsUser(types.RoleUser))
	assert.False(t, rbac.IsUser(types.RoleViewer))
	assert.False(t, rbac.IsUser(types.RoleAPIUser))

	// IsViewer (viewer or higher)
	assert.True(t, rbac.IsViewer(types.RoleAdmin))
	assert.True(t, rbac.IsViewer(types.RoleUser))
	assert.True(t, rbac.IsViewer(types.RoleViewer))
	assert.True(t, rbac.IsViewer(types.RoleAPIUser))
}

func TestRBAC_GetAllRoles(t *testing.T) {
	rbac := NewRBAC()

	roles := rbac.GetAllRoles()
	assert.Len(t, roles, 4) // Now only 4 roles without system_admin

	expectedRoles := []string{
		types.RoleAdmin,
		types.RoleUser,
		types.RoleViewer,
		types.RoleAPIUser,
	}

	for _, expectedRole := range expectedRoles {
		assert.Contains(t, roles, expectedRole)
	}
}

func TestRBAC_AdminSuperuserBypass(t *testing.T) {
	rbac := NewRBAC()

	// Admin should always return true for any permission check (superuser bypass)
	// Test with random permission strings
	assert.True(t, rbac.HasPermission(types.RoleAdmin, "any_permission"))
	assert.True(t, rbac.HasPermission(types.RoleAdmin, "random_permission"))
	assert.True(t, rbac.HasPermission(types.RoleAdmin, "non_existent_permission"))

	// Admin should be able to access any resource with any action
	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "any_resource", "any_action"))
	assert.True(t, rbac.CanAccessResource(types.RoleAdmin, "random", "delete"))

	// Admin should be able to manage any resource
	assert.True(t, rbac.CanManageResource(types.RoleAdmin, "any_resource"))
	assert.True(t, rbac.CanManageResource(types.RoleAdmin, "random_resource"))

	// Other roles should not have this bypass
	assert.False(t, rbac.HasPermission(types.RoleUser, "non_existent_permission"))
	assert.False(t, rbac.HasPermission(types.RoleViewer, "random_permission"))
}
