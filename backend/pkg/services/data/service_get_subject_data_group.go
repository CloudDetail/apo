package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetSubjectDataGroup(req *request.GetSubjectDataGroupRequest) (response.GetSubjectDataGroupResponse, error) {
	if req.SubjectType == model.DATA_GROUP_SUB_TYP_TEAM {
		return s.dbRepo.GetSubjectDataGroupList(req.SubjectID, req.SubjectType, req.Category)
	}

	return s.getUserDataGroup(req.SubjectID, req.Category)
}

// getUserDataGroup Get user's data group or default data group.
func (s *service) getUserDataGroup(userID int64, category string) ([]database.DataGroup, error) {
	teamIDs, err := s.dbRepo.GetUserTeams(userID)
	if err != nil {
		return nil, err
	}

	var groups []database.DataGroup
	for _, teamID := range teamIDs {
		gs, err := s.dbRepo.GetSubjectDataGroupList(teamID, model.DATA_GROUP_SUB_TYP_TEAM, category)
		if err != nil {
			return nil, err
		}

		groups = append(groups, gs...)
	}

	for i := range groups {
		groups[i].Source = model.DATA_GROUP_SUB_TYP_TEAM
	}

	gs, err := s.dbRepo.GetSubjectDataGroupList(userID, model.DATA_GROUP_SUB_TYP_USER, category)
	for i := range gs {
		gs[i].Source = model.DATA_GROUP_SUB_TYP_USER
	}

	if err != nil {
		return nil, err
	}

	groups = append(groups, gs...)

	// default data group which is all
	if len(groups) == 0 {
		datasource, err := s.GetDataSource()
		if err != nil {
			return nil, err
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

		defaultGroup := database.DataGroup{
			GroupName: "default",
			Source:    model.DATA_GROUP_SOURCE_DEFAULT,
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
		groups = append(groups, defaultGroup)
	}

	return groups, nil
}
