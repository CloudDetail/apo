// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) CreateUser(req *request.CreateUserRequest) error {
	if err := checkPasswordComplexity(req.Password); err != nil {
		return err
	}

	user := &database.User{
		UserID:      util.Generator.GenerateID(),
		Username:    req.Username,
		Password:    req.Password,
		Corporation: req.Corporation,
		Email:       req.Email,
		Phone:       req.Phone,
	}

	if len(req.RoleList) == 0 {
		filter := model.RoleFilter{
			Name: model.ROLE_VIEWER,
		}
		roles, err := s.dbRepo.GetRoles(filter)
		if err != nil {
			return err
		}

		if len(roles) > 0 {
			req.RoleList = append(req.RoleList, roles[0].RoleID)
		}
	} else {
		filter := model.RoleFilter{
			IDs: req.RoleList,
		}
		roles, err := s.dbRepo.GetRoles(filter)
		if err != nil {
			return err
		}
		if len(roles) != len(req.RoleList) {
			return model.NewErrWithMessage(errors.New("role does not exist"), code.RoleNotExistsError)
		}
	}

	var createUserFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateUser(ctx, user)
	}

	var grantRoleFunc = func(ctx context.Context) error {
		return s.dbRepo.GrantRole(ctx, user.UserID, req.RoleList)
	}

	return s.dbRepo.Transaction(context.Background(), createUserFunc, grantRoleFunc)
}
