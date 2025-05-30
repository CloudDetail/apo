// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"errors"
	"unicode"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) UpdateUserInfo(ctx core.Context, req *request.UpdateUserInfoRequest) error {
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
	//var grantFunc = func(ctx core.Context) error {
	//	return s.dbRepo.GrantRoleWithUser(ctx, req.UserID, addRole)
	//}
	//
	//var revokeFunc = func(ctx core.Context) error {
	//	return s.dbRepo.RevokeRole(ctx, req.UserID, deleteRole)
	//}

	var updateInfoFunc = func(ctx core.Context) error {
		return s.dbRepo.UpdateUserInfo(ctx, req.UserID, req.Phone, req.Email, req.Corporation)
	}

	return s.dbRepo.Transaction(ctx, updateInfoFunc)
}

func (s *service) UpdateUserPhone(ctx core.Context, req *request.UpdateUserPhoneRequest) error {
	return s.dbRepo.UpdateUserPhone(ctx, req.UserID, req.Phone)
}

func (s *service) UpdateUserEmail(ctx core.Context, req *request.UpdateUserEmailRequest) error {
	return s.dbRepo.UpdateUserEmail(ctx, req.UserID, req.Email)
}

func (s *service) UpdateUserPassword(ctx core.Context, req *request.UpdateUserPasswordRequest) error {
	if err := checkPasswordComplexity(req.NewPassword); err != nil {
		return err
	}

	user, err := s.dbRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return err
	}

	var updatePasswordFunc = func(ctx core.Context) error {
		return s.dbRepo.UpdateUserPassword(ctx, req.UserID, req.OldPassword, req.NewPassword)
	}

	var updateDifyPasswordFunc = func(ctx core.Context) error {
		resp, err := s.difyRepo.UpdatePassword(user.Username, req.OldPassword, req.NewPassword)
		if err != nil || resp.Result != "success" {
			return errors.New("failed to update password in dify")
		}
		return nil
	}
	return s.dbRepo.Transaction(ctx, updatePasswordFunc, updateDifyPasswordFunc)
}

func checkPasswordComplexity(password string) error {
	if len(password) < 8 {
		return core.Error(code.UserPasswdSimpleError, "length less than 8")
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
		return core.Error(code.UserPasswdSimpleError, "must contain at least one upper character")
	}
	if !hasLower {
		return core.Error(code.UserPasswdSimpleError, "must contain at least one lower character")
	}
	if !hasDigit {
		return core.Error(code.UserPasswdSimpleError, "must contain at least one digit")
	}
	if !hasSpecial {
		return core.Error(code.UserPasswdSimpleError, "must contain at least one special character")
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

func (s *service) RestPassword(ctx core.Context, req *request.ResetPasswordRequest) error {
	if err := checkPasswordComplexity(req.NewPassword); err != nil {
		return err
	}

	user, err := s.dbRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return err
	}

	var resetPasswordFunc = func(ctx core.Context) error {
		return s.dbRepo.RestPassword(ctx, req.UserID, req.NewPassword)
	}

	var resetDifyPasswordFunc = func(ctx core.Context) error {
		resp, err := s.difyRepo.ResetPassword(user.Username, req.NewPassword)
		if err != nil || resp.Result != "success" {
			return errors.New("failed to reset password in dify")
		}
		return nil
	}
	return s.dbRepo.Transaction(ctx, resetPasswordFunc, resetDifyPasswordFunc)
}

func (s *service) UpdateSelfInfo(ctx core.Context, req *request.UpdateSelfInfoRequest) error {
	return s.dbRepo.UpdateUserInfo(ctx, req.UserID, req.Phone, req.Email, req.Corporation)
}
