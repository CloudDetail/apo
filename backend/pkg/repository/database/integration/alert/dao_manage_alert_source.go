// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	input "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/google/uuid"
)

const (
	AlertSource2Cluster = `SELECT * from clusters where id in
(select cluster_id from alert_source2_clusters where source_id = ?)`
)

func (repo *subRepo) CreateAlertSource(ctx core.Context, alertSource *alert.AlertSource) error {
	newS2C := []alert.AlertSource2Cluster{}
	for _, cluster := range alertSource.Clusters {
		if cluster.ID == "" {
			cluster.ID = uuid.NewString()
			err := repo.GetContextDB(ctx).Create(&cluster).Error
			if err != nil {
				return err
			}
		}

		newS2C = append(newS2C, alert.AlertSource2Cluster{
			SourceID:  alertSource.SourceID,
			ClusterID: cluster.ID,
		})
	}

	if len(newS2C) > 0 {
		err := repo.GetContextDB(ctx).Create(&newS2C).Error
		if err != nil {
			return err
		}
	}

	return repo.GetContextDB(ctx).Create(&alertSource).Error
}

func (repo *subRepo) GetAlertSource(ctx core.Context, sourceId string) (*alert.AlertSource, error) {
	var res alert.AlertSource
	err := repo.GetContextDB(ctx).First(&res, "source_id = ?", sourceId).Error
	if err == nil {
		var clusters []input.Cluster
		err := repo.GetContextDB(ctx).Raw(AlertSource2Cluster, res.SourceID).Scan(&clusters).Error
		if err == nil {
			res.Clusters = clusters
		}
	}

	return &res, err
}

func (repo *subRepo) UpdateAlertSource(ctx core.Context, alertSource *alert.AlertSource) error {
	err := repo.GetContextDB(ctx).Delete(&alert.AlertSource2Cluster{}, "source_id = ?", alertSource.SourceID).Error
	if err != nil {
		return err
	}
	newS2C := []alert.AlertSource2Cluster{}
	for _, cluster := range alertSource.Clusters {
		newS2C = append(newS2C, alert.AlertSource2Cluster{
			SourceID:  alertSource.SourceID,
			ClusterID: cluster.ID,
		})
	}

	if len(newS2C) > 0 {
		err = repo.GetContextDB(ctx).Create(&newS2C).Error
		if err != nil {
			return err
		}
	}

	return repo.GetContextDB(ctx).Model(&alert.AlertSource{}).
		Where("source_id = ?", alertSource.SourceID).
		Updates(alertSource).Error
}

func (repo *subRepo) ListAlertSource(ctx core.Context) ([]alert.AlertSource, error) {
	var alertSources []alert.AlertSource
	err := repo.GetContextDB(ctx).Find(&alertSources, "source_name NOT LIKE ?", "APO_DEFAULT_ENRICH_RULE%").Error
	if err != nil {
		return nil, err
	}

	var clusters []input.Cluster
	err = repo.GetContextDB(ctx).Find(&clusters).Error
	if err != nil {
		return nil, err
	}

	var s2cs []alert.AlertSource2Cluster
	err = repo.GetContextDB(ctx).Find(&s2cs).Error

	var tmpClustersMap = make(map[string]input.Cluster)
	for i := 0; i < len(clusters); i++ {
		tmpClustersMap[clusters[i].ID] = clusters[i]
	}

	for _, s2c := range s2cs {
		for i := 0; i < len(alertSources); i++ {
			if alertSources[i].SourceID == s2c.SourceID {
				alertSources[i].Clusters = append(alertSources[i].Clusters, tmpClustersMap[s2c.ClusterID])
				break
			}
		}
	}

	return alertSources, err
}

func (repo *subRepo) DeleteAlertSource(ctx core.Context, alertSource alert.SourceFrom) (*alert.AlertSource, error) {
	deletedSource := alert.AlertSource{}
	err := repo.GetContextDB(ctx).First(&deletedSource, "source_id = ?", alertSource.SourceID).Error

	if err != nil || len(deletedSource.SourceID) == 0 {
		return nil, err
	}

	err = repo.GetContextDB(ctx).Delete(&alert.AlertSource2Cluster{}, "source_id = ?", alertSource.SourceID).Error
	if err != nil {
		return nil, err
	}

	err = repo.GetContextDB(ctx).Delete(&alert.AlertSource{}, "source_id = ?", alertSource.SourceID).Error
	if err != nil {
		return nil, err
	}
	return &deletedSource, nil
}
