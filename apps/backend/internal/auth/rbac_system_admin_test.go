package auth

import (
	"testing"

	"mcp-gateway/apps/backend/internal/types"
)

func TestSystemAdminHasAllAccess(t *testing.T) {
	rbac := NewRBAC()

	// Test that system admin has access to all resources
	resources := []string{"namespace", "server", "user", "organization", "session", "tool", "prompt", "resource"}
	actions := []string{"read", "write", "delete", "manage", "execute", "admin"}

	for _, resource := range resources {
		for _, action := range actions {
			if !rbac.CanAccessResource(types.RoleSystemAdmin, resource, action) {
				t.Errorf("System admin should have access to %s:%s", resource, action)
			}
		}
	}

	// Test that system admin has all specific permissions
	permissions := []string{
		types.PermissionNamespaceRead,
		types.PermissionNamespaceWrite,
		types.PermissionNamespaceDelete,
		types.PermissionNamespaceManage,
		types.PermissionNamespaceExecute,
		types.PermissionNamespaceAdmin,
		types.PermissionEndpointAccess,
		types.PermissionEndpointAdmin,
	}

	for _, permission := range permissions {
		if !rbac.HasPermission(types.RoleSystemAdmin, permission) {
			t.Errorf("System admin should have permission: %s", permission)
		}
	}

	// Test HasAnyPermission with system admin
	if !rbac.HasAnyPermission(types.RoleSystemAdmin, []string{"random_permission", "another_random"}) {
		t.Error("System admin should have any permission")
	}

	// Test HasAllPermissions with system admin
	if !rbac.HasAllPermissions(types.RoleSystemAdmin, []string{"random_permission", "another_random", "third_random"}) {
		t.Error("System admin should have all permissions")
	}

	// Test CanManageResource with system admin
	for _, resource := range resources {
		if !rbac.CanManageResource(types.RoleSystemAdmin, resource) {
			t.Errorf("System admin should be able to manage resource: %s", resource)
		}
	}

	// Test special access methods
	if !rbac.CanAccessNamespaceEndpoints(types.RoleSystemAdmin) {
		t.Error("System admin should have access to namespace endpoints")
	}

	if !rbac.CanAccessAllEndpoints(types.RoleSystemAdmin) {
		t.Error("System admin should have access to all endpoints")
	}
}

func TestNonSystemAdminLimitedAccess(t *testing.T) {
	rbac := NewRBAC()

	// Test that regular admin doesn't have system admin's universal bypass
	// They should only have the permissions explicitly assigned to them

	// Admin should NOT have namespace_admin permission (which we only gave to system_admin)
	if rbac.HasPermission(types.RoleAdmin, types.PermissionNamespaceAdmin) {
		t.Error("Regular admin should not have namespace_admin permission")
	}

	// Admin should NOT have endpoint_admin permission
	if rbac.HasPermission(types.RoleAdmin, types.PermissionEndpointAdmin) {
		t.Error("Regular admin should not have endpoint_admin permission")
	}

	// But admin should have namespace_read permission (explicitly assigned)
	if !rbac.HasPermission(types.RoleAdmin, types.PermissionNamespaceRead) {
		t.Error("Regular admin should have namespace_read permission")
	}
}
