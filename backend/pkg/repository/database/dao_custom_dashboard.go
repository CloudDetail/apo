// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"strconv"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

type CustomDashboard struct {
	ID            string         `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	Name          string         `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Description   string         `gorm:"column:description;type:text" json:"description"`
	OwnerID       int64          `gorm:"column:owner_id;index;not null" json:"ownerId"`
	Visibility    string         `gorm:"column:visibility;type:varchar(32);index;not null" json:"visibility"`
	TeamIDs       string         `gorm:"column:team_ids;type:text" json:"teamIds"`
	Config        string         `gorm:"column:config;type:text" json:"config"`
	LastVisitedAt *time.Time     `gorm:"column:last_visited_at" json:"lastVisitedAt"`
	CreatedAt     time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CustomDashboard) TableName() string {
	return "custom_dashboard"
}

type CustomDashboardQueryAudit struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64     `gorm:"column:user_id;index" json:"userId"`
	DashboardID      string    `gorm:"column:dashboard_id;type:varchar(64);index" json:"dashboardId"`
	WidgetID         string    `gorm:"column:widget_id;type:varchar(64)" json:"widgetId"`
	QueryMode        string    `gorm:"column:query_mode;type:varchar(32)" json:"queryMode"`
	Expression       string    `gorm:"column:expression;type:text" json:"expression"`
	ResolvedQuery    string    `gorm:"column:resolved_query;type:text" json:"resolvedQuery"`
	ResolvedVars     string    `gorm:"column:resolved_vars;type:text" json:"resolvedVars"`
	ConstrainedScope string    `gorm:"column:constrained_scope;type:text" json:"constrainedScope"`
	Status           string    `gorm:"column:status;type:varchar(32)" json:"status"`
	ErrorMessage     string    `gorm:"column:error_message;type:text" json:"errorMessage"`
	ResultSeries     int       `gorm:"column:result_series" json:"resultSeries"`
	ResultPoints     int       `gorm:"column:result_points" json:"resultPoints"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (CustomDashboardQueryAudit) TableName() string {
	return "custom_dashboard_query_audit"
}

func (repo *daoRepo) ListCustomDashboards(ctx core.Context, userID int64, teamIDs []int64, keyword string) ([]CustomDashboard, error) {
	db := repo.GetContextDB(ctx).Model(&CustomDashboard{})
	if keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	visibility := repo.GetContextDB(ctx).Where("visibility = ?", "everyone").
		Or("owner_id = ?", userID).
		Or("visibility = ? AND owner_id = ?", "personal", userID)
	if len(teamIDs) > 0 {
		for _, teamID := range teamIDs {
			teamIDText := strconv.FormatInt(teamID, 10)
			visibility = visibility.Or(
				"visibility = ? AND (team_ids = ? OR team_ids LIKE ? OR team_ids LIKE ? OR team_ids LIKE ?)",
				"team",
				"["+teamIDText+"]",
				"["+teamIDText+",%",
				"%,"+teamIDText+",%",
				"%,"+teamIDText+"]",
			)
		}
	}
	var dashboards []CustomDashboard
	err := db.Where(visibility).Order("updated_at DESC").Find(&dashboards).Error
	return dashboards, err
}

func (repo *daoRepo) GetCustomDashboard(ctx core.Context, id string) (*CustomDashboard, error) {
	var dashboard CustomDashboard
	err := repo.GetContextDB(ctx).Where("id = ?", id).First(&dashboard).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &dashboard, err
}

func (repo *daoRepo) CreateCustomDashboard(ctx core.Context, dashboard *CustomDashboard) error {
	return repo.GetContextDB(ctx).Create(dashboard).Error
}

func (repo *daoRepo) CreateCustomDashboardWithMenu(ctx core.Context, dashboard *CustomDashboard) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(dashboard).Error; err != nil {
			return err
		}
		return repo.upsertCustomDashboardMenuTx(ctx, tx, dashboard.ID, dashboard.Name, dashboard.Visibility)
	})
}

func (repo *daoRepo) UpdateCustomDashboard(ctx core.Context, dashboard *CustomDashboard) error {
	return repo.GetContextDB(ctx).Model(&CustomDashboard{}).
		Where("id = ?", dashboard.ID).
		Updates(map[string]any{
			"name":        dashboard.Name,
			"description": dashboard.Description,
			"visibility":  dashboard.Visibility,
			"team_ids":    dashboard.TeamIDs,
			"config":      dashboard.Config,
		}).Error
}

func (repo *daoRepo) UpdateCustomDashboardWithMenu(ctx core.Context, dashboard *CustomDashboard) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&CustomDashboard{}).
			Where("id = ?", dashboard.ID).
			Updates(map[string]any{
				"name":        dashboard.Name,
				"description": dashboard.Description,
				"visibility":  dashboard.Visibility,
				"team_ids":    dashboard.TeamIDs,
				"config":      dashboard.Config,
			}).Error; err != nil {
			return err
		}
		return repo.upsertCustomDashboardMenuTx(ctx, tx, dashboard.ID, dashboard.Name, dashboard.Visibility)
	})
}

func (repo *daoRepo) DeleteCustomDashboard(ctx core.Context, id string) error {
	return repo.GetContextDB(ctx).Where("id = ?", id).Delete(&CustomDashboard{}).Error
}

func (repo *daoRepo) DeleteCustomDashboardWithMenu(ctx core.Context, id string) error {
	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&CustomDashboard{}).Error; err != nil {
			return err
		}
		return repo.deleteCustomDashboardMenuTx(tx, id)
	})
}

func (repo *daoRepo) TouchCustomDashboardVisit(ctx core.Context, id string, at time.Time) error {
	return repo.GetContextDB(ctx).Model(&CustomDashboard{}).Where("id = ?", id).Update("last_visited_at", at).Error
}

func (repo *daoRepo) CreateCustomDashboardQueryAudit(ctx core.Context, audit *CustomDashboardQueryAudit) error {
	return repo.GetContextDB(ctx).Create(audit).Error
}
