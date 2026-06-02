// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCustomDashboardTestRepo(t *testing.T) *daoRepo {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&CustomDashboard{}); err != nil {
		t.Fatalf("migrate custom dashboard: %v", err)
	}
	return &daoRepo{DB: &driver.DB{DB: db}}
}

func TestListCustomDashboardsMatchesTeamIDsExactly(t *testing.T) {
	repo := newCustomDashboardTestRepo(t)

	rows := []CustomDashboard{
		{ID: "everyone", Name: "everyone", OwnerID: 100, Visibility: "everyone", TeamIDs: "[]", Config: "{}"},
		{ID: "personal", Name: "personal", OwnerID: 1, Visibility: "personal", TeamIDs: "[]", Config: "{}"},
		{ID: "team-2", Name: "team 2", OwnerID: 100, Visibility: "team", TeamIDs: "[2]", Config: "{}"},
		{ID: "team-12", Name: "team 12", OwnerID: 100, Visibility: "team", TeamIDs: "[12]", Config: "{}"},
	}
	if err := repo.GetContextDB(nil).Create(&rows).Error; err != nil {
		t.Fatalf("seed dashboards: %v", err)
	}

	got, err := repo.ListCustomDashboards(nil, 1, []int64{2}, "")
	if err != nil {
		t.Fatalf("list dashboards: %v", err)
	}

	visible := map[string]bool{}
	for _, row := range got {
		visible[row.ID] = true
	}
	if visible["team-12"] {
		t.Fatalf("team 2 user must not see dashboard shared only with team 12; got IDs %#v", visible)
	}
	for _, id := range []string{"everyone", "personal", "team-2"} {
		if !visible[id] {
			t.Fatalf("expected dashboard %q to be visible; got IDs %#v", id, visible)
		}
	}
}

func TestListCustomDashboardsAppliesKeywordToAllVisibleRows(t *testing.T) {
	repo := newCustomDashboardTestRepo(t)

	rows := []CustomDashboard{
		{ID: "matching", Name: "QA target dashboard", OwnerID: 100, Visibility: "everyone", TeamIDs: "[]", Config: "{}"},
		{ID: "owned-unmatched", Name: "owned unrelated", OwnerID: 1, Visibility: "personal", TeamIDs: "[]", Config: "{}"},
		{ID: "team-unmatched", Name: "team unrelated", OwnerID: 100, Visibility: "team", TeamIDs: "[2]", Config: "{}"},
	}
	if err := repo.GetContextDB(nil).Create(&rows).Error; err != nil {
		t.Fatalf("seed dashboards: %v", err)
	}

	got, err := repo.ListCustomDashboards(nil, 1, []int64{2}, "target")
	if err != nil {
		t.Fatalf("list dashboards: %v", err)
	}

	if len(got) != 1 || got[0].ID != "matching" {
		t.Fatalf("keyword must filter every visible row; got %#v", got)
	}
}

func TestCreateCustomDashboardWithMenuRollsBackDashboardWhenMenuFails(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:custom-dashboard-menu-rollback?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&CustomDashboard{},
		&Router{},
		&MenuItem{},
		&FeatureMapping{},
		&I18nTranslation{},
		&profile.Feature{},
		&AuthPermission{},
	); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}

	err = repo.CreateCustomDashboardWithMenu(nil, &CustomDashboard{
		ID:          "dashboard-1",
		Name:        "Dashboard 1",
		Description: "",
		OwnerID:     1,
		Visibility:  "everyone",
		TeamIDs:     "[]",
		Config:      "{}",
	})
	if err == nil {
		t.Fatalf("expected missing customDashboard parent menu to fail")
	}

	var count int64
	if err := db.Model(&CustomDashboard{}).Where("id = ?", "dashboard-1").Count(&count).Error; err != nil {
		t.Fatalf("count dashboard rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("dashboard row must roll back when menu mutation fails, got %d rows", count)
	}
}

func TestSyncCustomDashboardMenusBackfillsExistingDashboards(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:custom-dashboard-menu-backfill?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&CustomDashboard{},
		&Router{},
		&MenuItem{},
		&FeatureMapping{},
		&I18nTranslation{},
		&profile.Feature{},
		&AuthPermission{},
	); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}

	parentRouter := Router{RouterTo: "/custom-dashboard"}
	if err := db.Create(&parentRouter).Error; err != nil {
		t.Fatalf("seed parent router: %v", err)
	}
	parentMenu := MenuItem{Key: "customDashboard", RouterID: parentRouter.RouterID, Order: 25}
	if err := db.Create(&parentMenu).Error; err != nil {
		t.Fatalf("seed parent menu: %v", err)
	}
	parentFeature := profile.Feature{FeatureName: "我的大盘"}
	if err := db.Create(&parentFeature).Error; err != nil {
		t.Fatalf("seed parent feature: %v", err)
	}

	dashboardID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	if err := db.Create(&CustomDashboard{
		ID:         dashboardID,
		Name:       "Legacy Dashboard",
		OwnerID:    42,
		Visibility: "everyone",
		TeamIDs:    "[]",
		Config:     "{}",
	}).Error; err != nil {
		t.Fatalf("seed dashboard: %v", err)
	}

	if err := repo.syncCustomDashboardMenus(nil); err != nil {
		t.Fatalf("sync custom dashboard menus: %v", err)
	}

	var router Router
	route := customDashboardRoute(dashboardID)
	if err := db.Where("router_to = ?", route).First(&router).Error; err != nil {
		t.Fatalf("expected dashboard router to be backfilled: %v", err)
	}
	var menu MenuItem
	if err := db.Where(`"key" = ?`, customDashboardMenuKey(dashboardID)).First(&menu).Error; err != nil {
		t.Fatalf("expected dashboard menu item to be backfilled: %v", err)
	}
	if menu.ParentID == nil || *menu.ParentID != parentMenu.ItemID {
		t.Fatalf("expected menu parent %d, got %#v", parentMenu.ItemID, menu.ParentID)
	}
	if menu.RouterID != router.RouterID {
		t.Fatalf("expected menu router %d, got %d", router.RouterID, menu.RouterID)
	}

	var feature profile.Feature
	if err := db.Where("feature_name = ? AND custom = ?", customDashboardMenuKey(dashboardID), true).First(&feature).Error; err != nil {
		t.Fatalf("expected dashboard feature to be backfilled: %v", err)
	}
	if feature.ParentID == nil || *feature.ParentID != parentFeature.FeatureID {
		t.Fatalf("expected feature parent %d, got %#v", parentFeature.FeatureID, feature.ParentID)
	}

	for _, mapping := range []struct {
		mappedID   int
		mappedType string
	}{
		{menu.ItemID, "menu"},
		{router.RouterID, "router"},
	} {
		var count int64
		if err := db.Model(&FeatureMapping{}).
			Where("feature_id = ? AND mapped_id = ? AND mapped_type = ?", feature.FeatureID, mapping.mappedID, mapping.mappedType).
			Count(&count).Error; err != nil {
			t.Fatalf("count feature mapping: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected one %s mapping for %d, got %d", mapping.mappedType, mapping.mappedID, count)
		}
	}

	var permissionCount int64
	if err := db.Model(&AuthPermission{}).
		Where("type = ? AND subject_id = ? AND subject_type = ? AND permission_id = ?", "feature", int64(42), "user", feature.FeatureID).
		Count(&permissionCount).Error; err != nil {
		t.Fatalf("count owner permission: %v", err)
	}
	if permissionCount != 1 {
		t.Fatalf("expected owner feature permission to be backfilled, got %d", permissionCount)
	}
}
