// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"unicode"
)

func (s *service) UpdateUserInfo(req *request.UpdateUserInfoRequest) error {
	//userRoles, err := s.dbRepo.GetUserRole(req.UserID)
	//if err != nil {
	//	return err
	//}
	//
	//roles, err := s.dbRepo.GetRoles(model.RoleFilter{})
	//if err != nil {
	//	return err
	//}

	//addRole, deleteRole, err := role.GetAddDeleteRoles(userRoles, req.RoleList, roles)
	//if err != nil {
	//	return err
	//}
	//
	//var grantFunc = func(ctx context.Context) error {
	//	return s.dbRepo.GrantRoleWithUser(ctx, req.UserID, addRole)
	//}
	//
	//var revokeFunc = func(ctx context.Context) error {
	//	return s.dbRepo.RevokeRole(ctx, req.UserID, deleteRole)
	//}

	var updateInfoFunc = func(ctx context.Context) error {
		return s.dbRepo.UpdateUserInfo(ctx, req.UserID, req.Phone, req.Email, req.Corporation)
	}

	return s.dbRepo.Transaction(context.Background(), updateInfoFunc)
}

func (s *service) UpdateUserPhone(req *request.UpdateUserPhoneRequest) error {
	return s.dbRepo.UpdateUserPhone(req.UserID, req.Phone)
}

func (s *service) UpdateUserEmail(req *request.UpdateUserEmailRequest) error {
	return s.dbRepo.UpdateUserEmail(req.UserID, req.Email)
}

func (s *service) UpdateUserPassword(req *request.UpdateUserPasswordRequest) error {
	if err := checkPasswordComplexity(req.NewPassword); err != nil {
		return err
	}
	return s.dbRepo.UpdateUserPassword(req.UserID, req.OldPassword, req.NewPassword)
}

func checkPasswordComplexity(password string) error {
	if len(password) < 8 {
		return model.NewErrWithMessage(errors.New("length less than 8"), code.UserPasswordSimpleError)
	}
	var (
		hasUpper     bool
		hasLower     bool
		hasDigit     bool
		hasSpecial   bool
		specialChars = "!@#$%^&*()-_+=<>?/{}[]|:;.,~`"
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char) || containsRune(specialChars, char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return model.NewErrWithMessage(errors.New("must contain at least one upper character"), code.UserPasswordSimpleError)
	}
	if !hasLower {
		return model.NewErrWithMessage(errors.New("must contain at least one lower character"), code.UserPasswordSimpleError)
	}
	if !hasDigit {
		return model.NewErrWithMessage(errors.New("must contain at least one digit"), code.UserPasswordSimpleError)
	}
	if !hasSpecial {
		return model.NewErrWithMessage(errors.New("must contain at least one special character"), code.UserPasswordSimpleError)
	}

	return nil
}

func containsRune(set string, char rune) bool {
	for _, r := range set {
		if r == char {
			return true
		}
	}
	return false
}

func (s *service) RestPassword(req *request.ResetPasswordRequest) error {
	if err := checkPasswordComplexity(req.NewPassword); err != nil {
		return err
	}
	return s.dbRepo.RestPassword(req.UserID, req.NewPassword)
}

func (s *service) UpdateSelfInfo(req *request.UpdateSelfInfoRequest) error {
	return s.dbRepo.UpdateUserInfo(context.Background(), req.UserID, req.Phone, req.Email, req.Corporation)
}
