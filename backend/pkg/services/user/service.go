package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

var _ Service = (*service)(nil)

type Service interface {
	Login(req *request.LoginRequest) (response.LoginResponse, error)
	CreateUser(req *request.CreateUserRequest) error
	RefreshToken(token string) (response.RefreshTokenResponse, error)
	UpdateUserInfo(req *request.UpdateUserInfoRequest) error
	UpdateUserPhone(username string, req *request.UpdateUserPhoneRequest) error
	UpdateUserEmail(username string, req *request.UpdateUserEmailRequest) error
	UpdateUserPassword(username string, req *request.UpdateUserPasswordRequest) error
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{dbRepo: dbRepo}
}
