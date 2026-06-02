// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

type customDashboardMenuRepoStub struct {
	database.Repo
	teams      map[int64][]int64
	dashboards map[string]*database.CustomDashboard
}

func (r *customDashboardMenuRepoStub) GetUserTeams(ctx core.Context, userID int64) ([]int64, error) {
	return r.teams[userID], nil
}

func (r *customDashboardMenuRepoStub) GetCustomDashboard(ctx core.Context, id string) (*database.CustomDashboard, error) {
	return r.dashboards[id], nil
}

func TestConcreteCustomDashboardRouteDetection(t *testing.T) {
	if !isConcreteCustomDashboardRoute("/custom-dashboard/dashboard-1") {
		t.Fatalf("expected concrete dashboard route to be detected")
	}
	if isConcreteCustomDashboardRoute("/custom-dashboard/:id") {
		t.Fatalf("generic router pattern must not be treated as a concrete dashboard menu item")
	}
	if isConcreteCustomDashboardRoute("/custom-dashboard/dashboard-1/edit") {
		t.Fatalf("edit route must not be treated as a dashboard menu item")
	}
}

func TestVisibleAllowedMenuItemIDsExcludesFilteredItems(t *testing.T) {
	ids := visibleAllowedMenuItemIDs(
		[]database.MenuItem{{ItemID: 1}, {ItemID: 3}},
		map[int]struct{}{1: {}, 2: {}, 3: {}},
	)
	if len(ids) != 2 || ids[0] != 1 || ids[1] != 3 {
		t.Fatalf("expected only visible allowed item IDs, got %#v", ids)
	}
}

func TestVisibleAllowedMenuItemIDsIncludesSharedCustomDashboardChildren(t *testing.T) {
	ids := visibleAllowedMenuItemIDs(
		[]database.MenuItem{
			{ItemID: 1, Key: "customDashboard"},
			{ItemID: 2, Router: &database.Router{RouterTo: "/custom-dashboard/shared"}},
		},
		map[int]struct{}{1: {}},
	)
	if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("expected root dashboard menu and visible shared dashboard child, got %#v", ids)
	}
}

func TestCustomDashboardMenuVisibility(t *testing.T) {
	ctx := core.EmptyCtx()
	repo := &customDashboardMenuRepoStub{teams: map[int64][]int64{100: {7}, 101: {8}}}

	if !customDashboardVisibleToUser(ctx, repo, &database.CustomDashboard{OwnerID: 100, Visibility: "personal"}, 100) {
		t.Fatalf("owner should see personal dashboard menu item")
	}
	if customDashboardVisibleToUser(ctx, repo, &database.CustomDashboard{OwnerID: 100, Visibility: "personal"}, 101) {
		t.Fatalf("non-owner should not see personal dashboard menu item")
	}
	if !customDashboardVisibleToUser(ctx, repo, &database.CustomDashboard{OwnerID: 100, Visibility: "everyone"}, 101) {
		t.Fatalf("everyone-visible dashboard menu item should be visible")
	}
	if !customDashboardVisibleToUser(ctx, repo, &database.CustomDashboard{OwnerID: 100, Visibility: "team", TeamIDs: "[7]"}, 100) {
		t.Fatalf("team member should see team dashboard menu item")
	}
	if customDashboardVisibleToUser(ctx, repo, &database.CustomDashboard{OwnerID: 100, Visibility: "team", TeamIDs: "[7]"}, 101) {
		t.Fatalf("non-team member should not see team dashboard menu item")
	}
}

func TestFilterCustomDashboardMenuItemsBackfillsLabelFromDashboardName(t *testing.T) {
	ctx := core.EmptyCtx()
	repo := &customDashboardMenuRepoStub{
		dashboards: map[string]*database.CustomDashboard{
			"dashboard-1": {Name: "运行时新建大盘", OwnerID: 100, Visibility: "everyone"},
		},
	}
	s := &service{dbRepo: repo}

	items := s.filterCustomDashboardMenuItems(ctx, []database.MenuItem{
		{
			ItemID: 2,
			Key:    "cd-dashboard1",
			Router: &database.Router{RouterTo: "/custom-dashboard/dashboard-1"},
		},
	}, 101)

	if len(items) != 1 {
		t.Fatalf("expected visible dashboard menu item, got %d", len(items))
	}
	if items[0].Label != "运行时新建大盘" {
		t.Fatalf("expected dashboard name to backfill missing menu label, got %q", items[0].Label)
	}
	if items[0].Abbreviation != "运行时新建大盘" {
		t.Fatalf("expected dashboard name to backfill missing menu abbreviation, got %q", items[0].Abbreviation)
	}
}
