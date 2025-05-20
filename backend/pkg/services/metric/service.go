// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	ListPreDefinedMetrics(ctx core.Context) []QueryInfo
	ListQuerys(ctx core.Context) []Query
	QueryMetrics(ctx core.Context, req *QueryMetricsRequest) *QueryMetricsResult
}

type service struct {
	promRepo prometheus.Repo
}

func New(promRepo prometheus.Repo) Service {
	return &service{
		promRepo: promRepo,
	}
}
