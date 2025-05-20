package profile

type User struct {
	UserID      int64  `gorm:"column:user_id;primary_key" json:"userId,omitempty"`
	Username    string `gorm:"column:username;uniqueIdx;type:varchar(20)" json:"username,omitempty"`
	Password    string `gorm:"column:password;type:varchar(200)" json:"-"`
	Phone       string `gorm:"column:phone;type:varchar(20)" json:"phone,omitempty"`
	Email       string `gorm:"column:email;type:varchar(50)" json:"email,omitempty"`
	Corporation string `gorm:"column:corporation;type:varchar(50)" json:"corporation,omitempty"`

	RoleList    []Role    `gorm:"many2many:user_role;joinForeignKey:UserID;joinReferences:RoleID" json:"roleList,omitempty"`
	TeamList    []Team    `gorm:"many2many:user_team;joinForeignKey:UserID;joinReferences:TeamID" json:"teamList,omitempty"`
	FeatureList []Feature `gorm:"-" json:"featureList,omitempty"`
}

func (t *User) TableName() string {
	return "user"
}
