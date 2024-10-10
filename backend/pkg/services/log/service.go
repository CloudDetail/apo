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

	// 查询全量日志
	QueryLog(req *request.LogQueryRequest) (*response.LogQueryResponse, error)
	// 日志趋势图
	GetLogChart(req *request.LogQueryRequest) (*response.LogChartResponse, error)
	// 字段分析
	GetLogIndex(req *request.LogIndexRequest) (*response.LogIndexResponse, error)
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
