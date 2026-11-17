// Package models define los tipos de datos persistidos en el schema 'auth'
// de la BD compartida.
package models

// Usuario es la fila de la tabla auth.users.
// Password es el hash bcrypt (no el texto plano).
type Usuario struct {
	ID       int
	Email    string
	Password string
	Rol      string // 'admin' o 'cliente' (rol B2C; lo emite el Backend en sus tokens)
}

// SaasUser es la fila de auth.saas_users.
// HashedPassword es bcrypt; TenantID + RolSaas ('superadmin'|'admin'|'member')
// viven aqui, NO en B2C.
type SaasUser struct {
	ID             int
	UserID         int    // FK auth.users.id (vincula con Usuario B2C si existe)
	TenantID       int    // FK auth.saas_tenants.id
	Email          string // email del SaaS user (puede ser igual al del Usuario B2C)
	HashedPassword string
	RolSaas        string // 'superadmin'|'admin'|'member'
}

// Tenant es la fila de auth.saas_tenants.
type Tenant struct {
	ID        int
	Slug      string // UNIQUE
	Name      string
	CreatedAt string
}
