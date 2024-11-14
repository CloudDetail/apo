package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

var _ Service = (*service)(nil)

type Service interface {
	Login(req *request.LoginRequest) (response.LoginResponse, error)
	Logout(req *request.LogoutRequest) error
	CreateUser(req *request.CreateUserRequest) error
	RefreshToken(token string) (response.RefreshTokenResponse, error)
	UpdateUserInfo(username string, req *request.UpdateUserInfoRequest) error
	UpdateUserPhone(username string, req *request.UpdateUserPhoneRequest) error
	UpdateUserEmail(username string, req *request.UpdateUserEmailRequest) error
	UpdateUserPassword(username string, req *request.UpdateUserPasswordRequest) error
	GetUserInfo(username string) (response.GetUserInfoResponse, error)
}

type service struct {
	dbRepo    database.Repo
	cacheRepo cache.Repo
}

func New(dbRepo database.Repo, cacheRepo cache.Repo) Service {
	return &service{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
	}
}
