package response

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`  // accessToken用于调用接口获取资源
	RefreshToken string `json:"refreshToken"` // refreshToken用于刷新accessToken
}

type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"` // accessToken用于调用接口获取资源
}
