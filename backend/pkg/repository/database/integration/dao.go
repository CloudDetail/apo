// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"errors"
	"fmt"

	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/integration/alert"
	"gorm.io/gorm"
)

var IndependentTraceAPI = false
var IndependentMetricDatasource = false
var IndependentLogDatabase = false

type ObservabilityInputManage interface {
	// Manage Cluster
	CreateCluster(ctx core.Context, cluster *integration.Cluster) error
	UpdateCluster(ctx core.Context, cluster *integration.Cluster) error
	DeleteCluster(ctx core.Context, cluster *integration.Cluster) error
	ListCluster(ctx core.Context) ([]integration.Cluster, error)
	GetCluster(ctx core.Context, clusterID string) (integration.Cluster, error)
	CheckClusterNameExisted(ctx core.Context, clusterName string) (bool, error)

	SaveIntegrationConfig(ctx core.Context, iConfig integration.ClusterIntegration) error
	GetIntegrationConfig(ctx core.Context, clusterID string) (*integration.ClusterIntegration, error)
	DeleteIntegrationConfig(ctx core.Context, clusterID string) error

	GetLatestTraceAPIs(ctx core.Context, lastUpdateTS int64) (*integration.AdapterAPIConfig, error)

	alert.AlertInput
}

type subRepos struct {
	*driver.DB

	alert.AlertInput
}

func NewObservabilityInputManage(db *gorm.DB, cfg *config.Config) (*subRepos, error) {
	if db == nil {
		return nil, errors.New("database is not ready yet")
	}

	subRepos := &subRepos{
		DB: &driver.DB{DB: db},
	}

	var err error
	if subRepos.AlertInput, err = alert.NewAlertInputRepo(db, cfg); err != nil {
		return nil, fmt.Errorf("failed to init observability input manage, err: %v", err)
	}

	if err := subRepos.GetContextDB(nil).AutoMigrate(
		&integration.Cluster{},
		&integration.TraceIntegration{},
		&integration.MetricIntegration{},
		&integration.LogIntegration{},
	); err != nil {
		return nil, err
	}

	return subRepos, nil
}
