// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetFeature(req *request.GetFeatureRequest) (response.GetFeatureResponse, error) {
	features, err := s.dbRepo.GetFeature(nil)
	if err != nil {
		return nil, err
	}

	err = s.dbRepo.GetFeatureTans(&features, req.Language)
	if err != nil {
		return nil, err
	}

	featureMap := make(map[int]*database.Feature)
	var rootFeatures []*database.Feature

	for _, feature := range features {
		f := feature
		featureMap[f.FeatureID] = &f
	}

	for _, feature := range features {
		if feature.ParentID == nil {
			rootFeatures = append(rootFeatures, featureMap[feature.FeatureID])
		} else {
			if parent, exists := featureMap[*feature.ParentID]; exists {
				parent.Children = append(parent.Children, *featureMap[feature.FeatureID])
			}
		}
	}

	return rootFeatures, nil
}

func (s *service) GetSubjectFeature(req *request.GetSubjectFeatureRequest) (response.GetSubjectFeatureResponse, error) {
	if req.SubjectType == model.PERMISSION_SUB_TYP_USER {
		return s.getUserFeatureWithSource(req.SubjectID, req.Language)
	}

	featureIDs, err := s.dbRepo.GetSubjectPermission(req.SubjectID, req.SubjectType, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}

	featureList, err := s.dbRepo.GetFeature(featureIDs)
	if err != nil {
		return nil, err
	}
	err = s.dbRepo.GetFeatureTans(&featureList, req.Language)
	return featureList, nil
}

func (s *service) getUserFeatureWithSource(userID int64, language string) (response.GetSubjectFeatureResponse, error) {
	// Get user's roles and teams
	roles, err := s.dbRepo.GetUserRole(userID)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]int64, len(roles))
	for i := range roleIDs {
		roleIDs[i] = int64(roles[i].RoleID)
	}

	teamIDs, err := s.dbRepo.GetUserTeams(userID)
	if err != nil {
		return nil, err
	}

	// Get feature sources
	roleFeatures, err := s.dbRepo.GetSubjectsPermission(roleIDs, model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}
	teamFeatures, err := s.dbRepo.GetSubjectsPermission(teamIDs, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}
	userFeatures, err := s.dbRepo.GetSubjectPermission(userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}

	features, err := s.dbRepo.GetFeature(nil)
	featureMap := make(map[int]database.Feature)
	for _, f := range features {
		featureMap[f.FeatureID] = f
	}

	resp := make([]database.Feature, 0, len(roleFeatures)+len(teamFeatures)+len(userFeatures))
	for _, rf := range roleFeatures {
		f, ok := featureMap[rf.PermissionID]
		f.Source = model.PERMISSION_SUB_TYP_ROLE
		if ok {
			resp = append(resp, f)
		}
	}

	for _, rf := range teamFeatures {
		f, ok := featureMap[rf.PermissionID]
		f.Source = model.PERMISSION_SUB_TYP_TEAM
		if ok {
			resp = append(resp, f)
		}
	}

	for _, id := range userFeatures {
		f, ok := featureMap[id]
		f.Source = model.PERMISSION_SUB_TYP_USER
		if ok {
			resp = append(resp, f)
		}
	}

	err = s.dbRepo.GetFeatureTans(&features, language)
	return resp, err
}

func (s *service) PermissionOperation(req *request.PermissionOperationRequest) error {
	var exists bool
	var err error
	switch req.SubjectType {
	case model.PERMISSION_SUB_TYP_ROLE:
		exists, err = s.dbRepo.RoleExists(int(req.SubjectID))
	case model.PERMISSION_SUB_TYP_USER:
		exists, err = s.dbRepo.UserExists(req.SubjectID)
	case model.PERMISSION_SUB_TYP_TEAM:
		filter := model.TeamFilter{
			ID: req.SubjectID,
		}
		exists, err = s.dbRepo.TeamExist(filter)
	default:
		return nil
	}
	if err != nil {
		return err
	}
	if !exists {
		return core.Error(code.AuthSubjectNotExistError, "subject of authorisation does not exist")
	}

	addPermissions, deletePermissions, err := s.dbRepo.GetAddAndDeletePermissions(req.SubjectID, req.SubjectType, req.Type, req.PermissionList)
	if err != nil {
		return err
	}
	var grantFunc = func(ctx context.Context) error {
		if len(addPermissions) == 0 {
			return nil
		}
		return s.dbRepo.GrantPermission(ctx, req.SubjectID, req.SubjectType, req.Type, addPermissions)
	}

	var revokeFunc = func(ctx context.Context) error {
		if len(deletePermissions) == 0 {
			return nil
		}
		return s.dbRepo.RevokePermission(ctx, req.SubjectID, req.SubjectType, req.Type, deletePermissions)
	}

	return s.dbRepo.Transaction(context.Background(), grantFunc, revokeFunc)
}
