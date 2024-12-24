package user

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (s *service) RemoveUser(userID int64) error {
	exists, err := s.dbRepo.UserExists(userID)
	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("user does not exist"), code.UserNotExistsError)
	}

	// 1. Get roles.
	roles, err := s.dbRepo.GetUserRole(userID)
	if err != nil {
		return err
	}

	roleIDs := make([]int, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.RoleID)
	}

	// 2. Get feature permissions.
	fPermissionIDs, err := s.dbRepo.GetSubjectPermission(userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return err
	}

	// 3. TODO Get data permission.
	// 4. TODO Get teams.

	var removeUserFunc = func(ctx context.Context) error {
		return s.dbRepo.RemoveUser(ctx, userID)
	}

	var revokeRoleFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokeRole(ctx, userID, roleIDs)
	}

	var revokeFeaturePermFunc = func(ctx context.Context) error {
		return s.dbRepo.RevokePermission(ctx, userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE, fPermissionIDs)
	}

	return s.dbRepo.Transaction(context.Background(), revokeRoleFunc, revokeFeaturePermFunc, removeUserFunc)
}
