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
	if req.PageParam == nil {
		req.CurrentPage = 1
		req.PageSize = 10
	}
	users, count, err := s.dbRepo.GetUserList(req)
	resp := response.GetUserListResponse{}
	if err != nil {
		return resp, err
	}
	resp.Users = users
	resp.PageSize = req.PageSize
	resp.CurrentPage = req.CurrentPage
	resp.Total = count
	return resp, nil
}
