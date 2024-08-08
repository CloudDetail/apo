package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

type GetFaultLogPageListResponse struct {
	List       []clickhouse.FaultLogResult `json:"list"`
	Pagination *model.Pagination           `json:"pagination"`
}

type GetFaultLogContentResponse struct {
	Sources     []string         `json:"sources"`
	LogContents *clickhouse.Logs `json:"logContents"`
}
