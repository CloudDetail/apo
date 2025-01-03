// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
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

func (s *service) GetSubjectFeature(req *request.GetSubjectFeatureRequest) (resp response.GetSubjectFeatureResponse, err error) {
	if req.SubjectType == model.PERMISSION_SUB_TYP_USER {
		return s.getUserFeature(req.SubjectID, req.Language)
	}

	featureIDs, err := s.dbRepo.GetSubjectPermission(req.SubjectID, req.SubjectType, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return resp, err
	}

	featureList, err := s.dbRepo.GetFeature(featureIDs)
	if err != nil {
		return resp, err
	}
	err = s.dbRepo.GetFeatureTans(&featureList, req.Language)
	resp = featureList
	return resp, nil
}

func (s *service) PermissionOperation(req *request.PermissionOperationRequest) error {
	var exists bool
	var err error
	switch req.SubjectType {
	case model.PERMISSION_SUB_TYP_ROLE:
		exists, err = s.dbRepo.RoleExists(req.SubjectID)
	case model.PERMISSION_SUB_TYP_USER:
		exists, err = s.dbRepo.UserExists(req.SubjectID)
	}
	if err != nil {
		return err
	}
	if !exists {
		return model.NewErrWithMessage(errors.New("subject of authorisation does not exist"), code.AuthSubjectNotExistError)
	}

	addPermissions, deletePermissions, err := s.getAddAndDeletePermissions(req.SubjectID, req.SubjectType, req.Type, req.PermissionList)
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

func (s *service) getAddAndDeletePermissions(subID int64, subType, typ string, permList []int) (toAdd []int, toDelete []int, err error) {
	subPermissions, err := s.dbRepo.GetSubjectPermission(subID, subType, typ)
	if err != nil {
		return
	}

	permissions, err := s.dbRepo.GetFeature(nil)

	permissionMap := make(map[int]struct{})
	for _, permission := range permissions {
		permissionMap[permission.FeatureID] = struct{}{}
	}

	subPermissionMap := make(map[int]struct{})
	for _, id := range subPermissions {
		subPermissionMap[id] = struct{}{}
	}

	for _, permission := range permList {
		if _, exists := permissionMap[permission]; !exists {
			err = model.NewErrWithMessage(errors.New("permission does not exist"), code.PermissionNotExistError)
			return
		}
		if _, hasRole := subPermissionMap[permission]; !hasRole {
			toAdd = append(toAdd, permission)
		} else {
			delete(subPermissionMap, permission)
		}
	}

	for permission := range subPermissionMap {
		toDelete = append(toDelete, permission)
	}

	return
}
