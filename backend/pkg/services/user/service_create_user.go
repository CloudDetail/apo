// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"net/mail"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func checkUserName(username string) error {
	_, err := mail.ParseAddress(username + "@apo.com")
	if err != nil {
		return core.Error(code.UserNameError, "username format invaild")
	}
	return nil
}

func (s *service) CreateUser(ctx core.Context, req *request.CreateUserRequest) error {
	if err := checkUserName(req.Username); err != nil {
		return err
	}

	if err := checkPasswordComplexity(req.Password); err != nil {
		return err
	}

	user := &profile.User{
		UserID:      util.Generator.GenerateID(),
		Username:    req.Username,
		Password:    req.Password,
		Corporation: req.Corporation,
		Email:       req.Email,
		Phone:       req.Phone,
	}

	if len(req.RoleList) == 0 {
		filter := model.RoleFilter{
			Name: model.ROLE_ADMIN,
		}
		roles, err := s.dbRepo.GetRoles(ctx, filter)
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
		roles, err := s.dbRepo.GetRoles(ctx, filter)
		if err != nil {
			return err
		}
		if len(roles) != len(req.RoleList) {
			return core.Error(code.RoleNotExistsError, "role does not exist")
		}
	}

	// Check if the team exists
	if len(req.TeamList) > 0 {
		filter := model.TeamFilter{
			IDs: req.TeamList,
		}
		exist, err := s.dbRepo.TeamExist(ctx, filter)
		if err != nil {
			return err
		}
		if !exist {
			return core.Error(code.TeamNotExistError, "team does not exist")
		}
	}

	var assignTeamFunc = func(ctx core.Context) error {
		return s.dbRepo.AssignUserToTeam(ctx, user.UserID, req.TeamList)
	}

	var createDifyUserFunc = func(ctx core.Context) error {
		resp, err := s.difyRepo.AddUser(req.Username, req.Password, "admin")
		if err != nil || resp.Result != "success" {
			return errors.New("failed to create user in dify")
		}
		return nil
	}

	var createUserFunc = func(ctx core.Context) error {
		return s.dbRepo.CreateUser(ctx, user)
	}

	var grantRoleFunc = func(ctx core.Context) error {
		return s.dbRepo.GrantRoleWithUser(ctx, user.UserID, req.RoleList)
	}

	var assignGroupFunc = func(ctx core.Context) error {
		// TODO Check permission
		authGroups := make([]database.AuthDataGroup, 0, len(req.GroupIDs))
		for _, groupId := range req.GroupIDs {
			authGroups = append(authGroups, database.AuthDataGroup{
				SubjectID:   user.UserID,
				SubjectType: model.DATA_GROUP_SUB_TYP_USER,
				GroupID:     groupId,
				Type:        "view",
			})
		}
		return s.dbRepo.AssignDataGroup(ctx, authGroups)
	}

	return s.dbRepo.Transaction(ctx,
		createUserFunc,
		createDifyUserFunc,
		grantRoleFunc,
		assignTeamFunc,
		assignGroupFunc,
	)
}
