package user

import "github.com/CloudDetail/apo/backend/pkg/model/request"

func (s *service) CreateUser(req *request.CreateUserRequest) error {
	return s.dbRepo.CreateUser(req.Username, req.Password)
}
