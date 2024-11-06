package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

type GetTraceFilterValueResponse struct {
	TraceFilterOptions clickhouse.SpanTraceOptions `json:"traceFilterOptions"`
}

type GetTraceFiltersResponse struct {
	TraceFilters []request.SpanTraceFilter `json:"traceFilters"`
}
