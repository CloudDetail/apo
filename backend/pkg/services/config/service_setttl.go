package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

var typeRules = map[string][]string{
	"logs":     {"ilogtail_logs"},
	"trace":    {"span_trace", "jaeger_index_local", "jaeger_spans_archive_local", "jaeger_spans_local", "jaeger_operations_local"},
	"k8s":      {"k8s_events"},
	"topology": {"service_relation", "service_topology"},
}

// 包级别的正则表达式变量
var ttlRegex = regexp.MustCompile(`TTL\s+([^\s]+(?:\s*\+\s*toIntervalDay\((\d+)\))?)`)
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
func (s *service) SetTableTTL(blackTableNames []string, whiteTableNames []string, day int) error {
	tables, err := s.chRepo.GetTables(blackTableNames, whiteTableNames)
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
	blackTableNames := []string{}
	whiteTableNames := []string{}
	if req.DataType == "other" {
		types := []string{"logs", "trace", "k8s", "topology"}
		for _, t := range types {
			convertedNames := typeRules[t]
			blackTableNames = append(blackTableNames, convertedNames...)
		}
	} else {
		whiteTableNames = typeRules[req.DataType]
	}
	err := s.SetTableTTL(blackTableNames, whiteTableNames, req.Day)
	return err
}

func (s *service) SetSingleTableTTL(req *request.SetSingleTTLRequest) error {
	if req.Day <= 0 {
		return errors.New("[SetSingleTableTTL] Error : day should > 0  ")
	}
	err := s.SetTableTTL([]string{}, []string{req.Name}, req.Day)
	return err
}
