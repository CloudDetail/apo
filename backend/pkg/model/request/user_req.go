package request

type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"` // 用户名
	Password string `json:"password" form:"password" binding:"required"` // 密码
}

type CreateUserRequest struct {
	Username        string `json:"username" form:"username" binding:"required"`               // 用户名
	Password        string `json:"password" form:"password" binding:"required"`               // 密码
	ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" binding:"required"` // 确认密码
	Email           string `json:"email" form:"email,omitempty"`
	Phone           string `json:"phone" form:"phone,omitempty"`
	Corporation     string `json:"corporation,omitempty" form:"corporation,omitempty"`
}

type LogoutRequest struct {
	AccessToken  string `json:"accessToken" form:"accessToken" binding:"required"`
	RefreshToken string `json:"refreshToken" form:"refreshToken" binding:"required"`
}

type UpdateUserInfoRequest struct {
	UserID      string `json:"userId" form:"userId" binding:"required"`
	Corporation string `json:"corporation,omitempty" form:"corporation,omitempty"`
	Phone       string `json:"phone" form:"phone,omitempty"`
	Email       string `json:"email" form:"email,omitempty"`
}

type UpdateUserPhoneRequest struct {
	UserID int64  `json:"userId" form:"userId" binding:"required"`
	Phone  string `json:"phone" form:"phone" binding:"required"` // 手机号
	VCode  string `json:"vCode" form:"vCode,omitempty"`          // 验证码
}

type UpdateUserEmailRequest struct {
	UserID int64  `json:"userId" form:"userId" binding:"required"`
	Email  string `json:"email" form:"email" binding:"required"` // 邮箱
	VCode  string `json:"vCode,omitempty"`                       // 验证码
}

type UpdateUserPasswordRequest struct {
	UserID          int64  `json:"userId" form:"userId" binding:"required"`
	OldPassword     string `json:"oldPassword" form:"oldPassword" binding:"required"`
	NewPassword     string `json:"newPassword" form:"newPassword" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" binding:"required"`
}

type GetUserListRequest struct {
	Username    string `json:"username" form:"username"`
	Role        string `json:"role" form:"role"`
	Corporation string `json:"corporation" form:"corporation"`
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
