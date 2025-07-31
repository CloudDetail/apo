// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	ListPreDefinedMetrics(ctx core.Context) []QueryInfo
	ListQuerys(ctx core.Context) []Query
	QueryMetrics(ctx core.Context, req *QueryMetricsRequest) *QueryMetricsResult
	QueryPods(ctx core.Context, req *request.QueryPodsRequest) (*response.QueryPodsResponse, error)
}

type service struct {
	promRepo prometheus.Repo
}

func New(promRepo prometheus.Repo) Service {
	return &service{
		promRepo: promRepo,
	}
}
