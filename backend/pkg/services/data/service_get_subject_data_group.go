package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetSubjectDataGroup(req *request.GetSubjectDataGroupRequest) (response.GetSubjectDataGroupResponse, error) {
	var (
		resp response.GetSubjectDataGroupResponse
		err  error
	)
	if req.SubjectType == model.DATA_GROUP_SUB_TYP_TEAM {
		resp, err = s.dbRepo.GetSubjectDataGroupList(req.SubjectID, req.SubjectType)
	} else if req.SubjectType == model.DATA_GROUP_SUB_TYP_USER {
		resp, err = s.getUserDataGroup(req.SubjectID)
	}

	return resp, err
}

func (s *service) getUserDataGroup(userID int64) ([]database.DataGroup, error) {
	teamIDs, err := s.dbRepo.GetUserTeams(userID)
	if err != nil {
		return nil, err
	}

	var groups []database.DataGroup
	for _, teamID := range teamIDs {
		gs, err := s.dbRepo.GetSubjectDataGroupList(teamID, model.DATA_GROUP_SUB_TYP_TEAM)
		if err != nil {
			return nil, err
		}

		groups = append(groups, gs...)
	}

	for i := range groups {
		groups[i].Source = model.DATA_GROUP_SUB_TYP_TEAM
	}

	gs, err := s.dbRepo.GetSubjectDataGroupList(userID, model.DATA_GROUP_SUB_TYP_USER)
	for i := range gs {
		gs[i].Source = model.DATA_GROUP_SUB_TYP_USER
	}

	if err != nil {
		return nil, err
	}

	groups = append(groups, gs...)
	return groups, nil
}
