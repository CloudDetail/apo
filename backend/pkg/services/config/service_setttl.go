package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"log"
	"regexp"
	"strconv"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

var typeRules = map[string][]string{
	"logs":     {"ilogtail_logs"},
	"trace":    {"span_trace", "jaeger_index_local", "jaeger_spans_archive_local", "jaeger_spans_local"},
	"k8s":      {"k8s_events"},
	"topology": {"service_relation", "service_topology"},
	"other": {"agent_log", "alert_event", "error_propagation", "error_report", "jvm_gc", "onoff_metric", "onstack_profiling",
		"profiling_event", "report_metric", "slo_record", "slow_report"},
}

var clusterTypeRules = map[string][]string{
	"logs":     {"ilogtail_logs_local"},
	"trace":    {"span_trace_local", "jaeger_index_local", "jaeger_spans_archive_local", "jaeger_spans_local"},
	"k8s":      {"k8s_events_local"},
	"topology": {"service_relation_local", "service_topology_local"},
	"other": {"agent_log_local", "alert_event_local", "error_propagation_local", "error_report_local", "jvm_gc_local",
		"onoff_metric_local", "onstack_profiling_local",
		"profiling_event_local", "report_metric_local", "slo_record_local", "slow_report_local"},
}

func getAllTables() []string {
	var tables []string
	if len(clickhouse.GetCluster()) > 0 {
		for _, table := range clusterTypeRules {
			newTable := make([]string, len(table))
			copy(newTable, table)
			tables = append(tables, newTable...)
		}
	} else {
		for _, table := range typeRules {
			newTable := make([]string, len(table))
			copy(newTable, table)
			tables = append(tables, newTable...)
		}
	}
	return tables
}

// 包级别的正则表达式变量
var ttlRegex = regexp.MustCompile(`TTL\s+(\S+(?:\s*\+\s*toIntervalDay\((\d+)\))?)`)
var toIntervalDayRegex = regexp.MustCompile(`toIntervalDay\((\d+)\)`)

func prepareTTLInfo(tables []model.TablesQuery) []model.ModifyTableTTLMap {
	mapResult := []model.ModifyTableTTLMap{}
	for _, t := range tables {
		matches := ttlRegex.FindStringSubmatch(t.CreateTableQuery)
		originalTTLExpression := ""
		var originalDays *int
		if len(matches) >= 2 {
			originalTTLExpression = matches[1]
			if len(matches) >= 3 && matches[2] != "" {
				days, err := strconv.Atoi(matches[2])
				if err == nil {
					originalDays = &days
				}
			}
		}
		item := model.ModifyTableTTLMap{
			Name:          t.Name,
			TTLExpression: originalTTLExpression,
			OriginalDays:  originalDays,
		}
		mapResult = append(mapResult, item)
	}
	return mapResult
}

func (s *service) SetTableTTL(tableNames []string, day int) error {
	tables, err := s.chRepo.GetTables(tableNames)
	if err != nil {
		log.Println("[SetSingleTableTTL] Error getting tables: ", err)
		return err
	}
	mapResult, err := convertModifyTableTTLMap(tables, day)
	if err != nil {
		log.Println("[SetSingleTableTTL] Error convertModifyTableTTLMap: ", err)
		return err
	}
	err = s.chRepo.ModifyTableTTL(context.Background(), mapResult)
	if err != nil {
		log.Println("[SetSingleTableTTL] Error ModifyTableTTL: ", err)
		return err
	}
	return nil
}

func convertModifyTableTTLMap(tables []model.TablesQuery, day int) ([]model.ModifyTableTTLMap, error) {

	mapResult := prepareTTLInfo(tables)
	for i := range mapResult {
		newInterval := fmt.Sprintf("toIntervalDay(%d)", day)
		mapResult[i].TTLExpression = toIntervalDayRegex.ReplaceAllString(mapResult[i].TTLExpression, newInterval)
	}

	return mapResult, nil
}
func (s *service) SetTTL(req *request.SetTTLRequest) error {
	if req.Day <= 0 {
		return errors.New("[SetTTL] Error : day should > 0  ")
	}

	tableNames := make([]string, len(typeRules[req.DataType]))
	copy(tableNames, typeRules[req.DataType])

	err := s.SetTableTTL(tableNames, req.Day)
	return err
}

func (s *service) SetSingleTableTTL(req *request.SetSingleTTLRequest) error {
	if req.Day <= 0 {
		return errors.New("[SetSingleTableTTL] Error : day should > 0  ")
	}
	err := s.SetTableTTL([]string{req.Name}, req.Day)
	return err
}
