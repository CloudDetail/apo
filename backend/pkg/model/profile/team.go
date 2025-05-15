package profile

import "github.com/CloudDetail/apo/backend/pkg/model/permission"

type Team struct {
	TeamID      int64  `gorm:"column:team_id;primary_key" json:"teamId"`
	TeamName    string `gorm:"column:team_name;type:varchar(20)" json:"teamName"`
	Description string `gorm:"column:description;type:varchar(50)" json:"description"`

	UserList    []User               `gorm:"many2many:user_team;foreignKey:TeamID;joinForeignKey:TeamID;References:UserID;joinReferences:UserID" json:"userList,omitempty"`
	FeatureList []permission.Feature `gorm:"-" json:"featureList,omitempty"`
}
