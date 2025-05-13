// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GroupSubsOperation(req *request.GroupSubsOperationRequest) error {
	var (
		toDelete []int64
		toAdd    []database.AuthDataGroup
		subMap   = map[int64]database.AuthDataGroup{}
	)

	filter := model.DataGroupFilter{
		ID: req.DataGroupID,
	}
	exists, err := s.dbRepo.DataGroupExist(filter)
	if err != nil {
		return err
	}
	if !exists {
		return core.Error(code.DataGroupNotExistError, "data group not exist")
	}

	getAuthDataGroups := func(subjectType string) error {
		authGroups, err := s.dbRepo.GetGroupAuthDataGroupByGroup(req.DataGroupID, subjectType)
		if err != nil {
			return err
		}
		for _, ag := range authGroups {
			subMap[ag.SubjectID] = ag
		}
		return nil
	}

	if err := getAuthDataGroups(model.DATA_GROUP_SUB_TYP_USER); err != nil {
		return err
	}

	if err := getAuthDataGroups(model.DATA_GROUP_SUB_TYP_TEAM); err != nil {
		return err
	}

	handleSubjects := func(subjects []request.AuthDataGroup, subjectType string) error {
		for _, sub := range subjects {
			switch subjectType {
			case model.DATA_GROUP_SUB_TYP_USER:
				exists, err = s.dbRepo.UserExists(sub.SubjectID)
			case model.DATA_GROUP_SUB_TYP_TEAM:
				filter := model.TeamFilter{
					ID: sub.SubjectID,
				}
				exists, err = s.dbRepo.TeamExist(filter)
			}
			if err != nil {
				return err
			}
			if !exists {
				continue
			}

			ag, ok := subMap[sub.SubjectID]
			if !ok {
				toAdd = append(toAdd, database.AuthDataGroup{
					SubjectID:   sub.SubjectID,
					SubjectType: subjectType,
					GroupID:     req.DataGroupID,
					Type:        sub.Type,
				})
			} else {
				if ag.Type != sub.Type {
					ag.Type = sub.Type
					toAdd = append(toAdd, ag)
				}
				delete(subMap, sub.SubjectID)
			}
		}
		return nil
	}

	if err := handleSubjects(req.UserList, model.DATA_GROUP_SUB_TYP_USER); err != nil {
		return err
	}

	if err := handleSubjects(req.TeamList, model.DATA_GROUP_SUB_TYP_TEAM); err != nil {
		return err
	}

	for subID := range subMap {
		toDelete = append(toDelete, subID)
	}

	assignFunc := func(ctx context.Context) error {
		return s.dbRepo.AssignDataGroup(ctx, toAdd)
	}

	revokeFunc := func(ctx context.Context) error {
		return s.dbRepo.RevokeDataGroupBySub(ctx, toDelete, req.DataGroupID)
	}

	return s.dbRepo.Transaction(context.Background(), assignFunc, revokeFunc)
}
