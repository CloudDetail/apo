package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

type Repo interface {
	// ========== service_relation ==========
	// 查询上游节点列表
	ListParentNodes(req *request.GetServiceEndpointTopologyRequest) ([]TopologyNode, error)
	// 查询下游节点列表
	ListChildNodes(req *request.GetServiceEndpointTopologyRequest) ([]TopologyNode, error)

	// ========== service_topology ==========
	// 查询所有子孙节点列表
	ListDescendantNodes(req *request.GetDescendantMetricsRequest) ([]TopologyNode, error)
	// 查询所有子孙节点的调用关系
	ListDescendantRelations(req *request.GetServiceEndpointTopologyRequest) ([]ToplogyRelation, error)
	// 查询入口节点列表
	ListEntryEndpoints(req *request.GetServiceEntryEndpointsRequest) ([]EntryNode, error)

	// ========== error_propagation ==========
	// 查询实例相关的错误传播链
	ListErrorPropagation(req *request.GetErrorInstanceRequest) ([]ErrorInstancePropagation, error)

	// ========== span_trace ==========
	GetAvailableFilterKey(startTime, endTime time.Time, needUpdate bool) ([]request.SpanTraceFilter, error)
	UpdateFilterKey(startTime, endTime time.Time) ([]request.SpanTraceFilter, error)
	GetFieldValues(searchText string, filter *request.SpanTraceFilter, startTime, endTime time.Time) (*SpanTraceOptions, error)
	// 分页查询故障现场日志列表
	GetFaultLogPageList(query *FaultLogQuery) ([]FaultLogResult, int64, error)
	// 分页Trace列表
	GetTracePageList(req *request.GetTracePageListRequest) ([]QueryTraceResult, int64, error)

	// InfrastructureAlert(startTime time.Time, endTime time.Time, nodeNames []string) (*model.AlertEvent, bool, error)
	// NetworkAlert(startTime time.Time, endTime time.Time, pods []string, nodeNames []string, pids []string) (bool, error)

	CountK8sEvents(startTime int64, endTim int64, pods []string) ([]K8sEventsCount, error)

	// ========== application_logs ==========
	// 查询故障现场日志内容, sourceFrom 可为空, 将返回可查到的第一个来源中的日志
	QueryApplicationLogs(req *request.GetFaultLogContentRequest) (logContents *Logs, availableSource []string, err error)
	// 查询故障现场日志内容可用的来源
	QueryApplicationLogsAvailableSource(faultLog FaultLogResult) ([]string, error)

	InsertBatchAlertEvents(ctx context.Context, events []*model.AlertEvent) error
	ReadAlertEvent(ctx context.Context, id uuid.UUID) (*model.AlertEvent, error)
	GetConn() driver.Conn

	//config
	ModifyTableTTL(ctx context.Context, mapResult []model.ModifyTableTTLMap) error
	GetTables(tables []string) ([]model.TablesQuery, error)

	// ========== alert events ==========
	// 查询按group和级别采样的告警事件,sampleCount为每个group和级别采样的数量
	GetAlertEventCountGroupByInstance(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance) ([]model.AlertEventCount, error)
	// 查询按labels采样的告警事件,sampleCount为每个labels采样的数量
	GetAlertEventsSample(sampleCount int, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance) ([]AlertEventSample, error)
	// 查询按pageParam分页的告警事件
	GetAlertEvents(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances []*model.ServiceInstance, pageParam *request.PageParam) ([]PagedAlertEvent, int, error)

	// ========== k8s events ============
	// SeverityNumber > 9 (warning)
	GetK8sAlertEventsSample(startTime time.Time, endTime time.Time, instances []*model.ServiceInstance) ([]K8sEvents, error)
}

type chRepo struct {
	conn     driver.Conn
	database string
	AvailableFilters
}

type AvailableFilters struct {
	Filters          []request.SpanTraceFilter
	FilterUpdateTime time.Time
}

func New(logger *zap.Logger, address []string, database string, username string, password string) (Repo, error) {
	settings := clickhouse.Settings{}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr:     address,
		Settings: settings,
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		DialTimeout: time.Duration(5) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to clickhouse: %s", err)
	}

	var repo *chRepo
	// Debug 日志等级时使用包装的Conn，输出执行SQL的耗时
	if logger.Level() == zap.DebugLevel {
		repo = &chRepo{
			database: database,
			conn: &WrappedConn{
				Conn:   conn,
				logger: logger,
			},
		}
	} else {
		repo = &chRepo{
			database: database,
			conn:     conn,
		}
	}

	now := time.Now()
	filters, err := repo.UpdateFilterKey(now.Add(-48*time.Hour), now)
	if err == nil {
		repo.Filters = filters
		repo.FilterUpdateTime = now
	}

	return repo, nil
}

func (ch *chRepo) GetConn() driver.Conn {
	return ch.conn
}
