package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"unicode"
)

func (s *service) UpdateUserInfo(req *request.UpdateUserInfoRequest) error {
	return s.dbRepo.UpdateUserInfo(req)
}

func (s *service) UpdateUserPhone(req *request.UpdateUserPhoneRequest) error {
	return s.dbRepo.UpdateUserPhone(req.Username, req.Phone)
}

func (s *service) UpdateUserEmail(req *request.UpdateUserEmailRequest) error {
	return s.dbRepo.UpdateUserEmail(req.Username, req.Email)
}

func (s *service) UpdateUserPassword(req *request.UpdateUserPasswordRequest) error {
	if err := checkPasswordComplexity(req.NewPassword); err != nil {
		return err
	}
	return s.dbRepo.UpdateUserPassword(req.Username, req.OldPassword, req.NewPassword)
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
