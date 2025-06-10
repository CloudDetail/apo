// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetSubjectDataGroup(ctx core.Context, req *request.GetSubjectDataGroupRequest) (response.GetSubjectDataGroupResponse, error) {
	if req.SubjectType == model.DATA_GROUP_SUB_TYP_TEAM {
		return s.dbRepo.GetSubjectDataGroupList(ctx, req.SubjectID, req.SubjectType, req.Category)
	}

	return s.getUserDataGroup(ctx, req.SubjectID, req.Category)
}

// getUserDataGroup Get user's data group or default data group.
func (s *service) getUserDataGroup(ctx core.Context, userID int64, category string) ([]database.DataGroup, error) {
	teamIDs, err := s.dbRepo.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, err
	}

	seen := make(map[int64]struct{})
	// Get user's teams.
	var groups []database.DataGroup
	for _, teamID := range teamIDs {
		gs, err := s.dbRepo.GetSubjectDataGroupList(ctx, teamID, model.DATA_GROUP_SUB_TYP_TEAM, category)
		if err != nil {
			return nil, err
		}

		for _, g := range gs {
			if _, ok := seen[g.GroupID]; ok {
				continue
			}

			seen[g.GroupID] = struct{}{}
			groups = append(groups, g)
		}
	}

	for i := range groups {
		groups[i].Source = model.DATA_GROUP_SUB_TYP_TEAM
	}

	gs, err := s.dbRepo.GetSubjectDataGroupList(ctx, userID, model.DATA_GROUP_SUB_TYP_USER, category)
	for i := range gs {
		gs[i].Source = model.DATA_GROUP_SUB_TYP_USER
	}

	if err != nil {
		return nil, err
	}

	groups = append(groups, gs...)

	return groups, nil
}

func (s *service) getDefaultDataGroup(ctx core.Context, category string) (database.DataGroup, error) {
	defaultGroup := database.DataGroup{
		GroupName: "default",
		Source:    model.DATA_GROUP_SOURCE_DEFAULT,
	}

	datasource, err := s.GetDataSource(ctx)
	if err != nil {
		return defaultGroup, err
	}

	filteredSources := make([]model.Datasource, 0)
	for _, ds := range datasource.NamespaceList {
		if category == "" || ds.Category == category {
			filteredSources = append(filteredSources, ds)
		}
	}

	for _, ds := range datasource.ServiceList {
		if category == "" || ds.Category == category {
			filteredSources = append(filteredSources, ds)
		}
	}

	items := make([]database.DatasourceGroup, 0, len(filteredSources))
	for _, ds := range filteredSources {
		items = append(items, database.DatasourceGroup{
			Datasource: ds.Datasource,
			Type:       ds.Type,
			Category:   ds.Category,
		})
	}

	defaultGroup.DatasourceList = items
	return defaultGroup, nil
}
