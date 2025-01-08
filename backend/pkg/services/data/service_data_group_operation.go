package data

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) DataGroupOperation(req *request.DataGroupOperationRequest) error {
	var exists bool
	var err error
	switch req.SubjectType {
	case model.DATA_GROUP_SUB_TYP_TEAM:
		exists, err = s.dbRepo.TeamExist(req.SubjectID)
	case model.DATA_GROUP_SUB_TYP_USER:
		exists, err = s.dbRepo.UserExists(req.SubjectID)
	default:
		err = model.NewErrWithMessage(errors.New("unsupported subject type"), code.UnSupportedSubType)
	}

	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("subject does not exist"), code.AuthSubjectNotExistError)
	}

	toModify, toDelete, err := s.dbRepo.GetModifyAndDeleteDataGroup(req.SubjectID, req.SubjectType, req.DataGroupPermission)
	if err != nil {
		return err
	}

	var assignDataGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.AssignDataGroup(ctx, toModify)
	}

	var revokeDataGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokeDataGroup(ctx, toDelete)
	}

	return s.dbRepo.Transaction(context.Background(), assignDataGroupFunc, revokeDataGroupFunc)
}
