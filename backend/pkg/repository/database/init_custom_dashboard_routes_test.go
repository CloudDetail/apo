// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCustomDashboardInitRepo(t *testing.T, name string, models ...interface{}) (*daoRepo, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	return &daoRepo{DB: &driver.DB{DB: db}}, db
}

func TestCustomDashboardDoesNotExposeStandaloneNewRoute(t *testing.T) {
	if isValidRouter("/custom-dashboard/new") {
		t.Fatalf("custom dashboard creation should stay an in-page action, not a standalone router")
	}

	for _, routerTo := range []string{"/custom-dashboard", "/custom-dashboard/:id", "/custom-dashboard/:id/edit"} {
		if !isValidRouter(routerTo) {
			t.Fatalf("expected custom dashboard router %q to remain valid", routerTo)
		}
	}
}

func TestCustomDashboardDoesNotExposeStandaloneCreateFeature(t *testing.T) {
	if isValidFeature("创建自定义大盘") {
		t.Fatalf("custom dashboard creation should not be a standalone feature permission")
	}

	if !isValidFeature("我的大盘") {
		t.Fatalf("custom dashboard menu permission should remain valid")
	}
}

func TestInitFeatureRouterDoesNotMapCustomDashboardNewRoute(t *testing.T) {
	repo, db := newCustomDashboardInitRepo(
		t,
		"init-feature-router-custom-dashboard",
		&profile.Feature{},
		&Router{},
		&FeatureMapping{},
	)

	feature := profile.Feature{FeatureName: "我的大盘"}
	if err := db.Create(&feature).Error; err != nil {
		t.Fatalf("seed feature: %v", err)
	}

	routers := []Router{
		{RouterTo: "/custom-dashboard"},
		{RouterTo: "/custom-dashboard/new"},
		{RouterTo: "/custom-dashboard/:id"},
		{RouterTo: "/custom-dashboard/:id/edit"},
	}
	if err := db.Create(&routers).Error; err != nil {
		t.Fatalf("seed routers: %v", err)
	}

	if err := repo.initFeatureRouter(nil); err != nil {
		t.Fatalf("init feature router: %v", err)
	}

	for _, routerTo := range []string{"/custom-dashboard", "/custom-dashboard/:id", "/custom-dashboard/:id/edit"} {
		var router Router
		if err := db.Where("router_to = ?", routerTo).First(&router).Error; err != nil {
			t.Fatalf("query router %q: %v", routerTo, err)
		}

		var count int64
		if err := db.Model(&FeatureMapping{}).
			Where("feature_id = ? AND mapped_id = ? AND mapped_type = ?", feature.FeatureID, router.RouterID, model.MAPPED_TYP_ROUTER).
			Count(&count).Error; err != nil {
			t.Fatalf("count mapping for %q: %v", routerTo, err)
		}
		if count != 1 {
			t.Fatalf("expected router %q to be mapped to 我的大盘 once, got %d", routerTo, count)
		}
	}

	var staleRouter Router
	if err := db.Where("router_to = ?", "/custom-dashboard/new").First(&staleRouter).Error; err != nil {
		t.Fatalf("query stale router: %v", err)
	}

	var staleCount int64
	if err := db.Model(&FeatureMapping{}).
		Where("feature_id = ? AND mapped_id = ? AND mapped_type = ?", feature.FeatureID, staleRouter.RouterID, model.MAPPED_TYP_ROUTER).
		Count(&staleCount).Error; err != nil {
		t.Fatalf("count stale mapping: %v", err)
	}
	if staleCount != 0 {
		t.Fatalf("stale /custom-dashboard/new router must not be mapped to 我的大盘, got %d mappings", staleCount)
	}
}

func TestSyncauseInitKeepsCustomDashboardAccessSeeds(t *testing.T) {
	t.Setenv("PRODUCT_MODE", "syncause")
	repo, db := newCustomDashboardInitRepo(
		t,
		"syncause-custom-dashboard-access",
		&profile.Feature{},
		&Router{},
		&MenuItem{},
		&FeatureMapping{},
	)

	if err := repo.initFeature(nil); err != nil {
		t.Fatalf("init feature: %v", err)
	}
	if err := repo.initRouterData(nil); err != nil {
		t.Fatalf("init routers: %v", err)
	}
	if err := repo.initMenuItems(nil); err != nil {
		t.Fatalf("init menu items: %v", err)
	}
	if err := repo.initFeatureMenuItems(nil); err != nil {
		t.Fatalf("init feature menu mappings: %v", err)
	}
	if err := repo.initFeatureRouter(nil); err != nil {
		t.Fatalf("init feature router mappings: %v", err)
	}

	var feature profile.Feature
	if err := db.Where("feature_name = ?", "我的大盘").First(&feature).Error; err != nil {
		t.Fatalf("syncause must seed custom dashboard feature: %v", err)
	}
	var menu MenuItem
	if err := db.Where(`"key" = ?`, "customDashboard").First(&menu).Error; err != nil {
		t.Fatalf("syncause must seed custom dashboard root menu: %v", err)
	}
	if menu.RouterID == 0 {
		t.Fatalf("custom dashboard root menu must point at /custom-dashboard router")
	}

	for _, routerTo := range []string{"/custom-dashboard", "/custom-dashboard/:id", "/custom-dashboard/:id/edit"} {
		var router Router
		if err := db.Where("router_to = ?", routerTo).First(&router).Error; err != nil {
			t.Fatalf("syncause must seed router %q: %v", routerTo, err)
		}
		var count int64
		if err := db.Model(&FeatureMapping{}).
			Where("feature_id = ? AND mapped_id = ? AND mapped_type = ?", feature.FeatureID, router.RouterID, model.MAPPED_TYP_ROUTER).
			Count(&count).Error; err != nil {
			t.Fatalf("count router mapping for %q: %v", routerTo, err)
		}
		if count != 1 {
			t.Fatalf("expected router %q to be mapped to 我的大盘 once, got %d", routerTo, count)
		}
	}
}
