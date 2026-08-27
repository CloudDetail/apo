// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"sync"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestMenuItemKeyColumnSupportsExecutiveDashboardKeys(t *testing.T) {
	s, err := schema.Parse(&MenuItem{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	field := s.LookUpField("Key")
	require.NotNil(t, field)
	assert.Equal(t, "varchar(64)", field.TagSettings["TYPE"])
}

func TestIsValidMenuItem(t *testing.T) {
	tests := []struct {
		name     string
		menuItem string
		want     bool
	}{
		{"Valid menu item: service", "service", true},
		{"Valid menu item: workflows", "workflows", true},
		{"Valid menu item: knowledgeBases", "knowledgeBases", true},
		{"Invalid menu item", "invalid", false},
		{"Empty menu item", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidMenuItem(tt.menuItem)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidRoleName(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		want     bool
	}{
		{"Valid role: admin", "admin", true},
		{"Valid role: viewer", "viewer", true},
		{"Invalid role", "manager", false},
		{"Empty role", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidRoleName(tt.roleName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidFeature(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
		want        bool
	}{
		{"Valid feature: 服务概览", "服务概览", true},
		{"Valid feature: 工作流", "工作流", true},
		{"Valid feature: 知识库管理", "知识库管理", true},
		{"Invalid feature", "feature3", false},
		{"Empty feature", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidFeature(tt.featureName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   bool
	}{
		{"Valid method: GET", "GET", true},
		{"Valid method: post", "post", true},
		{"Valid method: PUT", "PUT", true},
		{"Valid method: DELETE", "DELETE", true},
		{"Valid method: PATCH", "PATCH", true},
		{"Invalid method", "OPTIONS", false},
		{"Empty method", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidMethod(tt.method)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"Valid path: /api", "/api", true},
		{"Valid path: /users/123", "/users/123", true},
		{"Invalid path: no leading slash", "api", false},
		{"Invalid path: contains ;", "/api;test", false},
		{"Invalid path: contains &", "/api&test", false},
		{"Invalid path: contains '", "/api'test", false},
		{"Invalid path: contains =", "/api=test", false},
		{"Empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidRouter(t *testing.T) {
	tests := []struct {
		name     string
		routerTo string
		want     bool
	}{
		{"Valid router: /service", "/service", true},
		{"Valid router: /logs/full", "/logs/full", true},
		{"Valid router: /knowledge-bases", "/knowledge-bases", true},
		{"Valid router: /probe-management/config", "/probe-management/config", true},
		{"Valid router: /probe-management/log-collection", "/probe-management/log-collection", true},
		{"Valid router: /probe-management/clusters", "/probe-management/clusters", true},
		{"Invalid router", "/api/v3", false},
		{"Empty router", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidRouter(tt.routerTo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidPageURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"Valid page: grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level", "grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level", true},
		{"Invalid page", "/login", false},
		{"Empty page", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPageURL(tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInitFeatureRouterMapsAlertEventsToReport(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&profile.Feature{}, &Router{}, &FeatureMapping{}))

	repo := &daoRepo{DB: &driver.DB{DB: db}}

	require.NoError(t, db.Create(&profile.Feature{FeatureName: "告警事件"}).Error)
	require.NoError(t, db.Create(&Router{RouterTo: "/report"}).Error)

	require.NoError(t, repo.initFeatureRouter(nil))

	var count int64
	require.NoError(t, db.Model(&FeatureMapping{}).
		Joins("JOIN feature ON feature.feature_id = feature_mapping.feature_id").
		Joins("JOIN router ON router.router_id = feature_mapping.mapped_id").
		Where("feature.feature_name = ? AND router.router_to = ? AND feature_mapping.mapped_type = ?",
			"告警事件",
			"/report",
			model.MAPPED_TYP_ROUTER,
		).
		Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestInitFeatureRouterMapsProbePagesToProbeManagement(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&profile.Feature{}, &Router{}, &FeatureMapping{}))

	repo := &daoRepo{DB: &driver.DB{DB: db}}

	require.NoError(t, db.Create(&profile.Feature{FeatureName: "探针管理"}).Error)
	probeRoutes := []string{"/probe-management", "/probe-management/config", "/probe-management/log-collection", "/probe-management/clusters"}
	for _, routerTo := range probeRoutes {
		require.NoError(t, db.Create(&Router{RouterTo: routerTo}).Error)
	}

	require.NoError(t, repo.initFeatureRouter(nil))

	for _, routerTo := range probeRoutes {
		var count int64
		require.NoError(t, db.Model(&FeatureMapping{}).
			Joins("JOIN feature ON feature.feature_id = feature_mapping.feature_id").
			Joins("JOIN router ON router.router_id = feature_mapping.mapped_id").
			Where("feature.feature_name = ? AND router.router_to = ? AND feature_mapping.mapped_type = ?",
				"探针管理",
				routerTo,
				model.MAPPED_TYP_ROUTER,
			).
			Count(&count).Error)
		assert.Equal(t, int64(1), count, routerTo)
	}
}

func TestProbeClusterManagementRouteIsSeededForAllProductModes(t *testing.T) {
	const route = "/probe-management/clusters"

	for name, routers := range map[string][]Router{
		"default":  validRouters,
		"syncause": sreRouters,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, routers, Router{RouterTo: route, HideTimeSelector: true})
		})
	}
}
