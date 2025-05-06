// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse/integration"
)

type Repo interface {
	// ========== service_relationship ==========
	// Query the list of upstream nodes
	ListParentNodes(req *request.GetServiceEndpointTopologyRequest) (*model.TopologyNodes, error)
	// Query the list of downstream nodes
	ListChildNodes(req *request.GetServiceEndpointTopologyRequest) (*model.TopologyNodes, error)
	// Query the list of all descendant service nodes
	ListDescendantNodes(req *request.GetDescendantMetricsRequest) (*model.TopologyNodes, error)
	// Query the calling relationship of all descendant nodes
	ListDescendantRelations(req *request.GetServiceEndpointTopologyRequest) ([]*model.ToplogyRelation, error)
	// Query the entry node list
	ListEntryEndpoints(req *request.GetServiceEntryEndpointsRequest) ([]EntryNode, error)

	// ========== error_propagation ==========
	// Query instance-related error propagation chain
	ListErrorPropagation(req *request.GetErrorInstanceRequest) ([]ErrorInstancePropagation, error)

	// ========== span_trace ==========
	GetAvailableFilterKey(startTime, endTime time.Time, needUpdate bool) ([]request.SpanTraceFilter, error)
	UpdateFilterKey(startTime, endTime time.Time) ([]request.SpanTraceFilter, error)
	GetFieldValues(searchText string, filter *request.SpanTraceFilter, startTime, endTime time.Time) (*SpanTraceOptions, error)
	// Paging query the fault site log list
	GetFaultLogPageList(query *FaultLogQuery) ([]FaultLogResult, int64, error)
	// Paged Trace List
	GetTracePageList(req *request.GetTracePageListRequest) ([]QueryTraceResult, int64, error)

	// InfrastructureAlert(startTime time.Time, endTime time.Time, nodeNames []string) (*model.AlertEvent, bool, error)
	// NetworkAlert(startTime time.Time, endTime time.Time, pods []string, nodeNames []string, pids []string) (bool, error)

	CountK8sEvents(startTime int64, endTim int64, pods []string) ([]K8sEventsCount, error)

	// ========== application_logs ==========
	// Query the log content of the fault site. The sourceFrom can be blank. The log in the first source that can be found will be returned.
	QueryApplicationLogs(req *request.GetFaultLogContentRequest) (logContents *Logs, availableSource []string, err error)
	// Query the available source of the fault field log content
	QueryApplicationLogsAvailableSource(faultLog FaultLogResult) ([]string, error)

	CreateLogTable(req *request.LogTableRequest) ([]string, error)
	DropLogTable(req *request.LogTableRequest) ([]string, error)
	UpdateLogTable(req *request.LogTableRequest, old []request.Field) ([]string, error)

	queryRowsData(sql string) ([]map[string]any, error)

	QueryAllLogs(req *request.LogQueryRequest) ([]map[string]any, string, error)
	QueryLogContext(req *request.LogQueryContextRequest) ([]map[string]any, []map[string]any, error)
	GetLogChart(req *request.LogQueryRequest) ([]map[string]any, int64, error)
	GetLogIndex(req *request.LogIndexRequest) (map[string]uint64, uint64, error)

	OtherLogTable() ([]map[string]any, error)
	OtherLogTableInfo(req *request.OtherTableInfoRequest) ([]map[string]any, error)

	InsertBatchAlertEvents(ctx context.Context, events []*model.AlertEvent) error
	ReadAlertEvent(ctx context.Context, id uuid.UUID) (*model.AlertEvent, error)
	GetConn() driver.Conn

	//config
	ModifyTableTTL(ctx context.Context, mapResult []model.ModifyTableTTLMap) error
	GetTables(tables []model.Table) ([]model.TablesQuery, error)

	// ========== alert =================
	GetAlertsWithEventCount(startTime, endTime time.Time, filter *alert.AlertEventFilter, maxSize int) ([]alert.AlertWithEventCount, uint64, error)

	// ========== alert events ==========
	// Query the alarm events sampled by group and level, and sampleCount the number of samples for each group and level.
	GetAlertEventCountGroupByInstance(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]model.AlertEventCount, error)
	// Query alarm events sampled by labels, sampleCount the number of samples for each label.
	GetAlertEventsSample(sampleCount int, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]AlertEventSample, error)
	// Query alarm events by pageParam
	GetAlertEvents(startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances, pageParam *request.PageParam) ([]alert.AlertEvent, uint64, error)
	// ========== k8s events ============
	// SeverityNumber > 9 (warning)
	GetK8sAlertEventsSample(startTime time.Time, endTime time.Time, instances []*model.ServiceInstance) ([]K8sEvents, error)

	// profiling_event
	// GetOnOffCPU get span execution consumption
	GetOnOffCPU(pid uint32, nodeName string, startTime, endTime int64) (*[]ProfilingEvent, error)

	// ========== network (deepflow) ==========
	GetNetworkSpanSegments(traceId string, spanId string) ([]NetSegments, error)

	// ========== flame graph ===========
	GetFlameGraphData(startTime, endTime int64, nodeName string, pid, tid int64, sampleType, spanId, traceId string) (*[]FlameGraphData, error)

	AddWorkflowRecord(ctx context.Context, record *model.WorkflowRecord) error
	AddWorkflowRecords(ctx context.Context, records []model.WorkflowRecord) error
	GetAlertEventWithWorkflowRecord(req *request.AlertEventSearchRequest, cacheMinutes int) ([]alert.AEventWithWRecord, int64, error)
	GetAlertEventCounts(req *request.AlertEventSearchRequest, cacheMinutes int) (map[string]int64, error)

	GetAlertDetail(req *request.GetAlertDetailRequest, cacheMinutes int) (*alert.AEventWithWRecord, error)
	GetRelatedAlertEvents(req *request.GetAlertDetailRequest, cacheMinutes int) ([]alert.AEventWithWRecord, int64, error)

	CreateAlertNotifyRecord(ctx context.Context, record model.AlertNotifyRecord) error

	integration.Input
}

type chRepo struct {
	conn     driver.Conn
	database string
	availableFilters
	db *sql.DB

	integration.Input
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

	dsn := fmt.Sprintf("clickhouse://%s:%s@%s/%s", username, url.QueryEscape(password), address[0], database)
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}
	var repo *chRepo
	// Use the wrapped Conn at the Debug log level, and output the time taken to execute SQL.
	if logger.Level() == zap.DebugLevel {
		repo = &chRepo{
			database: database,
			conn: &WrappedConn{
				Conn:   conn,
				logger: logger,
			},
			db: db,
		}
	} else {
		repo = &chRepo{
			database: database,
			conn:     conn,
			db:       db,
		}
	}

	now := time.Now()
	filters, err := repo.UpdateFilterKey(now.Add(-48*time.Hour), now)
	if err == nil {
		repo.SetAvailableFilters(filters, now)
	}

	repo.Input, err = integration.NewInputRepo(repo.conn, repo.database)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (ch *chRepo) GetConn() driver.Conn {
	return ch.conn
}
