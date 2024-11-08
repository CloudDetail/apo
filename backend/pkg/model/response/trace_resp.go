package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

type GetTracePageListResponse struct {
	List       []clickhouse.QueryTraceResult `json:"list"`
	Pagination *model.Pagination             `json:"pagination"`
}

type GetOnOffCPUResponse struct {
	ProfilingEvent *[]clickhouse.ProfilingEvent `json:"profilingEvent"`
}

type GetSingleTraceInfoResponse string
