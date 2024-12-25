// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
)

type Handler interface {
	// HealthCheck 用于k8s检查后端健康状态
	// @Tags API.health
	// @Router /api/health [get]
	HealthCheck() core.HandlerFunc
}

type handler struct {
}

func New() Handler {
	return &handler{}
}
