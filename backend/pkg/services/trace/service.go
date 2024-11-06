package trace

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

var _ Service = (*service)(nil)

type Service interface {
	GetTraceFilters(startTime, endTime time.Time, needUpdate bool) (*response.GetTraceFiltersResponse, error)
	GetTraceFilterValues(startTime, endTime time.Time, searchText string, filter request.SpanTraceFilter) (*response.GetTraceFilterValueResponse, error)
	GetTracePageList(req *request.GetTracePageListRequest) (*response.GetTracePageListResponse, error)
	GetOnOffCPU(req *request.GetOnOffCPURequest) (*response.GetOnOffCPUResponse, error)
}

type service struct {
	chRepo clickhouse.Repo
}

func New(chRepo clickhouse.Repo) Service {
	return &service{
		chRepo: chRepo,
	}
}
