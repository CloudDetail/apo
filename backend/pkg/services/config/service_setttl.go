// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

// Regular expression variables at the package level
var ttlRegex = regexp.MustCompile(`TTL\s+(toDateTime\(([^)]+)\)|([^+\s]+))\s*\+\s*toIntervalDay\((\d+)\)`)
var toIntervalDayRegex = regexp.MustCompile(`toIntervalDay\((\d+)\)`)

func prepareTTLInfo(tables []model.TablesQuery) []model.ModifyTableTTLMap {
	mapResult := []model.ModifyTableTTLMap{}
	for _, t := range tables {
		matches := ttlRegex.FindStringSubmatch(t.CreateTableQuery)
		originalTTLExpression := ""
		var originalDays *int
		if len(matches) >= 1 {
			originalTTLExpression = matches[0]

			if len(matches) >= 4 {
				days, err := strconv.Atoi(matches[4])
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

func (s *service) SetTableTTL(ctx core.Context, tableNames []model.Table, day int) error {
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

func (s *service) SetTTL(ctx core.Context, req *request.SetTTLRequest) error {
	if req.Day <= 0 {
		return errors.New("[SetTTL] Error : day should > 0  ")
	}

	tables := model.GetTables(req.DataType)
	if len(tables) == 0 {
		return fmt.Errorf("type: %s does not have tables", req.DataType)
	}
	err := s.SetTableTTL(ctx, tables, req.Day)
	return err
}

func (s *service) SetSingleTableTTL(ctx core.Context, req *request.SetSingleTTLRequest) error {
	if req.Day <= 0 {
		return errors.New("[SetSingleTableTTL] Error: day should > 0  ")
	}
	if !model.IsTableExists(req.Name) {
		return fmt.Errorf("[SetSingleTableTTL] Error: table %s does not exists", req.Name)
	}

	tables := []model.Table{
		{Name: req.Name},
	}
	err := s.SetTableTTL(ctx, tables, req.Day)
	return err
}
