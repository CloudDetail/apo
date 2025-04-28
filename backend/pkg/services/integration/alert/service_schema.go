// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/enrich"
)

func (s *service) CreateSchema(req *alert.CreateSchemaRequest) error {
	err := s.dbRepo.CreateSchema(req.Schema, req.Columns)
	if err != nil {
		return err
	}
	if len(req.FullRows) > 0 {
		return s.dbRepo.InsertSchemaData(req.Schema, req.Columns, req.FullRows)
	}
	return nil
}

func (s *service) ListSchema() ([]string, error) {
	return s.dbRepo.ListSchema()
}

func (s *service) GetSchemaData(schema string) ([]string, map[int64][]string, error) {
	return s.dbRepo.GetSchemaData(schema)
}

func (s *service) CheckSchemaIsUsed(schema string) ([]string, error) {
	return s.dbRepo.CheckSchemaIsUsed(schema)
}

func (s *service) DeleteSchema(ctx core.Context, schema string) error {
	s.dispatcher.EnricherMap.Range(func(key, value any) bool {
		enricher := value.(*enrich.AlertEnricher)
		enricher.RemoveRuleByDeletedSchema(schema)
		return true
	})

	return s.dbRepo.DeleteSchema(ctx, schema)
}

func (s *service) ListSchemaColumns(schema string) ([]string, error) {
	return s.dbRepo.ListSchemaColumns(schema)
}

func (s *service) UpdateSchemaData(req *alert.UpdateSchemaDataRequest) error {
	if req.ClearAll {
		err := s.dbRepo.ClearSchemaData(req.Schema)
		if err != nil {
			return err
		}
	}
	if len(req.NewRows) > 0 {
		return s.dbRepo.InsertSchemaData(req.Schema, req.Columns, req.NewRows)
	} else if len(req.UpdateRows) > 0 {
		return s.dbRepo.UpdateSchemaData(req.Schema, req.Columns, req.UpdateRows)
	}

	return nil
}
