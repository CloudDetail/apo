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

// 包级别的正则表达式变量
var ttlRegex = regexp.MustCompile(`TTL\s+((?:[^\s()]+\s*)+\(\s*[^)]+\s*\)\s*\+\s*toIntervalDay\((\d+)\))`)
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

func (s *service) SetTableTTL(tableNames []model.Table, day int) error {
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
		log.Printf("TTL statement: %s", mapResult[i].TTLExpression)
		mapResult[i].TTLExpression = toIntervalDayRegex.ReplaceAllString(mapResult[i].TTLExpression, newInterval)
	}

	return mapResult, nil
}
func (s *service) SetTTL(req *request.SetTTLRequest) error {
	if req.Day <= 0 {
		return errors.New("[SetTTL] Error : day should > 0  ")
	}

	tables := model.GetTables(req.DataType)
	if len(tables) == 0 {
		return fmt.Errorf("type: %s does not have tables", req.DataType)
	}
	err := s.SetTableTTL(tables, req.Day)
	return err
}

func (s *service) SetSingleTableTTL(req *request.SetSingleTTLRequest) error {
	if req.Day <= 0 {
		return errors.New("[SetSingleTableTTL] Error: day should > 0  ")
	}
	if !model.IsTableExists(req.Name) {
		return fmt.Errorf("[SetSingleTableTTL] Error: table %s does not exists", req.Name)
	}

	tables := []model.Table{
		{Name: req.Name},
	}
	err := s.SetTableTTL(tables, req.Day)
	return err
}
