package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	mi "github.com/CloudDetail/apo/backend/pkg/model/integration"
)

func (repo *daoRepo) ListDataPlaneType(ctx core.Context) ([]mi.DataPlaneType, error) {
	var res []mi.DataPlaneType
	err := repo.GetContextDB(ctx).Find(&res).Error
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

func (repo *daoRepo) ListDataPlane(ctx core.Context) ([]mi.DataPlaneWithClusterIDs, error) {
	var dps []mi.DataPlaneWithType

	err := repo.GetContextDB(ctx).Table("data_plane as dp").
		Select("id", "name", "typ", "params", "capability",
			"type_name", "desc", "param_spec", "capability_spec").
		Joins("LEFT JOIN data_plane_type dpt ON dp.typ = dpt.type_name").
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
	// TODO update time
	return repo.GetContextDB(ctx).Save(d).Error
}

func (repo *daoRepo) DeleteDataPlane(ctx core.Context, id int) error {
	return repo.GetContextDB(ctx).Where("id = ?", id).Delete(&mi.DataPlane{}).Error
}
