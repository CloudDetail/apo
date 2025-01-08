// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package permission

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

func (s *service) GetSubjectFeature(req *request.GetSubjectFeatureRequest) (response.GetSubjectFeatureResponse, error) {
	if req.SubjectType == model.PERMISSION_SUB_TYP_USER {
		featureIDs, err := s.getUserFeatureIDs(req.SubjectID)
		if err != nil {
			return nil, err
		}

		features, err := s.dbRepo.GetFeature(featureIDs)
		if err != nil {
			return nil, err
		}
		err = s.dbRepo.GetFeatureTans(&features, req.Language)
		return features, err
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

func (s *service) PermissionOperation(req *request.PermissionOperationRequest) error {
	var exists bool
	var err error
	switch req.SubjectType {
	case model.PERMISSION_SUB_TYP_ROLE:
		exists, err = s.dbRepo.RoleExists(int(req.SubjectID))
	case model.PERMISSION_SUB_TYP_USER:
		exists, err = s.dbRepo.UserExists(req.SubjectID)
	}
	if err != nil {
		return err
	}
	if !exists {
		return model.NewErrWithMessage(errors.New("subject of authorisation does not exist"), code.AuthSubjectNotExistError)
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
