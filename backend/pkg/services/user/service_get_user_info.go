package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetUserInfo(username string) (response.GetUserInfoResponse, error) {
	user, err := s.dbRepo.GetUserInfo(username)
	resp := response.GetUserInfoResponse{}
	if err != nil {
		return resp, err
	}
	resp.User = user
	return resp, nil
}

func (s *service) GetUserList(req *request.GetUserListRequest) (response.GetUserListResponse, error) {
	users, err := s.dbRepo.GetUserList(req)
	resp := response.GetUserListResponse{}
	if err != nil {
		return resp, err
	}
	resp.Users = users
	return resp, nil
}
