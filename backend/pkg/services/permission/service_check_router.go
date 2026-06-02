// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"errors"
	"strings"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
)

var allowedRouters = []string{
	"/agent-dataplane",
	"agent",
	"probe-installer",
}

func (s *service) CheckRouterPermission(ctx core.Context, userID int64, routerTo string) (bool, error) {
	for _, allowedRouter := range allowedRouters {
		if checkRouterMatch(routerTo, allowedRouter) {
			return true, nil
		}
	}

	features, err := s.getUserFeatureIDs(ctx, userID)
	if err != nil {
		return false, err
	}

	menuMappings, err := s.dbRepo.GetFeatureMappingByFeature(ctx, features, model.MAPPED_TYP_MENU)
	if err != nil {
		return false, err
	}

	routerMappings, err := s.dbRepo.GetFeatureMappingByFeature(ctx, features, model.MAPPED_TYP_ROUTER)
	if err != nil {
		return false, err
	}

	menuIDs := make([]int, len(menuMappings))
	for i := range menuMappings {
		menuIDs[i] = menuMappings[i].MappedID
	}

	routerIDs := make([]int, 0, len(routerMappings))
	for _, mapping := range routerMappings {
		routerIDs = append(routerIDs, mapping.MappedID)
	}

	routers, err := s.dbRepo.GetRouterByIDs(ctx, routerIDs)
	if err != nil {
		return false, err
	}

	itemRouters, err := s.dbRepo.GetItemsRouter(ctx, menuIDs)
	if err != nil {
		return false, err
	}

	authRouters := []string{}
	for _, r := range itemRouters {
		authRouters = append(authRouters, r.RouterTo)
	}

	for _, r := range routers {
		authRouters = append(authRouters, r.RouterTo)
	}

	for _, r := range authRouters {
		if authorizedRouterMatches(routerTo, r) {
			return true, nil
		}
	}

	return s.visibleCustomDashboardRouteAuthorized(ctx, userID, routerTo, authRouters)
}

func authorizedRouterMatches(checkRouter, authRouter string) bool {
	if isConcreteCustomDashboardRoute(checkRouter) || isConcreteCustomDashboardEditRoute(checkRouter) {
		return checkRouter == authRouter
	}
	return checkRouterMatch(checkRouter, authRouter)
}

func checkRouterMatch(checkRouter, matchRouter string) bool {
	checkRouter = strings.Trim(checkRouter, "/")
	matchRouter = strings.Trim(matchRouter, "/")
	checkParts := strings.Split(checkRouter, "/")
	matchParts := strings.Split(matchRouter, "/")

	if len(checkParts) != len(matchParts) {
		return false
	}

	for i := range matchParts {
		if strings.HasPrefix(matchParts[i], ":") {
			continue
		}

		if checkParts[i] != matchParts[i] {
			return false
		}
	}
	return true
}

func (s *service) visibleCustomDashboardRouteAuthorized(ctx core.Context, userID int64, routerTo string, authRouters []string) (bool, error) {
	isViewRoute := isConcreteCustomDashboardRoute(routerTo)
	isEditRoute := isConcreteCustomDashboardEditRoute(routerTo)
	if (!isViewRoute && !isEditRoute) ||
		(isViewRoute && !hasCustomDashboardRootRoutePermission(authRouters)) ||
		(isEditRoute && !hasCustomDashboardEditRoutePermission(authRouters)) {
		return false, nil
	}
	id := strings.TrimPrefix(routerTo, "/custom-dashboard/")
	id = strings.TrimSuffix(id, "/edit")
	row, err := s.dbRepo.GetCustomDashboard(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return isViewRoute, nil
	}
	if err != nil || row == nil {
		return false, err
	}
	if isEditRoute {
		return row.OwnerID == userID, nil
	}
	return customDashboardVisibleToUser(ctx, s.dbRepo, row, userID), nil
}

func hasCustomDashboardRootRoutePermission(authRouters []string) bool {
	for _, route := range authRouters {
		if route == "/custom-dashboard" || route == "/custom-dashboard/:id" {
			return true
		}
	}
	return false
}

func hasCustomDashboardEditRoutePermission(authRouters []string) bool {
	for _, route := range authRouters {
		if route == "/custom-dashboard/:id/edit" {
			return true
		}
	}
	return false
}
