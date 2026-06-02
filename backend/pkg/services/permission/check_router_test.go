// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"strings"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"gorm.io/gorm"
)

func TestCheckRouterMatch(t *testing.T) {
	tests := []struct {
		name         string
		checkRouter  string
		matchRouter  string
		want         bool
		errorMessage string
	}{
		{
			name:         "完全匹配的静态路由",
			checkRouter:  "/user/profile",
			matchRouter:  "/user/profile",
			want:         true,
			errorMessage: "静态路由应完全匹配",
		},
		{
			name:         "带动态参数的相同结构",
			checkRouter:  "/user/1",
			matchRouter:  "/user/:id",
			want:         true,
			errorMessage: "相同动态参数结构应匹配",
		},
		{
			name:         "路径长度不同",
			checkRouter:  "/user/123/profile",
			matchRouter:  "/user/:id",
			want:         false,
			errorMessage: "路径层级深度不同应不匹配",
		},
		{
			name:         "静态部分不匹配",
			checkRouter:  "/admin/profile",
			matchRouter:  "/user/profile",
			want:         false,
			errorMessage: "静态路径部分不匹配应返回失败",
		},
		{
			name:         "动态参数位置不匹配",
			checkRouter:  "/user/1/contacts",
			matchRouter:  "/user/profile/:id",
			want:         false,
			errorMessage: "动态参数位置不同应不匹配",
		},

		{
			name:         "根路径匹配",
			checkRouter:  "/",
			matchRouter:  "/",
			want:         true,
			errorMessage: "根路径应正确匹配",
		},
		{
			name:         "空路径处理",
			checkRouter:  "",
			matchRouter:  "",
			want:         true,
			errorMessage: "空路径应视为匹配",
		},
		{
			name:         "混合大小写路径",
			checkRouter:  "/User/Profile",
			matchRouter:  "/user/profile",
			want:         false,
			errorMessage: "应区分路径大小写",
		},

		{
			name:         "多级动态参数",
			checkRouter:  "/api/1/users/2",
			matchRouter:  "/api/:version/users/:id",
			want:         true,
			errorMessage: "多级动态参数结构应匹配",
		},
		{
			name:         "嵌套动态参数",
			checkRouter:  "/blog/2003/2/1",
			matchRouter:  "/blog/:year/:month/:slug",
			want:         true,
			errorMessage: "嵌套动态参数应匹配",
		},

		{
			name:         "含特殊字符的静态路径",
			checkRouter:  "/user/100%",
			matchRouter:  "/user/100%",
			want:         true,
			errorMessage: "特殊字符路径应精确匹配",
		},
		{
			name:         "含点的路径段",
			checkRouter:  "/file/v1.2.3",
			matchRouter:  "/file/v1.2.3",
			want:         true,
			errorMessage: "含点路径应精确匹配",
		},
		{
			name:         "前置/",
			checkRouter:  "/user/list",
			matchRouter:  "user/list",
			want:         true,
			errorMessage: "前置/需要忽略",
		},
		{
			name:         "后置/",
			checkRouter:  "user/list/",
			matchRouter:  "user/list",
			want:         true,
			errorMessage: "后置/需要忽略",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkRouterMatch(tt.checkRouter, tt.matchRouter)

			if got != tt.want {
				partsCheck := strings.Split(tt.checkRouter, "/")
				partsMatch := strings.Split(tt.matchRouter, "/")

				t.Errorf(
					"%s\nCheck路由: %s\nMatch路由: %s\n路径分解:\nCheck: %#v\nMatch: %#v\n预期: %v\n实际: %v",
					tt.errorMessage,
					tt.checkRouter,
					tt.matchRouter,
					partsCheck,
					partsMatch,
					tt.want,
					got,
				)
			}
		})
	}
}

func TestConcreteCustomDashboardRouteRequiresExactAuthorization(t *testing.T) {
	if authorizedRouterMatches("/custom-dashboard/dashboard-1", "/custom-dashboard/:id") {
		t.Fatalf("concrete custom dashboard route must not be authorized by generic dashboard route")
	}
	if authorizedRouterMatches("/custom-dashboard/dashboard-1/edit", "/custom-dashboard/:id/edit") {
		t.Fatalf("concrete custom dashboard edit route must not be authorized by generic dashboard edit route")
	}
	if !authorizedRouterMatches("/custom-dashboard/dashboard-1", "/custom-dashboard/dashboard-1") {
		t.Fatalf("concrete custom dashboard route should accept exact authorization")
	}
	if !authorizedRouterMatches("/custom-dashboard/dashboard-1/edit", "/custom-dashboard/dashboard-1/edit") {
		t.Fatalf("concrete custom dashboard edit route should accept exact authorization")
	}
	if !authorizedRouterMatches("/service/1", "/service/:id") {
		t.Fatalf("non-dashboard dynamic routes should keep existing matching behavior")
	}
}

type dashboardRouteVisibilityRepoStub struct {
	database.Repo
	userPerms     []int
	rolePerms     []database.AuthPermission
	teamPerms     []database.AuthPermission
	teams         map[int64][]int64
	routers       []database.Router
	itemRoutes    []database.Router
	dashboards    map[string]*database.CustomDashboard
	dashboardErrs map[string]error
}

func (r *dashboardRouteVisibilityRepoStub) GetUserRole(ctx core.Context, userID int64) ([]profile.UserRole, error) {
	return nil, nil
}

func (r *dashboardRouteVisibilityRepoStub) GetUserTeams(ctx core.Context, userID int64) ([]int64, error) {
	return r.teams[userID], nil
}

func (r *dashboardRouteVisibilityRepoStub) GetSubjectsPermission(ctx core.Context, subIDs []int64, subType string, typ string) ([]database.AuthPermission, error) {
	if subType == model.PERMISSION_SUB_TYP_TEAM {
		return r.teamPerms, nil
	}
	return r.rolePerms, nil
}

func (r *dashboardRouteVisibilityRepoStub) GetSubjectPermission(ctx core.Context, subID int64, subType string, typ string) ([]int, error) {
	return r.userPerms, nil
}

func (r *dashboardRouteVisibilityRepoStub) GetFeatureMappingByFeature(ctx core.Context, featureIDs []int, mappedType string) ([]database.FeatureMapping, error) {
	if mappedType == model.MAPPED_TYP_ROUTER {
		return []database.FeatureMapping{{MappedID: 1, MappedType: mappedType}}, nil
	}
	return []database.FeatureMapping{{MappedID: 2, MappedType: mappedType}}, nil
}

func (r *dashboardRouteVisibilityRepoStub) GetRouterByIDs(ctx core.Context, routerIDs []int) ([]database.Router, error) {
	return r.routers, nil
}

func (r *dashboardRouteVisibilityRepoStub) GetItemsRouter(ctx core.Context, itemIDs []int) ([]database.Router, error) {
	return r.itemRoutes, nil
}

func (r *dashboardRouteVisibilityRepoStub) GetCustomDashboard(ctx core.Context, id string) (*database.CustomDashboard, error) {
	if err := r.dashboardErrs[id]; err != nil {
		return nil, err
	}
	return r.dashboards[id], nil
}

func TestCheckRouterPermissionAllowsVisibleCustomDashboardWithRootPermission(t *testing.T) {
	repo := &dashboardRouteVisibilityRepoStub{
		userPerms:  []int{7},
		teams:      map[int64][]int64{101: {8}},
		routers:    []database.Router{{RouterTo: "/custom-dashboard/:id"}},
		itemRoutes: []database.Router{{RouterTo: "/custom-dashboard"}},
		dashboards: map[string]*database.CustomDashboard{
			"shared":  {OwnerID: 100, Visibility: "everyone"},
			"team":    {OwnerID: 100, Visibility: "team", TeamIDs: "[8]"},
			"private": {OwnerID: 100, Visibility: "personal"},
		},
	}
	s := &service{dbRepo: repo}
	ctx := core.EmptyCtx()

	if ok, err := s.CheckRouterPermission(ctx, 101, "/custom-dashboard/shared"); err != nil || !ok {
		t.Fatalf("everyone-visible dashboard should allow concrete route with root dashboard permission, ok=%v err=%v", ok, err)
	}
	if ok, err := s.CheckRouterPermission(ctx, 101, "/custom-dashboard/team"); err != nil || !ok {
		t.Fatalf("team-visible dashboard should allow concrete route for a team member, ok=%v err=%v", ok, err)
	}
	if ok, err := s.CheckRouterPermission(ctx, 101, "/custom-dashboard/private"); err != nil || ok {
		t.Fatalf("private dashboard should not allow non-owner concrete route, ok=%v err=%v", ok, err)
	}
}

func TestCheckRouterPermissionAllowsMissingCustomDashboardWithRootPermission(t *testing.T) {
	repo := &dashboardRouteVisibilityRepoStub{
		userPerms:     []int{7},
		routers:       []database.Router{{RouterTo: "/custom-dashboard/:id"}},
		itemRoutes:    []database.Router{{RouterTo: "/custom-dashboard"}},
		dashboards:    map[string]*database.CustomDashboard{},
		dashboardErrs: map[string]error{"deleted": gorm.ErrRecordNotFound},
	}
	s := &service{dbRepo: repo}

	if ok, err := s.CheckRouterPermission(core.EmptyCtx(), 101, "/custom-dashboard/deleted"); err != nil || !ok {
		t.Fatalf("missing dashboard route should reach viewer so it can show not-found state, ok=%v err=%v", ok, err)
	}
}

func TestCheckRouterPermissionRequiresOwnerForCustomDashboardEditRoute(t *testing.T) {
	repo := &dashboardRouteVisibilityRepoStub{
		userPerms:  []int{7},
		routers:    []database.Router{{RouterTo: "/custom-dashboard/:id/edit"}},
		itemRoutes: []database.Router{{RouterTo: "/custom-dashboard"}},
		dashboards: map[string]*database.CustomDashboard{
			"shared": {OwnerID: 100, Visibility: "everyone"},
		},
	}
	s := &service{dbRepo: repo}

	if ok, err := s.CheckRouterPermission(core.EmptyCtx(), 101, "/custom-dashboard/shared/edit"); err != nil || ok {
		t.Fatalf("non-owner must not edit a shared custom dashboard, ok=%v err=%v", ok, err)
	}
	if ok, err := s.CheckRouterPermission(core.EmptyCtx(), 100, "/custom-dashboard/shared/edit"); err != nil || !ok {
		t.Fatalf("owner should be allowed to edit own custom dashboard, ok=%v err=%v", ok, err)
	}
}
