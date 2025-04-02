// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

import "github.com/CloudDetail/apo/backend/pkg/model"

type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"` // username
	Password string `json:"password" form:"password" binding:"required"` // password
}

type CreateUserRequest struct {
	Username        string  `json:"username" form:"username" binding:"required"`               // 用户名
	Password        string  `json:"password" form:"password" binding:"required"`               // 密码
	ConfirmPassword string  `json:"confirmPassword" form:"confirmPassword" binding:"required"` // 确认密码
	Email           string  `json:"email" form:"email,omitempty"`
	Phone           string  `json:"phone" form:"phone,omitempty"`
	Corporation     string  `json:"corporation,omitempty" form:"corporation,omitempty"`
	RoleList        []int   `json:"roleList" form:"roleList"` // Role id list
	TeamList        []int64 `json:"teamList" form:"teamList"`
	//DataGroupPermissions []DataGroupPermission `json:"dataGroupPermission" form:"dataGroupPermission"`
}

type LogoutRequest struct {
	AccessToken  string `json:"accessToken" form:"accessToken" binding:"required"`
	RefreshToken string `json:"refreshToken" form:"refreshToken" binding:"required"`
}

type UpdateUserInfoRequest struct {
	UserID int64 `json:"userId" form:"userId" binding:"required"`
	//RoleList    []int  `json:"roleList" form:"roleList"`
	Corporation string `json:"corporation,omitempty" form:"corporation,omitempty"`
	Phone       string `json:"phone" form:"phone,omitempty"`
	Email       string `json:"email" form:"email,omitempty"`
}

type UpdateSelfInfoRequest struct {
	UserID      int64  `json:"userId" form:"userId" binding:"required"`
	Corporation string `json:"corporation,omitempty" form:"corporation,omitempty"`
	Phone       string `json:"phone" form:"phone,omitempty"`
	Email       string `json:"email" form:"email,omitempty"`
}

type UpdateUserPhoneRequest struct {
	UserID int64  `json:"userId" form:"userId" binding:"required"`
	Phone  string `json:"phone" form:"phone" binding:"required"` // phone number
	VCode  string `json:"vCode" form:"vCode,omitempty"`          // verification code
}

type UpdateUserEmailRequest struct {
	UserID int64  `json:"userId" form:"userId" binding:"required"`
	Email  string `json:"email" form:"email" binding:"required"` // email
	VCode  string `json:"vCode,omitempty"`                       // verification code
}

type UpdateUserPasswordRequest struct {
	UserID          int64  `json:"userId" form:"userId" binding:"required"`
	OldPassword     string `json:"oldPassword" form:"oldPassword" binding:"required"`
	NewPassword     string `json:"newPassword" form:"newPassword" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" binding:"required"`
}

type GetUserListRequest struct {
	Username    string  `json:"username" form:"username"`
	RoleList    []int   `json:"roleList" form:"roleList"`
	TeamList    []int64 `json:"teamList" form:"teamList"`
	Corporation string  `json:"corporation" form:"corporation"`
	*PageParam
}

type RemoveUserRequest struct {
	UserID int64 `json:"userId" form:"userId" binding:"required"`
}

type ResetPasswordRequest struct {
	UserID          int64  `json:"userId" form:"userId" binding:"required"`
	NewPassword     string `json:"newPassword" form:"newPassword" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" binding:"required"`
}

type GetUserConfigRequest struct {
	UserID int64 `form:"userId" binding:"required"`
	model.I18nLanguage
}

type GetSubjectFeatureRequest struct {
	SubjectID   int64  `form:"subjectId" binding:"required"`
	SubjectType string `form:"subjectType" binding:"required"`
	model.I18nLanguage
}

type PermissionOperationRequest struct {
	SubjectID      int64  `form:"subjectId" binding:"required"`
	SubjectType    string `form:"subjectType" binding:"required"` // "user", "role", "team"
	Type           string `form:"type" binding:"required"`        // "feature", "data"
	PermissionList []int  `form:"permissionList"`
}

type GetUserRoleRequest struct {
	UserID int64 `form:"userId"`
}

type ConfigureMenuRequest struct {
	PermissionList []int `form:"permissionList"`
}

type GetFeatureRequest struct {
	model.I18nLanguage
}

type GetUserInfoRequest struct {
	UserID int64 `form:"userId"`
}

type CheckRouterPermissionRequest struct {
	RouterID int `form:"routerId"`
}