package response

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`  // accessToken用于调用接口获取资源
	RefreshToken string `json:"refreshToken"` // refreshToken用于刷新accessToken
}

type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"` // accessToken用于调用接口获取资源
}

// TODO clean unused response

type CreateUserResponse struct {
}
type LogoutResponse struct {
}
type UpdateUserInfoResponse struct{}

type UpdateUserPhoneResponse struct {
}

type UpdateUserEmailResponse struct {
}

type UpdateUserPasswordResponse struct {
}
