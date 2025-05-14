// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/enrich"
)

func (s *service) CreateSchema(ctx core.Context, req *alert.CreateSchemaRequest) error {
	err := s.dbRepo.CreateSchema(ctx, req.Schema, req.Columns)
	if err != nil {
		return err
	}
	if len(req.FullRows) > 0 {
		return s.dbRepo.InsertSchemaData(ctx, req.Schema, req.Columns, req.FullRows)
	}
	return nil
}

func (s *service) ListSchema(ctx core.Context) ([]string, error) {
	return s.dbRepo.ListSchema(ctx)
}

func (s *service) GetSchemaData(ctx core.Context, schema string) ([]string, map[int64][]string, error) {
	return s.dbRepo.GetSchemaData(ctx, schema)
}

func (s *service) CheckSchemaIsUsed(ctx core.Context, schema string) ([]string, error) {
	return s.dbRepo.CheckSchemaIsUsed(ctx, schema)
}

func (s *service) DeleteSchema(ctx core.Context, schema string) error {
	s.dispatcher.EnricherMap.Range(func(key, value any) bool {
		enricher := value.(*enrich.AlertEnricher)
		enricher.RemoveRuleByDeletedSchema(schema)
		return true
	})

	return s.dbRepo.DeleteSchema(ctx, schema)
}

func (s *service) ListSchemaColumns(ctx core.Context, schema string) ([]string, error) {
	return s.dbRepo.ListSchemaColumns(ctx, schema)
}

func (s *service) UpdateSchemaData(ctx core.Context, req *alert.UpdateSchemaDataRequest) error {
	if req.ClearAll {
		err := s.dbRepo.ClearSchemaData(ctx, req.Schema)
		if err != nil {
			return err
		}
	}
	if len(req.NewRows) > 0 {
		return s.dbRepo.InsertSchemaData(ctx, req.Schema, req.Columns, req.NewRows)
	} else if len(req.UpdateRows) > 0 {
		return s.dbRepo.UpdateSchemaData(ctx, req.Schema, req.Columns, req.UpdateRows)
	}

	return nil
}
