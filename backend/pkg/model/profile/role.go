package profile

// Role is a collection of feature permission.
type Role struct {
	RoleID      int    `gorm:"column:role_id;primary_key;auto_increment" json:"roleId"`
	RoleName    string `gorm:"column:role_name;type:varchar(20);uniqueIndex" json:"roleName"`
	Description string `gorm:"column:description;type:varchar(50)" json:"description"`
}
