// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"testing"

	model "github.com/CloudDetail/apo/backend/pkg/model/integration"
	dbdriver "github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestListClusterReportsCompleteIntegrationConfiguration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Cluster{},
		&model.TraceIntegration{},
		&model.MetricIntegration{},
		&model.LogIntegration{},
	))

	require.NoError(t, db.Create(&[]model.Cluster{
		{ID: "configured", Name: "Configured", ClusterType: model.ClusterTypeK8s},
		{ID: "partial", Name: "Partial", ClusterType: model.ClusterTypeVM},
		{ID: "identity-only", Name: "Identity Only", ClusterType: model.ClusterTypeK8s},
	}).Error)
	require.NoError(t, db.Create(&[]model.TraceIntegration{
		{ClusterID: "configured"},
		{ClusterID: "partial"},
	}).Error)
	require.NoError(t, db.Create(&model.MetricIntegration{ClusterID: "configured"}).Error)
	require.NoError(t, db.Create(&model.LogIntegration{ClusterID: "configured"}).Error)

	repo := &subRepos{DB: &dbdriver.DB{DB: db}}
	clusters, err := repo.ListCluster(nil)
	require.NoError(t, err)

	configuredByID := make(map[string]bool, len(clusters))
	for _, cluster := range clusters {
		configuredByID[cluster.ID] = cluster.IntegrationConfigured
	}
	require.Equal(t, map[string]bool{
		"configured":    true,
		"partial":       false,
		"identity-only": false,
	}, configuredByID)
}
