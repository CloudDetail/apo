// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (repo *subRepo) LoadResolvedIncidents(ctx core.Context) ([]alert.Incident, error) {
	var incidents []alert.Incident
	err := repo.GetContextDB(ctx).Model(&alert.Incident{}).
		Where("status = ?", alert.StatusFiring).
		Find(&incidents).Error

	return incidents, err
}

type IncidentTempConditionWithTemp struct {
	alert.IncidentKeyTemp
	alert.IncidentCondition
}

func (repo *subRepo) CreateIncident(ctx core.Context, incident *alert.Incident) error {
	return repo.GetContextDB(ctx).Create(incident).Error
}

func (repo *subRepo) UpdateIncident(ctx core.Context, incident *alert.Incident) error {
	return repo.GetContextDB(ctx).Updates(incident).Error
}

func (repo *subRepo) LoadIncidentTemplates(ctx core.Context) ([]alert.IncidentKeyTemp, error) {
	var res []IncidentTempConditionWithTemp

	err := repo.GetContextDB(ctx).Table("incident_key_temp as ikt").
		Joins("LEFT JOIN incident_condition as ic ON ikt.id = ic.incident_temp_id").
		Find(&res).Error

	if err != nil {
		return nil, err
	}

	var incidentKeyTemps []alert.IncidentKeyTemp
	var incidentKeyMap = make(map[string]int)
	for _, cond := range res {
		idx, find := incidentKeyMap[cond.IncidentTempID]
		if !find {
			incidentKeyMap[cond.IncidentTempID] = len(incidentKeyTemps)

			incidentKeyTemp := cond.IncidentKeyTemp
			incidentKeyTemp.Conditions = append(incidentKeyTemp.Conditions, cond.IncidentCondition)

			incidentKeyTemps = append(incidentKeyTemps, incidentKeyTemp)
			continue
		}

		incidentKeyTemps[idx].Conditions = append(incidentKeyTemps[idx].Conditions, cond.IncidentCondition)
	}
	return incidentKeyTemps, nil
}

func (repo *subRepo) GetIncidentTemplatesBySourceId(ctx core.Context, sourceId string) ([]*alert.IncidentKeyTemp, error) {
	var res []IncidentTempConditionWithTemp

	err := repo.GetContextDB(ctx).Table("incident_key_temp as ikt").
		Joins("LEFT JOIN incident_condition as ic ON ikt.id = ic.incident_temp_id").
		Where("ikt.alert_source_id = ?", sourceId).
		Find(&res).Error

	if err != nil {
		return nil, err
	}

	var incidentKeyTemps []*alert.IncidentKeyTemp
	var incidentKeyMap = make(map[string]int)
	for _, cond := range res {
		idx, find := incidentKeyMap[cond.IncidentTempID]
		if !find {
			incidentKeyMap[cond.IncidentTempID] = len(incidentKeyTemps)

			incidentKeyTemp := cond.IncidentKeyTemp
			incidentKeyTemp.Conditions = append(incidentKeyTemp.Conditions, cond.IncidentCondition)

			incidentKeyTemps = append(incidentKeyTemps, &incidentKeyTemp)
			continue
		}

		incidentKeyTemps[idx].Conditions = append(incidentKeyTemps[idx].Conditions, cond.IncidentCondition)
	}
	return incidentKeyTemps, nil
}

func (repo *subRepo) CreateIncidentTemplates(ctx core.Context, temps []*alert.IncidentKeyTemp) error {
	err := repo.GetContextDB(ctx).Create(temps).Error
	if err != nil {
		return err
	}

	conditions := []alert.IncidentCondition{}
	for _, temp := range temps {
		conditions = append(conditions, temp.Conditions...)
	}

	return repo.GetContextDB(ctx).Create(conditions).Error
}

func (repo *subRepo) DeleteIncidentTemplates(ctx core.Context, tempIDs []string) error {
	err := repo.GetContextDB(ctx).Where("incident_temp_id IN ?", tempIDs).Delete(&alert.IncidentCondition{}).Error
	if err != nil {
		return err
	}
	return repo.GetContextDB(ctx).Where("id IN ?", tempIDs).Delete(&alert.IncidentKeyTemp{}).Error
}

func (repo *subRepo) UpdateIncidentTemplates(ctx core.Context, temps []*alert.IncidentKeyTemp) error {
	tempIDs := make([]string, 0, len(temps))
	for _, temp := range temps {
		tempIDs = append(tempIDs, temp.ID)
	}

	err := repo.DeleteIncidentTemplates(ctx, tempIDs)
	if err != nil {
		return err
	}

	return repo.CreateIncidentTemplates(ctx, temps)
}
