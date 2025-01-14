// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/config"
	"go.uber.org/zap"
)

type Handler interface {
	// SetTTL Configure TTL
	// @Tags API.config
	// @Router /api/config/setTTL [post]
	SetTTL() core.HandlerFunc

	// GetTTL Get TTL
	// @Tags API.config
	// @Router /api/config/getTTL [get]
	GetTTL() core.HandlerFunc

	// SetSingleTableTTL to configure the TTL of a single table
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
