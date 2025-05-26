// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (repo *daoRepo) CreateTeam(ctx core.Context, team profile.Team) error {
	var count int64
	err := repo.GetContextDB(ctx).Model(&profile.Team{}).Where("team_name = ?", team.TeamName).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return core.Error(code.TeamAlreadyExistError, "team already exists")
	}
	return repo.GetContextDB(ctx).Create(&team).Error
}

func (repo *daoRepo) TeamExist(ctx core.Context, filter model.TeamFilter) (bool, error) {
	var count int64

	query := repo.GetContextDB(ctx).Model(&profile.Team{})
	if filter.ID != 0 {
		query.Where("team_id = ?", filter.ID)
	} else if len(filter.IDs) > 0 {
		query.Where("team_id IN ?", filter.IDs)
	} else if len(filter.Name) > 0 {
		query.Where("team_name = ?", filter.Name)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *daoRepo) GetTeam(ctx core.Context, teamID int64) (profile.Team, error) {
	var team profile.Team
	err := repo.GetContextDB(ctx).Find(&team, teamID).Error
	return team, err
}

func (repo *daoRepo) UpdateTeam(ctx core.Context, team profile.Team) error {
	return repo.GetContextDB(ctx).Save(&team).Error
}

func (repo *daoRepo) GetTeamList(ctx core.Context, req *request.GetTeamRequest) ([]profile.Team, int64, error) {
	var teams []profile.Team
	var count int64

	query := repo.GetContextDB(ctx).Model(&profile.Team{})

	if len(req.TeamName) > 0 {
		query = query.Where("team_name LIKE ?", "%"+req.TeamName+"%")
	}

	if len(req.FeatureList) > 0 {
		subQuery := repo.GetContextDB(ctx).Model(&AuthPermission{}).
			Select("subject_id").
			Joins("JOIN feature f ON f.feature_id = auth_permission.permission_id").
			Where("f.feature_id IN ? AND auth_permission.subject_type = ?", req.FeatureList, model.PERMISSION_SUB_TYP_TEAM)
		query = query.Where("team_id IN (?)", subQuery)
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	if req.PageParam != nil {
		query = query.Limit(req.PageSize).Offset((req.CurrentPage - 1) * req.PageSize)
	}

	err = query.Find(&teams).Error
	if err != nil {
		return nil, 0, err
	}

	teamIDs := make([]int64, len(teams))
	for i, team := range teams {
		teamIDs[i] = team.TeamID
	}

	type TempFeature struct {
		TeamID      int64  `gorm:"column:team_id"`
		FeatureID   int    `gorm:"column:feature_id"`
		FeatureName string `gorm:"column:feature_name"`
	}

	var tempFeatures []TempFeature
	err = repo.GetContextDB(ctx).Model(&AuthPermission{}).
		Select("auth_permission.subject_id as team_id, f.feature_id, f.feature_name").
		Joins("JOIN feature f ON f.feature_id = auth_permission.permission_id").
		Where("auth_permission.subject_type = ? AND auth_permission.subject_id IN ?", model.PERMISSION_SUB_TYP_TEAM, teamIDs).
		Scan(&tempFeatures).Error

	if err != nil {
		return nil, 0, err
	}

	featureMap := make(map[int64][]profile.Feature)
	for _, tf := range tempFeatures {
		featureMap[tf.TeamID] = append(featureMap[tf.TeamID], profile.Feature{
			FeatureID:   tf.FeatureID,
			FeatureName: tf.FeatureName,
		})
	}

	for i := range teams {
		teams[i].FeatureList = featureMap[teams[i].TeamID]
	}

	return teams, count, nil
}

func (repo *daoRepo) DeleteTeam(ctx core.Context, teamID int64) error {
	if err := repo.GetContextDB(ctx).Model(&profile.Team{}).Where("team_id = ?", teamID).Delete(nil).Error; err != nil {
		return err
	}

	return repo.GetContextDB(ctx).Model(&profile.UserTeam{}).Where("team_id = ?", teamID).Delete(nil).Error
}

func (repo *daoRepo) GetUserTeams(ctx core.Context, userID int64) ([]int64, error) {
	var teamIDs []int64
	err := repo.GetContextDB(ctx).Model(&profile.UserTeam{}).Select("team_id").Where("user_id = ?", userID).Find(&teamIDs).Error
	return teamIDs, err
}

func (repo *daoRepo) GetTeamUsers(ctx core.Context, teamID int64) ([]int64, error) {
	var userIDs []int64
	err := repo.GetContextDB(ctx).Model(&profile.UserTeam{}).Select("user_id").Where("team_id = ?", teamID).Find(&userIDs).Error
	return userIDs, err
}

func (repo *daoRepo) GetAssignedTeam(ctx core.Context, userID int64) ([]profile.Team, error) {
	var teams []profile.Team
	subQuery := repo.GetContextDB(ctx).
		Model(&profile.UserTeam{}).
		Select("team_id").
		Where("user_id = ?", userID)
	err := repo.GetContextDB(ctx).Where("team_id IN (?)", subQuery).Find(&teams).Error
	return teams, err
}

func (repo *daoRepo) AssignUserToTeam(ctx core.Context, userID int64, teamIDs []int64) error {
	if len(teamIDs) == 0 {
		return nil
	}
	userTeams := make([]profile.UserTeam, 0, len(teamIDs))
	for _, teamID := range teamIDs {
		ut := profile.UserTeam{
			UserID: userID,
			TeamID: teamID,
		}
		userTeams = append(userTeams, ut)
	}

	return repo.GetContextDB(ctx).Create(&userTeams).Error
}

func (repo *daoRepo) InviteUserToTeam(ctx core.Context, teamID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	userTeams := make([]profile.UserTeam, 0, len(userIDs))
	for _, userID := range userIDs {
		ut := profile.UserTeam{
			UserID: userID,
			TeamID: teamID,
		}
		userTeams = append(userTeams, ut)
	}

	return repo.GetContextDB(ctx).Create(&userTeams).Error
}

// RemoveFromTeamByUser remove user from some of his teams.
func (repo *daoRepo) RemoveFromTeamByUser(ctx core.Context, userID int64, teamIDs []int64) error {
	return repo.GetContextDB(ctx).Model(&profile.UserTeam{}).Where("user_id = ? AND team_id IN ?", userID, teamIDs).Delete(nil).Error
}

// RemoveFromTeamByTeam remove team's some users.
func (repo *daoRepo) RemoveFromTeamByTeam(ctx core.Context, teamID int64, userIDs []int64) error {
	return repo.GetContextDB(ctx).Model(&profile.UserTeam{}).Where("team_id = ? AND user_id IN ?", teamID, userIDs).Delete(nil).Error
}

// DeleteAllUserTeam Used for user or team was deleted.
// If delete the user-related records, by is "user" otherwise by is "team"
func (repo *daoRepo) DeleteAllUserTeam(ctx core.Context, id int64, by string) error {
	query := repo.GetContextDB(ctx).Model(&profile.UserTeam{})
	if by == "user" {
		query.Where("user_id = ?", id)
	} else if by == "team" {
		query.Where("team_id = ?", id)
	}

	return query.Delete(nil).Error
}

func (repo *daoRepo) GetTeamUserList(ctx core.Context, teamID int64) ([]profile.User, error) {
	var users []profile.User
	subQuery := repo.GetContextDB(ctx).Model(&profile.UserTeam{}).Select("user_id").Where("team_id = ?", teamID)
	err := repo.GetContextDB(ctx).Where("user_id IN (?)", subQuery).Find(&users).Error
	return users, err
}

type userTeamMap = map[int64][]profile.Team
type userTeam struct {
	UserID int64 `gorm:"column:user_id"`
	profile.Team
}

func (repo *daoRepo) getTeamByUserID(ctx core.Context, userID ...int64) (userTeamMap, error) {
	var userTeams []userTeam
	err := repo.GetContextDB(ctx).Model(&profile.UserTeam{}).
		Select("user_id", "team.team_id", "team_name", "description").
		Where("user_id IN ?", userID).
		Joins("LEFT JOIN team ON user_team.team_id = team.team_id").
		Order("user_id").
		Scan(&userTeams).Error

	if err != nil {
		return nil, err
	}

	userTeamMap := make(userTeamMap)
	for _, userTeam := range userTeams {
		userTeamMap[userTeam.UserID] = append(userTeamMap[userTeam.UserID], userTeam.Team)
	}
	return userTeamMap, nil
}
