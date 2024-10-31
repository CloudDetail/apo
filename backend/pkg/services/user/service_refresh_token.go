package user

import (
	"github.com/CloudDetail/apo/backend/pkg/middleware"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) RefreshToken(token string) (response.RefreshTokenResponse, error) {
	accessToken, err := middleware.RefreshToken(token)
	var resp response.RefreshTokenResponse
	if err != nil {
		return resp, err
	}
	resp.AccessToken = accessToken
	return resp, nil
}
