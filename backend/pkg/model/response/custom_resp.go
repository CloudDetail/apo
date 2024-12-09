package response

type GetCustomResponse struct {
	Customs []CustonType `json:"customs"`
}

type CustonType struct {
	Name string   `json:"name"`
	Type []string `json:"type"`
}
