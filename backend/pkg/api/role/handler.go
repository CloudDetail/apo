package role

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/role"
	"go.uber.org/zap"
)

type Handler interface {
	// RoleOperation Grant or revoke user's role.
	// @Tags API.role
	// @Router /api/role/operation [post]
	RoleOperation() core.HandlerFunc

	// GetRole Gets all roles.
	// @Tags API.role
	// @Router /api/role/roles [get]
	GetRole() core.HandlerFunc

	// GetUserRole Get user's role.
	// @Tags API.role
	// @Router /api/role/user [get]
	GetUserRole() core.HandlerFunc

	// CreateRole Creates a role.
	// @Tags API.role
	// @Router /api/role/create [post]
	CreateRole() core.HandlerFunc

	// UpdateRole Update role's name and permission.
	// @Tags API.role
	// @Router /api/role/update [post]
	UpdateRole() core.HandlerFunc

	// DeleteRole Delete a role.
	// @Tags API.role
	// @router /api/role/delete [post]
	DeleteRole() core.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	roleService role.Service
}

func New(logger *zap.Logger, dbRepo database.Repo) Handler {
	return &handler{
		logger:      logger,
		roleService: role.New(dbRepo),
	}

}
