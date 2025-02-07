// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"go.uber.org/multierr"
)

func (repo *subRepos) updateTraceIntegration(t *integration.TraceIntegration) error {
	var updateError error

	updateError = repo.db.Save(t).Error

	if !IndependentTraceAPI && t.Mode == "sidecar" {
		// since traceAPI is not independent now
		// Any change will affect all sidecar trace integrations
		err := repo.db.Model(&integration.TraceIntegration{}).
			Where("apm_type = ? and mode = 'sidecar'", t.ApmType).
			Update("trace_api", t.TraceAPI).Error

		updateError = multierr.Append(updateError, err)
	}

	return updateError
}

// save basic datasource for the platform, usually come from configmap
func (repo *subRepos) UpdateAllMetricIntegration(m *integration.MetricIntegration) error {
	m.ClusterID = integration.PlatformClusterID.String()
	return repo.updateMetricIntegration(m)
}

func (repo *subRepos) updateMetricIntegration(m *integration.MetricIntegration) error {
	var updateError error
	updateError = repo.db.Save(m).Error

	if !IndependentMetricDatasource {
		// Since Metric integration is not independent now
		// Any change will affect all metric integrations
		err := repo.db.Model(&integration.MetricIntegration{}).Omit("cluster_id").
			Where("ds_type = ?", m.DSType).
			Updates(m).Error
		updateError = multierr.Append(updateError, err)
	}

	return updateError
}

// same as UpdateAllMetricIntegration
func (repo *subRepos) UpdateAllLogIntegration(l *integration.LogIntegration) error {
	l.ClusterID = integration.PlatformClusterID.String()
	return repo.updateLogIntegration(l)
}

func (repo *subRepos) updateLogIntegration(l *integration.LogIntegration) error {
	var updateError error
	updateError = repo.db.Save(l).Error

	if !IndependentLogDatabase {
		// same as updateMetricIntegration
		err := repo.db.Model(&integration.LogIntegration{}).Omit("cluster_id").
			Where("db_type = ?", l.DBType).
			Updates(l).Error
		updateError = multierr.Append(updateError, err)
	}

	return updateError
}

func (repo *subRepos) SaveIntegrationConfig(iConfig integration.ClusterIntegration) error {
	// update clusterType of cluster
	err := repo.db.Model(&integration.Cluster{}).
		Where("id = ?", iConfig.ClusterID).
		Update("cluster_type", iConfig.ClusterType).Error

	if err != nil {
		return err
	}

	iConfig.Trace.ClusterID = iConfig.ClusterID
	iConfig.Metric.ClusterID = iConfig.ClusterID
	iConfig.Log.ClusterID = iConfig.ClusterID

	var storeErr error

	storeErr = repo.updateTraceIntegration(&iConfig.Trace)

	err = repo.updateMetricIntegration(&iConfig.Metric)
	storeErr = multierr.Append(storeErr, err)

	err = repo.updateLogIntegration(&iConfig.Log)
	storeErr = multierr.Append(storeErr, err)

	return storeErr
}

// get integration config for the cluster
func (repo *subRepos) GetIntegrationConfig(clusterID string) (*integration.ClusterIntegration, error) {
	cluster, err := repo.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}

	var res = &integration.ClusterIntegration{
		ClusterID:   clusterID,
		ClusterType: cluster.ClusterType,
		ClusterName: cluster.Name,
	}

	var traceIntegration integration.TraceIntegration
	err = repo.db.Find(&traceIntegration, "cluster_id = ?", clusterID).Error
	if err != nil {
		return res, err
	}
	res.Trace = traceIntegration

	var metricIntegration integration.MetricIntegration
	err = repo.db.Find(&metricIntegration, "cluster_id = ?", clusterID).Error
	if err != nil {
		return res, err
	}
	res.Metric = metricIntegration

	var logIntegration integration.LogIntegration
	err = repo.db.Find(&logIntegration, "cluster_id = ?", clusterID).Error
	if err != nil {
		return res, err
	}
	res.Log = logIntegration

	return res, err
}

func (repo *subRepos) DeleteIntegrationConfig(clusterID string) error {
	err := repo.db.Delete(&integration.TraceIntegration{}, "cluster_id = ?", clusterID).Error
	if err != nil {
		return err
	}

	err = repo.db.Delete(&integration.MetricIntegration{}, "cluster_id = ?", clusterID).Error
	if err != nil {
		return err
	}

	return repo.db.Delete(&integration.LogIntegration{}, "cluster_id = ?", clusterID).Error
}

// func (repo *subRepos) GetLatestTraceAPIs() *integration.TraceAPI {
// 	var traceIntegrations []integration.TraceIntegration

// }
