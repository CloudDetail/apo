package user

import "github.com/CloudDetail/apo/backend/pkg/model/request"

func (s *service) CreateUser(req *request.CreateUserRequest) error {
	if err := checkPasswordComplexity(req.Password); err != nil {
		return err
	}
	return s.dbRepo.CreateUser(req.Username, req.Password)
}
