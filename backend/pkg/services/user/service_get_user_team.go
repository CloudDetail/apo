package user

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetUserTeam(req *request.GetUserTeamRequest) (response.GetUserTeamResponse, error) {
	exists, err := s.dbRepo.UserExists(req.UserID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, model.NewErrWithMessage(errors.New("user does not exist"), code.UserNotExistsError)
	}

	return s.dbRepo.GetAssignedTeam(req.UserID)
}
