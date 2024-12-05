package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) CreateUser(req *request.CreateUserRequest) error {
	if err := checkPasswordComplexity(req.Password); err != nil {
		return err
	}
	user := &database.User{
		UserID:      util.Generator.GenerateID(),
		Username:    req.Username,
		Password:    req.Password,
		Corporation: req.Corporation,
		Email:       req.Email,
		Phone:       req.Phone,
	}
	return s.dbRepo.CreateUser(user)
}
