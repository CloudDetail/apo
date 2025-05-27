// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse/factory"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse/integration"
)

type Repo interface {
	// ========== service_relationship ==========
	// Query the list of upstream nodes
	ListParentNodes(ctx core.Context, req *request.GetServiceEndpointTopologyRequest) (*model.TopologyNodes, error)
	// Query the list of downstream nodes
	ListChildNodes(ctx core.Context, req *request.GetServiceEndpointTopologyRequest) (*model.TopologyNodes, error)
	// Query the list of all descendant service nodes
	ListDescendantNodes(ctx core.Context, req *request.GetDescendantMetricsRequest) (*model.TopologyNodes, error)
	// Query the calling relationship of all descendant nodes
	ListDescendantRelations(ctx core.Context, req *request.GetServiceEndpointTopologyRequest) ([]*model.ToplogyRelation, error)
	// Query the entry node list
	ListEntryEndpoints(ctx core.Context, req *request.GetServiceEntryEndpointsRequest) ([]EntryNode, error)

	// ========== error_propagation ==========
	// Query instance-related error propagation chain
	ListErrorPropagation(ctx core.Context, req *request.GetErrorInstanceRequest) ([]ErrorInstancePropagation, error)

	// ========== span_trace ==========
	GetAvailableFilterKey(ctx core.Context, startTime, endTime time.Time, needUpdate bool) ([]request.SpanTraceFilter, error)
	UpdateFilterKey(ctx core.Context, startTime, endTime time.Time) ([]request.SpanTraceFilter, error)
	GetFieldValues(ctx core.Context, searchText string, filter *request.SpanTraceFilter, startTime, endTime time.Time) (*SpanTraceOptions, error)
	// Paging query the fault site log list
	GetFaultLogPageList(ctx core.Context, query *FaultLogQuery) ([]FaultLogResult, int64, error)
	// Paged Trace List
	GetTracePageList(ctx core.Context, req *request.GetTracePageListRequest) ([]QueryTraceResult, int64, error)

	// InfrastructureAlert(startTime time.Time, endTime time.Time, nodeNames []string) (*model.AlertEvent, bool, error)
	// NetworkAlert(startTime time.Time, endTime time.Time, pods []string, nodeNames []string, pids []string) (bool, error)

	CountK8sEvents(ctx core.Context, startTime int64, endTim int64, pods []string) ([]K8sEventsCount, error)

	// ========== application_logs ==========
	// Query the log content of the fault site. The sourceFrom can be blank. The log in the first source that can be found will be returned.
	QueryApplicationLogs(ctx core.Context, req *request.GetFaultLogContentRequest) (logContents *Logs, availableSource []string, err error)
	// Query the available source of the fault field log content
	QueryApplicationLogsAvailableSource(ctx core.Context, faultLog FaultLogResult) ([]string, error)

	CreateLogTable(ctx core.Context, req *request.LogTableRequest) ([]string, error)
	DropLogTable(ctx core.Context, req *request.LogTableRequest) ([]string, error)
	UpdateLogTable(ctx core.Context, req *request.LogTableRequest, old []request.Field) ([]string, error)

	queryRowsData(ctx core.Context, sql string) ([]map[string]any, error)

	QueryAllLogs(ctx core.Context, req *request.LogQueryRequest) ([]map[string]any, string, error)
	QueryLogContext(ctx core.Context, req *request.LogQueryContextRequest) ([]map[string]any, []map[string]any, error)
	GetLogChart(ctx core.Context, req *request.LogQueryRequest) ([]map[string]any, int64, error)
	GetLogIndex(ctx core.Context, req *request.LogIndexRequest) (map[string]uint64, uint64, error)

	OtherLogTable(ctx core.Context) ([]map[string]any, error)
	OtherLogTableInfo(ctx core.Context, req *request.OtherTableInfoRequest) ([]map[string]any, error)

	InsertBatchAlertEvents(ctx core.Context, events []*model.AlertEvent) error
	ReadAlertEvent(ctx core.Context, id uuid.UUID) (*model.AlertEvent, error)
	GetConn(ctx core.Context) driver.Conn

	//config
	ModifyTableTTL(ctx core.Context, mapResult []model.ModifyTableTTLMap) error
	GetTables(ctx core.Context, tables []model.Table) ([]model.TablesQuery, error)

	// ========== alert =================
	GetAlertsWithEventCount(ctx core.Context, startTime, endTime time.Time, filter *alert.AlertEventFilter, maxSize int) ([]alert.AlertWithEventCount, uint64, error)

	// ========== alert events ==========
	// Query the alarm events sampled by group and level, and sampleCount the number of samples for each group and level.
	GetAlertEventCountGroupByInstance(ctx core.Context, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]model.AlertEventCount, error)
	// Query alarm events sampled by labels, sampleCount the number of samples for each label.
	GetAlertEventsSample(ctx core.Context, sampleCount int, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances) ([]AlertEventSample, error)
	// Query alarm events by pageParam
	GetAlertEvents(ctx core.Context, startTime time.Time, endTime time.Time, filter request.AlertFilter, instances *model.RelatedInstances, pageParam *request.PageParam) ([]alert.AlertEvent, uint64, error)
	// ========== k8s events ============
	// SeverityNumber > 9 (warning)
	GetK8sAlertEventsSample(ctx core.Context, startTime time.Time, endTime time.Time, instances []*model.ServiceInstance) ([]K8sEvents, error)

	// profiling_event
	// GetOnOffCPU get span execution consumption
	GetOnOffCPU(ctx core.Context, pid uint32, nodeName string, startTime, endTime int64) (*[]ProfilingEvent, error)

	// ========== network (deepflow) ==========
	GetNetworkSpanSegments(ctx core.Context, traceId string, spanId string) ([]NetSegments, error)

	// ========== flame graph ===========
	GetFlameGraphData(ctx core.Context, startTime, endTime int64, nodeName string, pid, tid int64, sampleType, spanId, traceId string) (*[]FlameGraphData, error)

	AddWorkflowRecord(ctx core.Context, record *model.WorkflowRecord) error
	AddWorkflowRecords(ctx core.Context, records []model.WorkflowRecord) error
	GetAlertEventWithWorkflowRecord(ctx core.Context, req *request.AlertEventSearchRequest, cacheMinutes int) ([]alert.AEventWithWRecord, int64, error)
	GetAlertEventCounts(ctx core.Context, req *request.AlertEventSearchRequest, cacheMinutes int) (map[string]int64, error)

	GetAlertDetail(ctx core.Context, req *request.GetAlertDetailRequest, cacheMinutes int) (*alert.AEventWithWRecord, error)
	GetRelatedAlertEvents(ctx core.Context, req *request.GetAlertDetailRequest, cacheMinutes int) ([]alert.AEventWithWRecord, int64, error)

	CreateAlertNotifyRecord(ctx core.Context, record model.AlertNotifyRecord) error
	GetLatestAlertEventByAlertID(ctx core.Context, alertID string) (*alert.AlertEvent, error)

	ManualResolveLatestAlertEventByAlertID(ctx core.Context, alertID string) error

	integration.Input
}

type chRepo struct {
	*factory.Conn

	database string
	availableFilters

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

	var repo *chRepo
	// Use the wrapped Conn at the Debug log level, and output the time taken to execute SQL.
	if logger.Level() == zap.DebugLevel {
		repo = &chRepo{
			database: database,
			Conn: &factory.Conn{Conn: &WrappedConn{
				Conn:   conn,
				logger: logger,
			}},
		}
	} else {
		repo = &chRepo{
			database: database,
			Conn:     &factory.Conn{Conn: conn},
		}
	}

	now := time.Now()
	filters, err := repo.UpdateFilterKey(core.EmptyCtx(), now.Add(-48*time.Hour), now)
	if err == nil {
		repo.SetAvailableFilters(filters, now)
	}

	repo.Input, err = integration.NewInputRepo(repo.Conn, repo.database)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (ch *chRepo) GetConn(ctx core.Context) driver.Conn {
	return ch.GetContextDB(ctx)
}
