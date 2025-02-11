// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"encoding/json"
	"errors"

	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"go.uber.org/multierr"
	"gorm.io/gorm"
)

func (repo *subRepos) updateTraceIntegration(t *integration.TraceIntegration) error {
	var updateError error

	if t.Mode == "sidecar" {
		oldAPI := integration.TraceIntegration{}
		err := repo.db.First(&oldAPI, "apm_type = ? and mode = 'sidecar'", t.ApmType).
			Order("updated_at DESC").Error
		if err == nil {
			t.TraceAPI.AcceptExistedSecret(oldAPI.TraceAPI.Obj)
		}
	}

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

	// oldAPI := integration.MetricIntegration{}
	// err := repo.db.First(&oldAPI, "ds_type = ? ", m.DSType).
	// 	Order("updated_at DESC").Error
	// if err == nil {
	// 	m.MetricAPI.AcceptExistedSecret(oldAPI.MetricAPI.Obj)
	// }
	// m.Mode = "pql"
	// } else if errors.Is(err, gorm.ErrRecordNotFound) {
	// 	vmCfg := config.Get().Promethues
	// 	m.MetricAPI.Obj.VMConfig.Username = vmCfg
	// 	m.MetricAPI.Obj.VMConfig.Password = vmCfg.Password
	// }

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

	// oldAPI := integration.LogIntegration{}
	// err := repo.db.First(&oldAPI, "db_type = ? ", l.DBType).
	// 	Order("updated_at DESC").Error
	// if err == nil {
	// 	l.LogAPI.AcceptExistedSecret(oldAPI.LogAPI.Obj)
	// } else if errors.Is(err, gorm.ErrRecordNotFound) {
	// 	chCfg := config.Get().ClickHouse
	// 	l.LogAPI.Obj.Clickhouse.UserName = chCfg.Username
	// 	l.LogAPI.Obj.Clickhouse.Password = chCfg.Password
	// }
	// l.Mode = "sql"

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
	err = repo.db.Find(&traceIntegration, "cluster_id = ?", clusterID, "is_deleted = ?", false).Error
	if err != nil {
		return res, err
	}
	res.Trace = traceIntegration

	var metricIntegration integration.MetricIntegration
	err = repo.db.Find(&metricIntegration, "cluster_id = ?", clusterID, "is_deleted = ?", false).Error
	if err != nil {
		return res, err
	}
	res.Metric = metricIntegration

	var logIntegration integration.LogIntegration
	err = repo.db.Find(&logIntegration, "cluster_id = ?", clusterID, "is_deleted = ?", false).Error
	if err != nil {
		return res, err
	}
	res.Log = logIntegration

	return res, err
}

func (repo *subRepos) DeleteIntegrationConfig(clusterID string) error {
	err := repo.db.Model(&integration.TraceIntegration{}).
		Where("cluster_id = ?", clusterID).
		Update("is_deleted", true).Error

	err2 := repo.db.Model(&integration.MetricIntegration{}).
		Where("cluster_id = ?", clusterID).
		Update("is_deleted", true).Error

	err = multierr.Append(err, err2)

	err3 := repo.db.Model(&integration.LogIntegration{}).
		Where("cluster_id = ?", clusterID).
		Update("is_deleted", true).Error

	err = multierr.Append(err, err3)

	return err
}

type traceAPI struct {
	ApmType  string `gorm:"apm_type"`
	TraceAPI string `gorm:"trace_api"`
}

func (repo *subRepos) GetLatestTraceAPIs(lastUpdateTS int64) (*integration.AdapterAPIConfig, error) {
	var latestUpdateTraceAPI integration.TraceIntegration
	err := repo.db.First(&latestUpdateTraceAPI, "updated_at > ?", lastUpdateTS).
		Order("updated_at DESC").Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &integration.AdapterAPIConfig{
			APIs:    map[string]any{},
			Timeout: 0,
		}, nil
	} else if err != nil {
		return nil, err
	}

	var latestTraceAPIs []traceAPI
	sql := `WITH latestAPI AS (
  SELECT apm_type,trace_api,
    ROW_NUMBER() OVER (PARTITION BY apm_type ORDER BY updated_at DESC) AS rn
  FROM trace_integrations WHERE is_deleted = false)
SELECT apm_type, trace_api
FROM latestAPI
WHERE rn = 1`

	err = repo.db.Raw(sql).Scan(&latestTraceAPIs).Error

	if err != nil {
		return nil, err
	}

	latestAPI := make(map[string]any)
	var apmList []string
	var maxTimeout int = 0
	for _, api := range latestTraceAPIs {
		var apiSpec map[string]interface{}
		// 反序列化 JSON 数据到 map
		err := json.Unmarshal([]byte(api.TraceAPI), &apiSpec)
		if err != nil {
			continue
		}

		cfg, ok := apiSpec[api.ApmType]
		if !ok {
			continue
		}

		latestAPI[api.ApmType] = cfg
		apmList = append(apmList, api.ApmType)
		timeout, ok := apiSpec["timeout"]
		if ok && timeout.(int) > maxTimeout {
			maxTimeout = timeout.(int)
		}
	}

	latestAPI["apm_list"] = apmList
	return &integration.AdapterAPIConfig{
		APIs:    latestAPI,
		Timeout: maxTimeout,
	}, nil
}
