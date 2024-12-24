package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetUserInfo(userID int64) (response.GetUserInfoResponse, error) {
	var user database.User
	var err error
	if userID == 0 {
		user, err = s.dbRepo.GetAnonymousUser()
	} else {
		user, err = s.dbRepo.GetUserInfo(userID)
	}

	resp := response.GetUserInfoResponse{}
	if err != nil {
		return resp, err
	}

	userRoles, err := s.dbRepo.GetUserRole(user.UserID)
	if err != nil {
		return resp, err
	}

	roleIDs := make([]int, len(userRoles))
	for i := range userRoles {
		roleIDs[i] = userRoles[i].RoleID
	}
	filter := model.RoleFilter{
		IDs: roleIDs,
	}
	roles, err := s.dbRepo.GetRoles(filter)

	permission, err := s.dbRepo.GetSubjectPermission(userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return resp, err
	}
	feature, err := s.dbRepo.GetFeature(permission)
	if err != nil {
		return resp, err
	}
	user.RoleList = roles
	user.FeatureList = feature
	resp.User = user
	return resp, nil
}

func (s *service) GetUserList(req *request.GetUserListRequest) (response.GetUserListResponse, error) {
	users, count, err := s.dbRepo.GetUserList(req)
	resp := response.GetUserListResponse{}
	if err != nil {
		return resp, err
	}
	resp.Users = users
	resp.PageSize = req.PageSize
	resp.CurrentPage = req.CurrentPage
	resp.Total = count
	return resp, nil
}
