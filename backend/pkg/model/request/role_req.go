// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type RoleOperationRequest struct {
	UserID   int64 `form:"userId" binding:"required"`
	RoleList []int `form:"roleList" binding:"required"`
}

type CreateRoleRequest struct {
	RoleName       string  `form:"roleName" binding:"required"`
	Description    string  `form:"description"`
	PermissionList []int   `form:"permissionList"`
	UserList       []int64 `form:"userList"`
}

type UpdateRoleRequest struct {
	RoleID         int    `form:"roleId" binding:"required"`
	RoleName       string `form:"roleName"`
	Description    string `form:"description"`
	PermissionList []int  `form:"permissionList"`
}

type DeleteRoleRequest struct {
	RoleID int `form:"roleId" binding:"required"`
}
