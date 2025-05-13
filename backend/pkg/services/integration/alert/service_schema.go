// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/enrich"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) CreateSchema(ctx_core core.Context, req *alert.CreateSchemaRequest) error {
	err := s.dbRepo.CreateSchema(req.Schema, req.Columns)
	if err != nil {
		return err
	}
	if len(req.FullRows) > 0 {
		return s.dbRepo.InsertSchemaData(req.Schema, req.Columns, req.FullRows)
	}
	return nil
}

func (s *service) ListSchema(ctx_core core.Context,) ([]string, error) {
	return s.dbRepo.ListSchema()
}

func (s *service) GetSchemaData(ctx_core core.Context, schema string) ([]string, map[int64][]string, error) {
	return s.dbRepo.GetSchemaData(schema)
}

func (s *service) CheckSchemaIsUsed(ctx_core core.Context, schema string) ([]string, error) {
	return s.dbRepo.CheckSchemaIsUsed(schema)
}

func (s *service) DeleteSchema(ctx_core core.Context, schema string) error {
	s.dispatcher.EnricherMap.Range(func(key, value any) bool {
		enricher := value.(*enrich.AlertEnricher)
		enricher.RemoveRuleByDeletedSchema(schema)
		return true
	})

	return s.dbRepo.DeleteSchema(schema)
}

func (s *service) ListSchemaColumns(ctx_core core.Context, schema string) ([]string, error) {
	return s.dbRepo.ListSchemaColumns(schema)
}

func (s *service) UpdateSchemaData(ctx_core core.Context, req *alert.UpdateSchemaDataRequest) error {
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
