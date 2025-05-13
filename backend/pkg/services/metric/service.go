// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

var _ Service = (*service)(nil)

type Service interface {
	ListPreDefinedMetrics(ctx_core core.Context,) []QueryInfo
	ListQuerys(ctx_core core.Context,) []Query
	QueryMetrics(ctx_core core.Context, req *QueryMetricsRequest) *QueryMetricsResult
}

type service struct {
	promRepo prometheus.Repo
}

func New(promRepo prometheus.Repo) Service {
	return &service{
		promRepo: promRepo,
	}
}
