package user

import "github.com/CloudDetail/apo/backend/pkg/model/response"

func (s *service) GetUserInfo(username string) (response.GetUserInfoResponse, error) {
	user, err := s.dbRepo.GetUserInfo(username)
	resp := response.GetUserInfoResponse{}
	if err != nil {
		return resp, err
	}
	resp.User = user
	return resp, nil
}
