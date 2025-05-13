// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"gorm.io/gorm"
)

type Team struct {
	TeamID		int64	`gorm:"column:team_id;primary_key" json:"teamId"`
	TeamName	string	`gorm:"column:team_name;type:varchar(20)" json:"teamName"`
	Description	string	`gorm:"column:description;type:varchar(50)" json:"description"`

	UserList	[]User		`gorm:"many2many:user_team;foreignKey:TeamID;joinForeignKey:TeamID;References:UserID;joinReferences:UserID" json:"userList,omitempty"`
	FeatureList	[]Feature	`gorm:"-" json:"featureList,omitempty"`
}

type UserTeam struct {
	UserID	int64	`gorm:"column:user_id;primary_key"`
	TeamID	int64	`gorm:"column:team_id;primary_key"`
}

func (UserTeam) TableName() string {
	return "user_team"
}

func (Team) TableName() string {
	return "team"
}

func (repo *daoRepo) CreateTeam(ctx_core core.Context, ctx context.Context, team Team) error {
	var count int64
	err := repo.GetContextDB(ctx).Model(&Team{}).Where("team_name = ?", team.TeamName).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return core.Error(code.TeamAlreadyExistError, "team already exists")
	}
	return repo.GetContextDB(ctx).Create(&team).Error
}

func (repo *daoRepo) TeamExist(ctx_core core.Context, filter model.TeamFilter) (bool, error) {
	var count int64

	query := repo.db.Model(&Team{})
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

func (repo *daoRepo) GetTeam(ctx_core core.Context, teamID int64) (Team, error) {
	var team Team
	err := repo.db.Find(&team, teamID).Error
	return team, err
}

func (repo *daoRepo) UpdateTeam(ctx_core core.Context, ctx context.Context, team Team) error {
	return repo.GetContextDB(ctx).Save(&team).Error
}

func (repo *daoRepo) GetTeamList(ctx_core core.Context, req *request.GetTeamRequest) ([]Team, int64, error) {
	var teams []Team
	var count int64

	query := repo.db.Model(&Team{}).Preload("UserList", func(db *gorm.DB) *gorm.DB {
		return db.Select("user_id, username")
	})

	if len(req.TeamName) > 0 {
		query = query.Where("team_name LIKE ?", "%"+req.TeamName+"%")
	}

	if len(req.FeatureList) > 0 {
		subQuery := repo.db.Model(&AuthPermission{}).
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
		TeamID		int64	`gorm:"column:team_id"`
		FeatureID	int	`gorm:"column:feature_id"`
		FeatureName	string	`gorm:"column:feature_name"`
	}

	var tempFeatures []TempFeature
	err = repo.db.Model(&AuthPermission{}).
		Select("auth_permission.subject_id as team_id, f.feature_id, f.feature_name").
		Joins("JOIN feature f ON f.feature_id = auth_permission.permission_id").
		Where("auth_permission.subject_type = ? AND auth_permission.subject_id IN ?", model.PERMISSION_SUB_TYP_TEAM, teamIDs).
		Scan(&tempFeatures).Error

	if err != nil {
		return nil, 0, err
	}

	featureMap := make(map[int64][]Feature)
	for _, tf := range tempFeatures {
		featureMap[tf.TeamID] = append(featureMap[tf.TeamID], Feature{
			FeatureID:	tf.FeatureID,
			FeatureName:	tf.FeatureName,
		})
	}

	for i := range teams {
		teams[i].FeatureList = featureMap[teams[i].TeamID]
	}

	return teams, count, nil
}

func (repo *daoRepo) DeleteTeam(ctx_core core.Context, ctx context.Context, teamID int64) error {
	if err := repo.GetContextDB(ctx).Model(&Team{}).Where("team_id = ?", teamID).Delete(nil).Error; err != nil {
		return err
	}

	return repo.GetContextDB(ctx).Model(&UserTeam{}).Where("team_id = ?", teamID).Delete(nil).Error
}

func (repo *daoRepo) GetUserTeams(ctx_core core.Context, userID int64) ([]int64, error) {
	var teamIDs []int64
	err := repo.db.Model(&UserTeam{}).Select("team_id").Where("user_id = ?", userID).Find(&teamIDs).Error
	return teamIDs, err
}

func (repo *daoRepo) GetTeamUsers(ctx_core core.Context, teamID int64) ([]int64, error) {
	var userIDs []int64
	err := repo.db.Model(&UserTeam{}).Select("user_id").Where("team_id = ?", teamID).Find(&userIDs).Error
	return userIDs, err
}

func (repo *daoRepo) GetAssignedTeam(ctx_core core.Context, userID int64) ([]Team, error) {
	var teams []Team
	subQuery := repo.db.
		Model(&UserTeam{}).
		Select("team_id").
		Where("user_id = ?", userID)
	err := repo.db.Where("team_id IN (?)", subQuery).Find(&teams).Error
	return teams, err
}

func (repo *daoRepo) AssignUserToTeam(ctx_core core.Context, ctx context.Context, userID int64, teamIDs []int64) error {
	if len(teamIDs) == 0 {
		return nil
	}
	userTeams := make([]UserTeam, 0, len(teamIDs))
	for _, teamID := range teamIDs {
		ut := UserTeam{
			UserID:	userID,
			TeamID:	teamID,
		}
		userTeams = append(userTeams, ut)
	}

	return repo.GetContextDB(ctx).Create(&userTeams).Error
}

func (repo *daoRepo) InviteUserToTeam(ctx_core core.Context, ctx context.Context, teamID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	userTeams := make([]UserTeam, 0, len(userIDs))
	for _, userID := range userIDs {
		ut := UserTeam{
			UserID:	userID,
			TeamID:	teamID,
		}
		userTeams = append(userTeams, ut)
	}

	return repo.GetContextDB(ctx).Create(&userTeams).Error
}

// RemoveFromTeamByUser remove user from some of his teams.
func (repo *daoRepo) RemoveFromTeamByUser(ctx_core core.Context, ctx context.Context, userID int64, teamIDs []int64) error {
	return repo.GetContextDB(ctx).Model(&UserTeam{}).Where("user_id = ? AND team_id IN ?", userID, teamIDs).Delete(nil).Error
}

// RemoveFromTeamByTeam remove team's some users.
func (repo *daoRepo) RemoveFromTeamByTeam(ctx_core core.Context, ctx context.Context, teamID int64, userIDs []int64) error {
	return repo.GetContextDB(ctx).Model(&UserTeam{}).Where("team_id = ? AND user_id IN ?", teamID, userIDs).Delete(nil).Error
}

// DeleteAllUserTeam Used for user or team was deleted.
// If delete the user-related records, by is "user" otherwise by is "team"
func (repo *daoRepo) DeleteAllUserTeam(ctx_core core.Context, ctx context.Context, id int64, by string) error {
	query := repo.GetContextDB(ctx).Model(&UserTeam{})
	if by == "user" {
		query.Where("user_id = ?", id)
	} else if by == "team" {
		query.Where("team_id = ?", id)
	}

	return query.Delete(nil).Error
}

func (repo *daoRepo) GetTeamUserList(ctx_core core.Context, teamID int64) ([]User, error) {
	var users []User
	subQuery := repo.db.Model(&UserTeam{}).Select("user_id").Where("team_id = ?", teamID)
	err := repo.db.Where("user_id IN (?)", subQuery).Find(&users).Error
	return users, err
}
