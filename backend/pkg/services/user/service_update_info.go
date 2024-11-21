package user

import "github.com/CloudDetail/apo/backend/pkg/model/request"

func (s *service) UpdateUserInfo(username string, req *request.UpdateUserInfoRequest) error {
	return s.dbRepo.UpdateUserInfo(username, req)
}

func (s *service) UpdateUserPhone(username string, req *request.UpdateUserPhoneRequest) error {
	return s.dbRepo.UpdateUserPhone(username, req.Phone)
}

func (s *service) UpdateUserEmail(username string, req *request.UpdateUserEmailRequest) error {
	return s.dbRepo.UpdateUserEmail(username, req.Email)
}

func (s *service) UpdateUserPassword(username string, req *request.UpdateUserPasswordRequest) error {
	return s.dbRepo.UpdateUserPassword(username, req.OldPassword, req.NewPassword)
}
