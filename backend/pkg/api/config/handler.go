package config

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/config"
	"go.uber.org/zap"
)

type Handler interface {
	// SetTTL 配置TTL
	// @Tags API.config
	// @Router /api/config/setTTL [post]
	SetTTL() core.HandlerFunc

	// GetTTL 获取TTL
	// @Tags API.config
	// @Router /api/config/getTTL [get]
	GetTTL() core.HandlerFunc

	// SetSingleTableTTL 配置单个表格的TTL
	// @Tags API.config
	// @Router /api/config/setSingleTableTTL [post]
	SetSingleTableTTL() core.HandlerFunc
}

type handler struct {
	logger        *zap.Logger
	configService config.Service
}

func New(logger *zap.Logger, chRepo clickhouse.Repo) Handler {
	return &handler{
		logger:        logger,
		configService: config.New(chRepo),
	}
}
