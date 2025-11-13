// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"slices"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	mi "github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (repo *daoRepo) ListDataPlaneType(ctx core.Context) ([]mi.DataPlaneType, error) {
	var res []mi.DataPlaneType
	err := repo.GetContextDB(ctx).Find(&res).Order("type_name asc").Error
	return res, err
}

func (repo *daoRepo) GetDataPlaneType(ctx core.Context, typeName string) (*mi.DataPlaneType, error) {
	var res mi.DataPlaneType
	err := repo.GetContextDB(ctx).Where("type_name = ?", typeName).First(&res).Error
	return &res, err
}

func (repo *daoRepo) CreateDataPlane(ctx core.Context, d *mi.DataPlane) error {
	return repo.GetContextDB(ctx).Create(d).Error
}

func (repo *daoRepo) CreateCluster2DataPlane(ctx core.Context, c2dp *mi.Cluster2DataPlane) error {
	return repo.GetContextDB(ctx).Create(c2dp).Error
}

func (repo *daoRepo) ListDataPlane(ctx core.Context) ([]mi.DataPlaneWithClusterIDs, error) {
	var dps []mi.DataPlaneWithType

	err := repo.GetContextDB(ctx).Table("data_plane as dp").
		Select("id", "name", "typ", "params", "capability",
			"type_name", "type_display_name", "desc", "param_spec", "capability_spec").
		Where("is_deleted = ?", false).
		Joins("LEFT JOIN data_plane_type dpt ON dp.typ = dpt.type_name").
		Order("id asc").
		Find(&dps).Error

	if err != nil {
		return nil, err
	}

	var c2Ds []mi.Cluster2DataPlane
	err = repo.GetContextDB(ctx).Model(&mi.Cluster2DataPlane{}).
		Find(&c2Ds).Error

	if err != nil {
		return nil, err
	}

	var res []mi.DataPlaneWithClusterIDs
	for i := 0; i < len(dps); i++ {
		var clusterIDs []string
		for _, c2D := range c2Ds {
			if c2D.DataPlaneID == dps[i].ID {
				clusterIDs = append(clusterIDs, c2D.ClusterID)
			}
		}
		res = append(res, mi.DataPlaneWithClusterIDs{
			DataPlaneWithType: dps[i],
			ClusterIDs:        clusterIDs,
		})
	}
	return res, nil

}

func (repo *daoRepo) CheckDataPlaneExist(ctx core.Context, id int) (bool, error) {
	var count int64
	err := repo.GetContextDB(ctx).Model(&mi.DataPlane{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (repo *daoRepo) UpdateDataPlane(ctx core.Context, d *mi.DataPlane) error {
	return repo.GetContextDB(ctx).Save(d).Error
}

func (repo *daoRepo) DeleteDataPlane(ctx core.Context, id int) error {
	return repo.GetContextDB(ctx).Model(&mi.DataPlane{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

func (repo *daoRepo) GetDataplane(ctx core.Context, id int) (*mi.DataPlane, error) {
	var res mi.DataPlane
	err := repo.GetContextDB(ctx).Where("id = ?", id).First(&res).Error
	return &res, err
}

func (repo *daoRepo) GetDatapalenDsCount(ctx core.Context) ([]model.DatasourceCount, error) {
	var results []model.DatasourceCount
	if err := repo.DB.Table("data_plane").
		Select("typ, COUNT(*) AS record_count").
		Group("typ").
		Where("is_deleted", false).
		Find(&results).
		Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (repo *daoRepo) DeleteCluster2DataPlane(ctx core.Context, id int) error {
	return repo.GetContextDB(ctx).Model(&mi.Cluster2DataPlane{}).
		Where("data_plane_id = ?", id).
		Delete(&mi.Cluster2DataPlane{}).Error
}

func (repo *daoRepo) InitAPODataplane(ctx core.Context) error {
	var count int64
	err := repo.GetContextDB(ctx).Model(&mi.DataPlaneType{}).Where("type_name = ?", "apo").Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		apoDataplaneType := &mi.DataPlaneType{
			TypeName:        "apo",
			ParamSpec:       mi.JSONField[mi.ParamSpec]{Obj: mi.ParamSpec{Name: "root", Type: "object"}},
			CapabilitySpec:  mi.JSONField[[]string]{Obj: []string{"RedMetric", "ServiceInstance", "Topology", "Log", "Trace", "Metrics"}},
			TypeDisplayName: "APO",
		}
		err = repo.GetContextDB(ctx).Create(apoDataplaneType).Error
		if err != nil {
			return err
		}
	}

	err = repo.GetContextDB(ctx).Model(&mi.DataPlane{}).Where("name = ?", "apo").Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		apoDataplane := &mi.DataPlane{
			Name:       "apo",
			Typ:        "apo",
			Capability: mi.JSONField[[]string]{Obj: []string{"RedMetric", "ServiceInstance", "Topology", "Log", "Trace", "Metrics"}},
			Params:     "{}",
		}
		err = repo.GetContextDB(ctx).Create(apoDataplane).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *daoRepo) CheckGetApoMetric(ctx core.Context) (bool, error) {
	var dps []mi.DataPlane
	err := repo.GetContextDB(ctx).Model(&mi.DataPlane{}).
		Where("is_deleted = ?", false).Find(&dps).Error

	for _, dp := range dps {
		if dp.Typ != "apo" {
			if slices.Contains(dp.Capability.Obj, "ServiceInstance") {
				return false, nil
			}
		}
	}
	return true, err
}
