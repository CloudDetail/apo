package request

type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"` // 用户名
	Password string `json:"password" form:"password" binding:"required"` // 密码
}

type CreateUserRequest struct {
	Username        string `json:"username" form:"username" binding:"required"`               // 用户名
	Password        string `json:"password" form:"password" binding:"required"`               // 密码
	ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" binding:"required"` // 确认密码
}

type LogoutRequest struct {
	AccessToken  string `json:"accessToken" form:"accessToken" binding:"required"`
	RefreshToken string `json:"refreshToken" form:"refreshToken" binding:"required"`
}

type UpdateUserInfoRequest struct {
	Corporation string `json:"corporation,omitempty" form:"corporation"`
}

type UpdateUserPhoneRequest struct {
	Phone string `json:"phone" form:"phone" binding:"required"` // 手机号
	VCode string `json:"vCode" form:"vCode"`                    // 验证码
}

type UpdateUserEmailRequest struct {
	Email string `json:"email" form:"email" binding:"required"` // 邮箱
	VCode string `json:"vCode"`                                 // 验证码
}

type UpdateUserPasswordRequest struct {
	OldPassword string `json:"oldPassword" form:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" form:"newPassword" binding:"required"`
}

type GetUserListRequest struct {
	Username    string `json:"username" form:"username"`
	Role        string `json:"role" form:"role"`
	Corporation string `json:"corporation" form:"corporation"`
}

type RemoveUserRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
}
