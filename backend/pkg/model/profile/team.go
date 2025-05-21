// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package profile

type Team struct {
	TeamID      int64  `gorm:"column:team_id;primary_key" json:"teamId"`
	TeamName    string `gorm:"column:team_name;type:varchar(20)" json:"teamName"`
	Description string `gorm:"column:description;type:varchar(50)" json:"description"`

	UserList    []User    `gorm:"many2many:user_team;foreignKey:TeamID;joinForeignKey:TeamID;References:UserID;joinReferences:UserID" json:"userList,omitempty"`
	FeatureList []Feature `gorm:"-" json:"featureList,omitempty"`
}

type UserTeam struct {
	UserID int64 `gorm:"column:user_id;primary_key"`
	TeamID int64 `gorm:"column:team_id;primary_key"`
}

func (UserTeam) TableName() string {
	return "user_team"
}

func (Team) TableName() string {
	return "team"
}
