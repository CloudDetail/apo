// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/CloudDetail/apo/backend/internal/repository/database"
	"github.com/CloudDetail/apo/backend/internal/service/mock"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"go.uber.org/zap"
)

type Handler interface {
	// Create/Edit xx
	// @Tags API.mock
	// @Router /api/mock [post]
	Create() core.HandlerFunc

	// List xx list
	// @Tags API.mock
	// @Router /api/mock [get]
	List() core.HandlerFunc

	// Detail xx Details
	// @Tags API.mock
	// @Router /api/mock/{id} [get]
	Detail() core.HandlerFunc

	// Delete Delete xx
	// @Tags API.mock
	// @Router /api/mock/{id} [delete]
	Delete() core.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	mockService mock.Service
}

func New(logger *zap.Logger, dbRepo database.Repo) Handler {
	return &handler{
		logger:      logger,
		mockService: mock.New(dbRepo),
	}
}
