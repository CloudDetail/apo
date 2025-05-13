// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) CreateDataGroup(ctx_core core.Context, req *request.CreateDataGroupRequest) error {
	filter := model.DataGroupFilter{
		Name: req.GroupName,
	}
	exists, err := s.dbRepo.DataGroupExist(filter)
	if err != nil {
		return err
	}

	if exists {
		return core.Error(code.DataGroupExistError, "data group already exists")
	}
	group := database.DataGroup{
		Description:	req.Description,
		GroupName:	req.GroupName,
		GroupID:	util.Generator.GenerateID(),
	}

	var createGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateDataGroup(ctx, &group)
	}

	var createDSGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateDatasourceGroup(ctx, req.DatasourceList, group.GroupID)
	}

	return s.dbRepo.Transaction(context.Background(), createGroupFunc, createDSGroupFunc)
}

func (s *service) DeleteDataGroup(ctx_core core.Context, req *request.DeleteDataGroupRequest) error {
	filter := model.DataGroupFilter{
		ID: req.GroupID,
	}
	exists, err := s.dbRepo.DataGroupExist(filter)
	if err != nil {
		return err
	}

	if !exists {
		return core.Error(code.DataGroupNotExistError, "data group does not exist")
	}

	var deleteGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteDataGroup(ctx, req.GroupID)
	}

	var deleteDSGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteDSGroup(ctx, req.GroupID)
	}

	return s.dbRepo.Transaction(context.Background(), deleteDSGroupFunc, deleteGroupFunc)
}

func (s *service) UpdateDataGroup(ctx_core core.Context, req *request.UpdateDataGroupRequest) error {
	idFilter := model.DataGroupFilter{
		ID: req.GroupID,
	}
	groups, _, err := s.dbRepo.GetDataGroup(idFilter)
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		return core.Error(code.DataGroupNotExistError, "data group does not exist")
	}

	group := groups[0]
	nameFilter := model.DataGroupFilter{
		Name: req.GroupName,
	}

	if group.GroupName != req.GroupName {
		exists, err := s.dbRepo.DataGroupExist(nameFilter)
		if err != nil {
			return err
		}

		if exists {
			return core.Error(code.DataGroupExistError, "data group already exist")
		}
	}

	// 1. Get data group's datasource
	dsGroups, err := s.dbRepo.GetGroupDatasource(req.GroupID)
	if err != nil {
		return err
	}

	// 2. Get all datasource
	datasource, err := s.GetDataSource()
	if err != nil {
		return err
	}

	dsMap := make(map[string]struct{})
	for _, data := range datasource.NamespaceList {
		dsMap[data.Datasource] = struct{}{}
	}
	for _, data := range datasource.ServiceList {
		dsMap[data.Datasource] = struct{}{}
	}

	groupDsMap := make(map[string]struct{})
	for _, dsg := range dsGroups {
		groupDsMap[dsg.Datasource] = struct{}{}
	}

	// 3. Determine assign and retrieve
	var addData []model.Datasource
	var deleteData []string
	for _, data := range req.DatasourceList {
		if _, exists := dsMap[data.Datasource]; !exists {
			// skip if datasource does not exist
			continue
		}
		if _, hasData := groupDsMap[data.Datasource]; !hasData {
			addData = append(addData, data)
		} else {
			delete(groupDsMap, data.Datasource)
		}
	}

	for ds := range groupDsMap {
		deleteData = append(deleteData, ds)
	}

	var updateNameFunc = func(ctx context.Context) error {
		return s.dbRepo.UpdateDataGroup(ctx, req.GroupID, req.GroupName, req.Description)
	}

	var assignFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateDatasourceGroup(ctx, addData, req.GroupID)
	}

	var retrieveFunc = func(ctx context.Context) error {
		return s.dbRepo.RetrieveDataFromGroup(ctx, req.GroupID, deleteData)
	}

	return s.dbRepo.Transaction(context.Background(), updateNameFunc, assignFunc, retrieveFunc)
}

func (s *service) GetDataGroup(ctx_core core.Context, req *request.GetDataGroupRequest) (resp response.GetDataGroupResponse, err error) {
	filter := model.DataGroupFilter{
		Name:		req.GroupName,
		PageSize:	&req.PageSize,
		CurrentPage:	&req.CurrentPage,
		DatasourceList:	req.DataSourceList,
	}

	dataGroups, count, err := s.dbRepo.GetDataGroup(filter)
	if err != nil {
		return
	}

	resp.DataGroupList = dataGroups
	resp.Total = count
	resp.CurrentPage = req.CurrentPage
	resp.PageSize = req.PageSize
	return
}
