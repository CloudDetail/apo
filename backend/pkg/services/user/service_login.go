package user

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) Login(req *request.LoginRequest) (response.LoginResponse, error) {
	err := s.dbRepo.Login(req.Username, req.Password)
	if err != nil {
		return response.LoginResponse{}, err
	}
	accessToken, refreshToken, err := middleware.GenerateTokens(req.Username)
	if err != nil {
		return response.LoginResponse{}, err
	}
	resp := response.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
}
