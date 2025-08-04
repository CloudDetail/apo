// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/incident"
)

func (s *service) GetIncidentTempBySource(ctx core.Context, req *request.GetIncidentTempBySourceRequest) (*response.GetIncidentTempBySourceResponse, error) {
	if req.NeedFresh {
		temps, err := s.dbRepo.GetIncidentTemplatesBySourceId(ctx, req.SourceID)
		if err != nil {
			return nil, err
		}
		incident.IncidentMemCacheInstance.UpdateIncidentTemp(req.SourceID, temps)
		return &response.GetIncidentTempBySourceResponse{
			ITemps: temps,
		}, nil
	}

	temps, _ := incident.IncidentMemCacheInstance.GetTemps(req.SourceID)
	return &response.GetIncidentTempBySourceResponse{
		ITemps: temps,
	}, nil
}

func (s *service) SetIncidentTempBySource(ctx core.Context, req *request.SetIncidentTempBySourceRequest) error {
	for _, temp := range req.ITemps {
		temp.AlertSourceID = req.SourceID
		if err := temp.Compile(); err != nil {
			return err
		}
	}

	toDeleted := []string{}
	toUpdated := []*alert.IncidentKeyTemp{}
	toCreated := []*alert.IncidentKeyTemp{}

	oldITemps, _ := incident.IncidentMemCacheInstance.GetTemps(req.SourceID)

	newITempMap := make(map[string]*alert.IncidentKeyTemp)
	for _, temp := range req.ITemps {
		newITempMap[temp.ID] = temp
	}

	oldITempMap := make(map[string]*alert.IncidentKeyTemp)
	for _, temp := range oldITemps {
		oldITempMap[temp.ID] = temp
	}

	for key := range oldITempMap {
		if _, exists := newITempMap[key]; !exists {
			toDeleted = append(toDeleted, key)
		}
	}

	for key, newTemp := range newITempMap {
		if oldTemp, exists := oldITempMap[key]; exists {
			if !oldTemp.Equal(newTemp) {
				toUpdated = append(toUpdated, newTemp)
			}
		} else {
			toCreated = append(toCreated, newTemp)
		}
	}

	var transactions = []func(txCtx core.Context) error{}

	if len(toDeleted) > 0 {
		transactions = append(transactions, func(txCtx core.Context) error {
			return s.dbRepo.DeleteIncidentTemplates(txCtx, toDeleted)
		})
	}

	if len(toUpdated) > 0 {
		transactions = append(transactions, func(txCtx core.Context) error {
			return s.dbRepo.UpdateIncidentTemplates(txCtx, toUpdated)
		})
	}

	if len(toCreated) > 0 {
		transactions = append(transactions, func(txCtx core.Context) error {
			return s.dbRepo.CreateIncidentTemplates(txCtx, toCreated)
		})
	}

	err := s.dbRepo.Transaction(ctx, transactions...)
	if err != nil {
		return err
	}

	incident.IncidentMemCacheInstance.UpdateIncidentTemp(req.SourceID, req.ITemps)
	return nil
}

func (s *service) ClearIncidentTempBySource(ctx core.Context, req *request.ClearIncidentTempBySourceRequest) error {
	temps, err := s.dbRepo.GetIncidentTemplatesBySourceId(ctx, req.SourceID)
	if err != nil {
		return err
	}

	tempIDs := make([]string, 0, len(temps))
	for _, temp := range temps {
		tempIDs = append(tempIDs, temp.ID)
	}

	return s.dbRepo.Transaction(ctx, func(txCtx core.Context) error {
		return s.dbRepo.DeleteIncidentTemplates(txCtx, tempIDs)
	})
}
