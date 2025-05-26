// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package profile

// Role is a collection of feature permission.
type Role struct {
	RoleID      int    `gorm:"column:role_id;primary_key;auto_increment" json:"roleId"`
	RoleName    string `gorm:"column:role_name;type:varchar(20);uniqueIndex" json:"roleName"`
	Description string `gorm:"column:description;type:varchar(50)" json:"description"`
}

func (t *Role) TableName() string {
	return "role"
}

type UserRole struct {
	UserID int64 `gorm:"column:user_id;primary_key"`
	RoleID int   `gorm:"column:role_id;primary_key"`
}

func (t *UserRole) TableName() string {
	return "user_role"
}
