package custom

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/custom"
	"go.uber.org/zap"
)

type Handler interface {

	// CreateCustomQuery 创建一个用户自定义查询数据类型
	// @Tags API.custom
	// @Router /api/custom/create [post]
	CreateCustomQuery() core.HandlerFunc

	// CetCustomQuery 获取用户自定义查询数据类型
	// @Tags API.custom
	// @Router /api/custom/get [get]
	GetCustomQuery() core.HandlerFunc

	// UpdateCustomQuery 更新用户自定义查询数据类型
	// @Tags API.custom
	// @Router /api/custom/update [post]
	UpdateCustomQuery() core.HandlerFunc

	// DeleteCustomQuery 删除用户自定义查询数据类型
	// @Tags API.custom
	// @Router /api/custom/delete [delete]
	DeleteCustomQuery() core.HandlerFunc
}

type handler struct {
	logger  *zap.Logger
	service custom.Service
}

func New(logger *zap.Logger, dbRepo database.Repo) Handler {
	return &handler{
		logger:  logger,
		service: custom.New(dbRepo),
	}
}
