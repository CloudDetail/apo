package data

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) CreateDataGroup(req *request.CreateDataGroupRequest) error {
	filter := model.DataGroupFilter{
		Name: req.GroupName,
	}
	exists, err := s.dbRepo.DataGroupExist(filter)
	if err != nil {
		return err
	}

	if exists {
		return model.NewErrWithMessage(errors.New("data group already exists"), code.DataGroupExistError)
	}

	datasource, err := s.GetDataSource()
	if err != nil {
		return err
	}

	dsMap := make(map[model.Datasource]struct{})
	for _, data := range datasource {
		dsMap[data] = struct{}{}
	}

	for _, ds := range req.DatasourceList {
		if _, ok := dsMap[ds]; !ok {
			return model.NewErrWithMessage(errors.New("datasource does not exist"), code.DatasourceNotExistError)
		}
	}

	var users, teams []int64
	groupID := util.Generator.GenerateID()
	authDataGroup := make([]database.AuthDataGroup, 0, len(req.AssignedSubjects))

	for _, sub := range req.AssignedSubjects {
		switch sub.SubjectType {
		case model.DATA_GROUP_SUB_TYP_USER:
			users = append(users, sub.SubjectID)
		case model.DATA_GROUP_SUB_TYP_TEAM:
			teams = append(teams, sub.SubjectID)
		default:
			continue
		}

		dg := database.AuthDataGroup{
			SubjectType: sub.SubjectType,
			SubjectID:   sub.SubjectID,
			Type:        sub.Type,
			DataGroupID: groupID,
		}
		authDataGroup = append(authDataGroup, dg)
	}

	exists, err = s.dbRepo.UserExists(users...)
	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("user does not exist"), code.UserNotExistsError)
	}

	exists, err = s.dbRepo.TeamExist(teams...)
	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("team does not exist"), code.TeamNotExistError)
	}

	var assignToTeamFunc = func(ctx context.Context) error {
		return s.dbRepo.AssignDataGroup(ctx, authDataGroup)
	}

	var createGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateDataGroup(ctx, groupID, req.GroupName, req.Description)
	}

	var createDSGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateDatasourceGroup(ctx, req.DatasourceList, groupID)
	}

	return s.dbRepo.Transaction(context.Background(), createGroupFunc, createDSGroupFunc, assignToTeamFunc)
}

func (s *service) DeleteDataGroup(req *request.DeleteDataGroupRequest) error {
	filter := model.DataGroupFilter{
		ID: req.GroupID,
	}
	exists, err := s.dbRepo.DataGroupExist(filter)
	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("data group does not exist"), code.DataGroupNotExistError)
	}

	var deleteGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteDataGroup(ctx, req.GroupID)
	}

	var deleteDSGroupFunc = func(ctx context.Context) error {
		return s.dbRepo.DeleteDSGroup(ctx, req.GroupID)
	}

	return s.dbRepo.Transaction(context.Background(), deleteGroupFunc, deleteDSGroupFunc)
}

func (s *service) UpdateDataGroupName(req *request.UpdateDataGroupNameRequest) error {
	filter := model.DataGroupFilter{
		ID: req.GroupID,
	}
	exists, err := s.dbRepo.DataGroupExist(filter)
	if err != nil {
		return err
	}

	if !exists {
		return model.NewErrWithMessage(errors.New("data group does not exist"), code.DataGroupNotExistError)
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
	for _, data := range datasource {
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
		if _, exists = dsMap[data.Datasource]; !exists {
			return model.NewErrWithMessage(errors.New("datasource does not exist"), code.DataSourceNotExistError)
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
		return s.dbRepo.UpdateDataGroupName(ctx, req.GroupID, req.GroupName, req.Description)
	}

	var assignFunc = func(ctx context.Context) error {
		return s.dbRepo.CreateDatasourceGroup(ctx, addData, req.GroupID)
	}

	var retrieveFunc = func(ctx context.Context) error {
		return s.dbRepo.RetrieveDataFromGroup(ctx, req.GroupID, deleteData)
	}

	return s.dbRepo.Transaction(context.Background(), updateNameFunc, assignFunc, retrieveFunc)
}

func (s *service) GetDataGroup(req *request.GetDataGroupRequest) (resp response.GetDataGroupResponse, err error) {
	filter := model.DataGroupFilter{
		Name:           req.GroupName,
		PageSize:       &req.PageSize,
		CurrentPage:    &req.CurrentPage,
		DatasourceList: req.DataSourceList,
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
