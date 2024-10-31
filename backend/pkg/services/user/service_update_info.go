package user

import "github.com/CloudDetail/apo/backend/pkg/model/request"

func (s *service) UpdateUserInfo(req *request.UpdateUserInfoRequest) error {
	return nil
}

func (s *service) UpdateUserPhone(username string, req *request.UpdateUserPhoneRequest) error {
	// TODO check vcode
	return s.dbRepo.UpdateUserPhone(username, req.Phone)
}

func (s *service) UpdateUserEmail(username string, req *request.UpdateUserEmailRequest) error {
	// TODO check vcode
	return s.dbRepo.UpdateUserEmail(username, req.Email)
}

func (s *service) UpdateUserPassword(username string, req *request.UpdateUserPasswordRequest) error {
	return s.dbRepo.UpdateUserPassword(username, req.OldPassword, req.NewPassword)
}
