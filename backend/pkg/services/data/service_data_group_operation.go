// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) DataGroupOperation(ctx core.Context, req *request.DataGroupOperationRequest) error {
	var exists bool
	var err error
	switch req.SubjectType {
	case model.DATA_GROUP_SUB_TYP_TEAM:
		filter := model.TeamFilter{
			ID: req.SubjectID,
		}
		exists, err = s.dbRepo.TeamExist(ctx, filter)
	case model.DATA_GROUP_SUB_TYP_USER:
		exists, err = s.dbRepo.UserExists(ctx, req.SubjectID)
	default:
		err = core.Error(code.UnSupportedSubType, "unsupported subject type")
	}

	if err != nil {
		return err
	}

	if !exists {
		return core.Error(code.AuthSubjectNotExistError, "subject does not exist")
	}

	toModify, toDelete, err := s.dbRepo.GetModifyAndDeleteDataGroup(ctx, req.SubjectID, req.SubjectType, req.DataGroupPermission)
	if err != nil {
		return err
	}

	var assignDataGroupFunc = func(ctx core.Context) error {
		return s.dbRepo.AssignDataGroup(ctx, toModify)
	}

	var revokeDataGroupFunc = func(ctx core.Context) error {
		return s.dbRepo.RevokeDataGroupByGroup(ctx, toDelete, req.SubjectID)
	}

	return s.dbRepo.Transaction(ctx, assignDataGroupFunc, revokeDataGroupFunc)
}
