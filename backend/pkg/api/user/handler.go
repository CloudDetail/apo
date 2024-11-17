package user

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/cache"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/services/user"
	"go.uber.org/zap"
)

type Handler interface {
	// Login 登录
	// @Tags API.user
	// @Router /api/user/login [post]
	Login() core.HandlerFunc
	// Logout 退出登录
	// @Tags API.user
	// @Router /api/user/logout [post]
	Logout() core.HandlerFunc
	// CreateUser 创建用户
	// @Tags API.user
	// @Router /api/user/create [post]
	CreateUser() core.HandlerFunc
	// RefreshToken 刷新accessToken
	// @Tags API.user
	// @Router /api/user/refresh
	RefreshToken() core.HandlerFunc
	// UpdateUserInfo 更新个人信息
	// @Tags API.user
	// @Router /api/user/update/info [post]
	UpdateUserInfo() core.HandlerFunc
	// UpdateUserPassword 更新密码
	// @Tags API.user
	// @Router /api/user/update/password [post]
	UpdateUserPassword() core.HandlerFunc
	// UpdateUserPhone 更新/绑定手机号
	// @Tags API.user
	// @Router /api/user/update/phone [post]
	UpdateUserPhone() core.HandlerFunc
	// UpdateUserEmail 更新/绑定邮箱
	// @Tags API.user
	// @Router /api/user/update/email [post]
	UpdateUserEmail() core.HandlerFunc
	// GetUserInfo 获取个人信息
	// @Tags API.user
	// @Router /api/user/info [get]
	GetUserInfo() core.HandlerFunc

	// GetUserList 获取用户列表
	// @Tags API.user
	// @Router /api/user/list [get]
	GetUserList() core.HandlerFunc

	// RemoveUser 移除用户
	// @Tags API.user
	// @Router /api/user/remove [post]
	RemoveUser() core.HandlerFunc
}

type handler struct {
	logger      *zap.Logger
	userService user.Service
}

func New(logger *zap.Logger, dbRepo database.Repo, cacheRepo cache.Repo) Handler {
	return &handler{
		logger:      logger,
		userService: user.New(dbRepo, cacheRepo),
	}
}
