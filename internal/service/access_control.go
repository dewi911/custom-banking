package service

import (
	"context"
	"custom-banking/internal/models"
	"strings"
)

// AccessControl provides a simple role-based access control mechanism
type AccessControl struct {
	roleRepository RoleRepository
	permissionMap  map[string]map[string][]string // role -> path -> methods
}

// NewAccessControl creates a new AccessControl instance
func NewAccessControl(roleRepository RoleRepository) *AccessControl {
	ac := &AccessControl{
		roleRepository: roleRepository,
		permissionMap:  make(map[string]map[string][]string),
	}

	ac.AddPermission("admin", "/*", []string{"GET", "POST", "PUT", "DELETE", "PATCH"})

	ac.AddPermission("user", "/api/auth/*", []string{"GET", "POST"})
	ac.AddPermission("user", "/api/accounts/*", []string{"GET"})
	ac.AddPermission("user", "/api/transactions/*", []string{"GET", "POST"})
	ac.AddPermission("user", "/api/cards/*", []string{"GET", "POST"})
	ac.AddPermission("user", "/api/events/*", []string{"GET"})

	ac.AddPermission("anonymous", "/api/auth/login", []string{"POST"})
	ac.AddPermission("anonymous", "/api/auth/register", []string{"POST"})

	return ac
}

// AddPermission adds a permission for a role
func (ac *AccessControl) AddPermission(role, path string, methods []string) {
	if _, exists := ac.permissionMap[role]; !exists {
		ac.permissionMap[role] = make(map[string][]string)
	}
	ac.permissionMap[role][path] = methods
}

// CheckPermission checks if a role has permission to access a path with a method
func (ac *AccessControl) CheckPermission(role, path, method string) bool {
	if role == "admin" {
		return true
	}

	if pathMap, exists := ac.permissionMap[role]; exists {
		if methods, pathExists := pathMap[path]; pathExists {
			for _, m := range methods {
				if m == method || m == "*" {
					return true
				}
			}
		}

		// Check wildcard paths
		for pattern, methods := range pathMap {
			if strings.HasSuffix(pattern, "/*") {
				prefix := strings.TrimSuffix(pattern, "/*")
				if strings.HasPrefix(path, prefix) {
					for _, m := range methods {
						if m == method || m == "*" {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// RoleRepository is the interface for accessing role data
type RoleRepository interface {
	GetByID(ctx context.Context, id int) (models.Role, error)
	GetByName(ctx context.Context, name string) (models.Role, error)
}
