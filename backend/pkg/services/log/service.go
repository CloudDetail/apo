package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

var _ Service = (*service)(nil)

type Service interface {
	// 获取故障现场分页日志
	GetFaultLogPageList(req *request.GetFaultLogPageListRequest) (*response.GetFaultLogPageListResponse, error)

	GetFaultLogContent(req *request.GetFaultLogContentRequest) (*response.GetFaultLogContentResponse, error)

	CreateLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error)

	DropLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error)

	UpdateLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error)

	GetLogTableInfo(req *request.LogTableRequest) (*response.LogTableResponse, error)
}

type service struct {
	chRepo clickhouse.Repo
	dbRepo database.Repo
}

func New(chRepo clickhouse.Repo, dbRepo database.Repo) Service {
	return &service{
		chRepo: chRepo,
		dbRepo: dbRepo,
	}
}
