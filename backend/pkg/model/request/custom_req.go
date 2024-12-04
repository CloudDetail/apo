package request

type CreateCustomRequest struct {
	UserName   string   `json:"userName"`
	GroupName  string   `json:"groupName"`
	CustomName string   `json:"customName"`
	Type       []string `json:"type"`
}

type GetCustomRequest struct {
	UserName  string `form:"userName"`
	GroupName string `form:"groupName"`
	//CustomName string `form:"customName"`
}

type UpdateCustomRequest struct {
	UserName   string   `json:"userName"`
	GroupName  string   `json:"groupName"`
	CustomName string   `json:"customName"`
	Type       []string `json:"type"`
}

type DeleteCustomRequest struct {
	UserName   string `form:"userName"`
	GroupName  string `form:"groupName"`
	CustomName string `form:"customName"`
}
