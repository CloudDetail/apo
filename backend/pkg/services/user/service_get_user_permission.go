package user

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetUserFeature(userID int64) ([]database.Feature, error) {
	// 1. Get user's role
	userRoles, err := s.dbRepo.GetUserRole(userID)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]int, len(userRoles))
	for i := range userRoles {
		roleIDs[i] = userRoles[i].RoleID
	}
	// 2. TODO Get user's team
	subIDs := make([]int64, len(userRoles)+1)
	var i int
	for ; i < len(userRoles); i++ {
		subIDs[i] = int64(userRoles[i].RoleID)
	}
	subIDs[i] = userID
	// 3. Get feature permission.
	permissions, err := s.dbRepo.GetSubjectsPermission(subIDs, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}
	featureIDs := make([]int, len(permissions))
	for i := range permissions {
		featureIDs[i] = permissions[i].PermissionID
	}
	// 4. Get features.
	features, err := s.dbRepo.GetFeature(featureIDs)
	if err != nil {
		return nil, err
	}
	return features, nil
}
