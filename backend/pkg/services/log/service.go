package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	// 获取故障现场分页日志
	GetFaultLogPageList(req *request.GetFaultLogPageListRequest) (*response.GetFaultLogPageListResponse, error)

	GetFaultLogContent(req *request.GetFaultLogContentRequest) (*response.GetFaultLogContentResponse, error)

	CreateLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error)

	DropLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error)

	UpdateLogTable(req *request.LogTableRequest) (*response.LogTableResponse, error)

	GetLogTableInfo(req *request.LogTableInfoRequest) (*response.LogTableInfoResponse, error)

	// 查询全量日志
	QueryLog(req *request.LogQueryRequest) (*response.LogQueryResponse, error)

	QueryLogContext(req *request.LogQueryContextRequest) (*response.LogQueryContextResponse, error)
	// 日志趋势图
	GetLogChart(req *request.LogQueryRequest) (*response.LogChartResponse, error)
	// 字段分析
	GetLogIndex(req *request.LogIndexRequest) (*response.LogIndexResponse, error)

	GetServiceRoute(req *request.GetServiceRouteRequest) (*response.GetServiceRouteResponse, error)

	GetLogParseRule(req *request.QueryLogParseRequest) (*response.LogParseResponse, error)

	UpdateLogParseRule(req *request.UpdateLogParseRequest) (*response.LogParseResponse, error)

	AddLogParseRule(req *request.AddLogParseRequest) (*response.LogParseResponse, error)

	DeleteLogParseRule(req *request.DeleteLogParseRequest) (*response.LogParseResponse, error)

	OtherTable(req *request.OtherTableRequest) (*response.OtherTableResponse, error)

	OtherTableInfo(req *request.OtherTableInfoRequest) (*response.OtherTableInfoResponse, error)

	AddOtherTable(req *request.AddOtherTableRequest) (*response.AddOtherTableResponse, error)

	DeleteOtherTable(req *request.DeleteOtherTableRequest) (*response.DeleteOtherTableResponse, error)
}

type service struct {
	chRepo   clickhouse.Repo
	dbRepo   database.Repo
	k8sApi   kubernetes.Repo
	promRepo prometheus.Repo
}

func New(chRepo clickhouse.Repo, dbRepo database.Repo, k8sApi kubernetes.Repo, promRepo prometheus.Repo) Service {
	return &service{
		chRepo:   chRepo,
		dbRepo:   dbRepo,
		k8sApi:   k8sApi,
		promRepo: promRepo,
	}
}
