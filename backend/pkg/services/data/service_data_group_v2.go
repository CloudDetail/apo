package data

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) ListDataGroupV2(ctx core.Context) (*datagroup.DataGroupTreeNode, error) {
	userID := ctx.UserID()
	permGroupIDs, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	userGroups := s.DataGroupStore.CloneWithPermission(permGroupIDs)
	return userGroups, nil
}

type DataGroupWithScopes struct {
	datagroup.DataGroup

	Scopes         []datagroup.DataScope `json:"datasources"`
	PermissionType string                `json:"permissionType"`
}

type SubGroupDetailResponse struct {
	SubGroups []DataGroupWithScopes `json:"subGroups"`
}

func (s *service) GetGroupDetailWithSubGroup(ctx core.Context, groupID int64) (*SubGroupDetailResponse, error) {
	userID := ctx.UserID()
	permGroupIDs, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	group := s.DataGroupStore.CloneGroupNodeWithSubGroup(groupID, permGroupIDs)
	if group == nil {
		return nil, fmt.Errorf("group %d not found", groupID)
	}

	var subGroups []DataGroupWithScopes
	for _, subGroup := range group.SubGroups {
		scopes, err := s.dbRepo.GetScopesByGroupID(ctx, subGroup.GroupID, "")
		if err != nil {
			return nil, err
		}
		subGroups = append(subGroups, DataGroupWithScopes{
			DataGroup:      subGroup.DataGroup,
			PermissionType: subGroup.PermissionType,
			Scopes:         scopes,
		})
	}
	return &SubGroupDetailResponse{
		SubGroups: subGroups,
	}, nil
}

func (s *service) CreateDataGroupV2(ctx core.Context, req *request.CreateDataGroupRequest) error {
	// TODO Check Group With Same name?

	// Check Scope exist
	selected, err := s.dbRepo.GetScopesSelectedByGroupID(ctx, req.ParentGId)
	if err != nil {
		return err
	}

	fullPermissionScope := s.DataGroupStore.GetFullPermissionScopeList(selected)
	for _, id := range req.DataScopeIDs {
		if !containsInStr(fullPermissionScope, id) {
			return fmt.Errorf("scope %s not in group %d", id, req.ParentGId)
		}
	}

	group := &datagroup.DataGroup{
		GroupID:       util.Generator.GenerateID(),
		GroupName:     req.GroupName,
		Description:   req.Description,
		ParentGroupID: req.ParentGId,
	}

	var createGroupFunc = func(ctx core.Context) error {
		return s.dbRepo.CreateDataGroup(ctx, group)
	}

	var createG2SFunc = func(ctx core.Context) error {
		return s.dbRepo.UpdateGroup2Scope(ctx, group.GroupID, req.DataScopeIDs)
	}

	return s.dbRepo.Transaction(ctx, createGroupFunc, createG2SFunc)
}

func (s *service) UpdateDataGroupV2(ctx core.Context, req *request.UpdateDataGroupRequest) error {
	// Check Scope exist
	options, err := s.dbRepo.GetScopesOptionByGroupID(ctx, req.GroupID)
	if err != nil {
		return err
	}
	for _, id := range req.DataScopeIDs {
		if !containsInStr(options, id) {
			return fmt.Errorf("scope %s not in group %d", id, req.GroupID)
		}
	}

	// Check childGroup Used
	groupNode := s.DataGroupStore.CloneGroupNodeWithSubGroup(req.GroupID, nil)
	if groupNode == nil {
		return fmt.Errorf("group %d not found", req.GroupID)
	}

	fullOptions := s.DataGroupStore.GetFullPermissionScopeList(req.DataScopeIDs)
	for _, subGroup := range groupNode.SubGroups {
		selected, err := s.dbRepo.GetScopesByGroupID(ctx, subGroup.GroupID, "")
		if err != nil {
			return err
		}

		for _, scope := range selected {
			if !containsInStr(fullOptions, scope.ScopeID) {
				return fmt.Errorf("remove datasource [%s:%s] in group [%s] first", scope.Type, scope.Name, subGroup.GroupName)
			}
		}
	}

	var updateNameFunc = func(ctx core.Context) error {
		return s.dbRepo.UpdateDataGroup(ctx, req.GroupID, req.GroupName, req.Description)
	}

	var updateG2SFunc = func(ctx core.Context) error {
		return s.dbRepo.UpdateGroup2Scope(ctx, req.GroupID, req.DataScopeIDs)
	}
	return s.dbRepo.Transaction(ctx, updateNameFunc, updateG2SFunc)
}

func containsInStr(options []string, input string) bool {
	for _, v := range options {
		if v == input {
			return true
		}
	}
	return false
}

func containsInInt(options []int64, input int64) bool {
	for _, v := range options {
		if v == input {
			return true
		}
	}
	return false
}
