// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"strings"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (s *service) CheckRouterPermission(ctx core.Context, userID int64, routerTo string) (bool, error) {
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
		if checkRouterMatch(routerTo, r) {
			return true, nil
		}
	}

	return false, nil
}

func checkRouterMatch(checkRouter, matchRouter string) bool {
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
