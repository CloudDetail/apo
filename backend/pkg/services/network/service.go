package network

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

var _ Service = (*service)(nil)

type Service interface {
	GetPodMap(req *request.PodMapRequest) (*response.PodMapResponse, error)
	GetSpanSegmentsMetrics(req *request.SpanSegmentMetricsRequest) (*response.SpanSegmentMetricsResponse, error)
}

type service struct {
	chRepo clickhouse.Repo
}

func New(chRepo clickhouse.Repo) Service {
	return &service{
		chRepo: chRepo,
	}
}
