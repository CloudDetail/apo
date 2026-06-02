// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package customdashboard

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/customdashboard"
	"go.uber.org/zap"
)

type Handler interface {
	List() core.HandlerFunc
	Get() core.HandlerFunc
	Create() core.HandlerFunc
	Update() core.HandlerFunc
	Copy() core.HandlerFunc
	Delete() core.HandlerFunc
	Templates() core.HandlerFunc
	PresetQueries() core.HandlerFunc
	VariableCandidates() core.HandlerFunc
	VariableCandidatesQuery() core.HandlerFunc
	ValidateQuery() core.HandlerFunc
	PreviewQuery() core.HandlerFunc
}

type handler struct {
	logger  *zap.Logger
	service customdashboard.Service
}

func New(logger *zap.Logger, dbRepo database.Repo, prom prometheus.Repo) Handler {
	return &handler{logger: logger, service: customdashboard.New(dbRepo, prom)}
}
