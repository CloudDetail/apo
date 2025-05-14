// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package network

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

var _ Service = (*service)(nil)

type Service interface {
	GetPodMap(ctx core.Context, req *request.PodMapRequest) (*response.PodMapResponse, error)
	GetSpanSegmentsMetrics(ctx core.Context, req *request.SpanSegmentMetricsRequest) (response.SpanSegmentMetricsResponse, error)
}

type service struct {
	chRepo clickhouse.Repo
}

func New(chRepo clickhouse.Repo) Service {
	return &service{
		chRepo: chRepo,
	}
}
