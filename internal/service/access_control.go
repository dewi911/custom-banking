package service

import (
	"strings"
)

type AccessControl struct {
	roleRepository RoleRepository
	permissionMap  map[string]map[string][]string // role -> path -> methods
}

func NewAccessControl(roleRepository RoleRepository) *AccessControl {
	ac := &AccessControl{
		roleRepository: roleRepository,
		permissionMap:  make(map[string]map[string][]string),
	}

	ac.AddPermission("admin", "/*", []string{"GET", "POST", "PUT", "DELETE", "PATCH"})

	ac.AddPermission("user", "/account/*", []string{"GET", "POST", "PUT", "DELETE", "PATCH"})
	ac.AddPermission("user", "/card/", []string{"GET"})
	ac.AddPermission("user", "/event/", []string{"GET"})
	ac.AddPermission("user", "/loans/*", []string{"GET", "POST", "PUT", "DELETE", "PATCH"})
	ac.AddPermission("user", "/staking/*", []string{"GET", "POST", "PUT", "DELETE", "PATCH"})
	ac.AddPermission("user", "/api/auth/*", []string{"GET", "POST"})
	ac.AddPermission("user", "/api/accounts/*", []string{"GET"})
	ac.AddPermission("user", "/api/transactions/*", []string{"GET", "POST"})
	ac.AddPermission("user", "/api/cards/*", []string{"GET", "POST"})
	ac.AddPermission("user", "/api/events/*", []string{"GET"})

	ac.AddPermission("anonymous", "/auth/login", []string{"POST"})
	ac.AddPermission("anonymous", "/auth/register", []string{"POST"})
	ac.AddPermission("anonymous", "/auth/refresh", []string{"GET"})

	ac.AddPermission("anonymous", "/api/auth/login", []string{"POST"})
	ac.AddPermission("anonymous", "/api/auth/register", []string{"POST"})

	return ac
}

func (ac *AccessControl) AddPermission(role, path string, methods []string) {
	if _, exists := ac.permissionMap[role]; !exists {
		ac.permissionMap[role] = make(map[string][]string)
	}
	ac.permissionMap[role][path] = methods
}

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
