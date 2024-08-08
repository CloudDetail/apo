package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

var _ Service = (*service)(nil)

type Service interface {
	// 获取故障现场分页日志
	GetFaultLogPageList(req *request.GetFaultLogPageListRequest) (*response.GetFaultLogPageListResponse, error)

	GetFaultLogContent(req *request.GetFaultLogContentRequest) (*response.GetFaultLogContentResponse, error)
}

type service struct {
	chRepo clickhouse.Repo
}

func New(chRepo clickhouse.Repo) Service {
	return &service{
		chRepo: chRepo,
	}
}
