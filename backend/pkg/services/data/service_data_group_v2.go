// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) ListDataGroupV2(ctx core.Context) (*datagroup.DataGroupTreeNode, error) {
	userID := ctx.UserID()
	permGroupIDs, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	userGroups := common.DataGroupStorage.CloneWithPermission(permGroupIDs)
	return userGroups, nil
}

type DataGroupWithScopes struct {
	datagroup.DataGroup

	Scopes         []datagroup.DataScope `json:"datasources"`
	PermissionType string                `json:"permissionType"`
}

type SubGroupDetailResponse struct {
	Datasources []datagroup.DataScope `json:"datasources"`
	SubGroups   []DataGroupWithScopes `json:"subGroups"`
}

func (s *service) GetGroupDetailWithSubGroup(ctx core.Context, groupID int64) (*SubGroupDetailResponse, error) {
	userID := ctx.UserID()
	permGroupIDs, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	group := common.DataGroupStorage.CloneGroupNodeWithSubGroup(groupID, permGroupIDs)
	if group == nil {
		return nil, fmt.Errorf("group %d not found", groupID)
	}

	var subGroups []DataGroupWithScopes = make([]DataGroupWithScopes, 0)
	for _, subGroup := range group.SubGroups {
		scopes, err := s.dbRepo.GetScopesByGroupIDAndCat(ctx, subGroup.GroupID, "")
		if err != nil {
			return nil, err
		}
		subGroups = append(subGroups, DataGroupWithScopes{
			DataGroup:      subGroup.DataGroup,
			PermissionType: subGroup.PermissionType,
			Scopes:         scopes,
		})
	}

	scopes, err := s.dbRepo.GetScopesByGroupIDAndCat(ctx, groupID, "")
	if err != nil {
		return nil, err
	}
	return &SubGroupDetailResponse{
		Datasources: scopes,
		SubGroups:   subGroups,
	}, nil
}

func (s *service) CreateDataGroupV2(ctx core.Context, req *request.CreateDataGroupRequest) error {
	parentGroup := common.DataGroupStorage.GetGroupNodeRef(req.ParentGId)
	if parentGroup == nil {
		return fmt.Errorf("parent group %d not found", req.ParentGId)
	}

	// Check Scope exist
	selected, err := s.dbRepo.GetScopeIDsSelectedByGroupID(ctx, req.ParentGId)
	if err != nil {
		return err
	}

	fullPermissionScope := common.DataGroupStorage.GetFullPermissionScopeList(selected)
	for _, id := range req.DataScopeIDs {
		if !containsInStr(fullPermissionScope, id) {
			scope := common.DataGroupStorage.GetScopeRef(id)
			if scope == nil {
				return fmt.Errorf("scope %s not found", id)
			}

			var msg string
			switch ctx.LANG() {
			case code.LANG_EN:
				msg = fmt.Sprintf("permission denied: please add [%s:%s] into parent group[%s] first", "", "", parentGroup.GroupName)
			case code.LANG_ZH:
				msg = fmt.Sprintf("权限不足:请先在上级数据组[%s]中添加[%s:%s]", parentGroup.GroupName, scope.Type, scope.Name)
			}
			return core.Error(code.CreateDataGroupError, msg)
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

	err = s.dbRepo.Transaction(ctx, createGroupFunc, createG2SFunc)
	if err != nil {
		return err
	}
	newGroupTree, err := s.dbRepo.LoadDataGroupTree(ctx)
	if err != nil {
		return err
	}

	common.DataGroupStorage.DataGroupTreeNode = newGroupTree
	return nil
}

func (s *service) UpdateDataGroupV2(ctx core.Context, req *request.UpdateDataGroupRequest) error {
	// Check ParentGroup Option
	options, err := s.dbRepo.GetScopeIDsOptionByGroupID(ctx, req.GroupID)
	fullParentOptions := common.DataGroupStorage.GetFullPermissionScopeList(options)
	if err != nil {
		return err
	}
	for _, id := range req.DataScopeIDs {
		if !containsInStr(fullParentOptions, id) {
			return core.Error(code.UpdateDataGroupError, fmt.Sprintf("unauthorized datasource: %s", id))
		}
	}

	// remove ChildUsed Scope First
	oldSelected, err := s.dbRepo.GetScopeIDsSelectedByGroupID(ctx, req.GroupID)
	if err != nil {
		return err
	}
	oldPermScopeIDs := common.DataGroupStorage.GetFullPermissionScopeList(oldSelected)
	newPermScopeIDs := common.DataGroupStorage.GetFullPermissionScopeList(req.DataScopeIDs)

	removedScopeIDs := []string{}
	for _, id := range oldPermScopeIDs {
		if !containsInStr(newPermScopeIDs, id) {
			removedScopeIDs = append(removedScopeIDs, id)
		}
	}

	subGroupIDs := common.DataGroupStorage.GetFullSubGroupIDs(req.GroupID)

	var updateG2SFunc = func(ctx core.Context) error {
		if len(removedScopeIDs) > 0 {
			err := s.dbRepo.DeleteG2SByGroupsIDsAndScopeIDs(ctx, subGroupIDs, removedScopeIDs)
			if err != nil {
				return err
			}
		}

		return s.dbRepo.UpdateGroup2Scope(ctx, req.GroupID, req.DataScopeIDs)
	}

	var updateNameFunc = func(ctx core.Context) error {
		return s.dbRepo.UpdateDataGroup(ctx, req.GroupID, req.GroupName, req.Description)
	}

	err = s.dbRepo.Transaction(ctx, updateNameFunc, updateG2SFunc)
	if err != nil {
		return err
	}

	newGroupTree, err := s.dbRepo.LoadDataGroupTree(ctx)
	if err != nil {
		return err
	}

	common.DataGroupStorage.DataGroupTreeNode = newGroupTree
	return nil
}

func (s *service) DeleteDataGroupV2(ctx core.Context, req *request.DeleteDataGroupRequest) error {
	filter := model.DataGroupFilter{
		ID: req.GroupID,
	}
	exists, err := s.dbRepo.DataGroupExist(ctx, filter)
	if err != nil {
		return err
	}

	groupNode := common.DataGroupStorage.GetGroupNodeRef(req.GroupID)
	if groupNode != nil && len(groupNode.SubGroups) > 0 {
		return core.Error(code.DeleteDataGroupError, "data group has sub groups")
	}

	if !exists {
		return core.Error(code.DataGroupNotExistError, "data group does not exist")
	}

	var deleteGroupFunc = func(ctx core.Context) error {
		return s.dbRepo.DeleteDataGroup(ctx, req.GroupID)
	}

	var deleteGroup2ScopeFunc = func(ctx core.Context) error {
		return s.dbRepo.DeleteGroup2Scope(ctx, req.GroupID)
	}

	err = s.dbRepo.Transaction(ctx, deleteGroup2ScopeFunc, deleteGroupFunc)
	if err != nil {
		return err
	}

	newGroupTree, err := s.dbRepo.LoadDataGroupTree(ctx)
	if err != nil {
		return err
	}

	common.DataGroupStorage.DataGroupTreeNode = newGroupTree
	return nil
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
