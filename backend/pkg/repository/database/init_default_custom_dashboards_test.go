// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"encoding/json"
	"path/filepath"
	"testing"

	customdashboardmodel "github.com/CloudDetail/apo/backend/pkg/model/customdashboard"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDefaultCustomDashboardSQLIsIdempotentAndPreservesChanges(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:default-custom-dashboard-sql?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&CustomDashboard{}, &profile.User{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}
	if err := db.Create(&profile.User{UserID: 100, Username: "admin"}).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}

	path := filepath.Join("..", "..", "..", "sqlscripts", "default_custom_dashboards", "default_custom_dashboards-sqlite.sql")
	if err := executeDefaultCustomDashboardSQL(db, path); err != nil {
		t.Fatalf("execute default dashboard SQL: %v", err)
	}

	var dashboards []CustomDashboard
	if err := db.Order("id").Find(&dashboards).Error; err != nil {
		t.Fatalf("list imported dashboards: %v", err)
	}
	if len(dashboards) != 18 {
		t.Fatalf("expected 18 imported dashboards, got %d", len(dashboards))
	}
	for _, row := range dashboards {
		if row.OwnerID != 100 {
			t.Fatalf("dashboard %s owner = %d, want 100", row.ID, row.OwnerID)
		}
		var dashboard customdashboardmodel.Dashboard
		if err := json.Unmarshal([]byte(row.Config), &dashboard); err != nil {
			t.Fatalf("dashboard %s config is invalid JSON: %v", row.ID, err)
		}
		if len(dashboard.Widgets) == 0 || len(dashboard.Widgets) > maxDashboardWidgetsForSQLTest {
			t.Fatalf("dashboard %s widget count = %d", row.ID, len(dashboard.Widgets))
		}
	}

	changedID := dashboards[0].ID
	if err := db.Model(&CustomDashboard{}).Where("id = ?", changedID).Update("name", "用户修改后的名称").Error; err != nil {
		t.Fatalf("modify imported dashboard: %v", err)
	}
	if err := executeDefaultCustomDashboardSQL(db, path); err != nil {
		t.Fatalf("re-execute default dashboard SQL: %v", err)
	}

	var count int64
	if err := db.Model(&CustomDashboard{}).Count(&count).Error; err != nil {
		t.Fatalf("count imported dashboards: %v", err)
	}
	if count != 18 {
		t.Fatalf("expected 18 dashboards after repeated initialization, got %d", count)
	}
	var changed CustomDashboard
	if err := db.Where("id = ?", changedID).First(&changed).Error; err != nil {
		t.Fatalf("reload modified dashboard: %v", err)
	}
	if changed.Name != "用户修改后的名称" {
		t.Fatalf("repeated initialization overwrote user change: %q", changed.Name)
	}
}

const maxDashboardWidgetsForSQLTest = 60

func TestDefaultCustomDashboardSQLPath(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql", "postgres"} {
		path, err := defaultCustomDashboardSQLPath("sql", dialect)
		if err != nil {
			t.Fatalf("dialect %s: %v", dialect, err)
		}
		want := filepath.Join("sql", "default_custom_dashboards-"+dialect+".sql")
		if path != want {
			t.Fatalf("dialect %s path = %q, want %q", dialect, path, want)
		}
	}
	if _, err := defaultCustomDashboardSQLPath("sql", "unknown"); err == nil {
		t.Fatal("unsupported dialect should fail")
	}
}
