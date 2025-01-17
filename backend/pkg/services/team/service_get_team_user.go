package team

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetTeamUser(req *request.GetTeamUserRequest) (resp response.GetTeamUserResponse, err error) {
	exists, err := s.dbRepo.TeamExist(req.TeamID)
	if err != nil {
		return
	}

	if !exists {
		err = model.NewErrWithMessage(errors.New("team does not exist"), code.TeamNotExistError)
		return
	}

	users, err := s.dbRepo.GetTeamUserList(req.TeamID)
	if err != nil {
		return
	}
	resp = users
	return
}