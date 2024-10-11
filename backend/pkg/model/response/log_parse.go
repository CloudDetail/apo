package response

type LogParseResponse struct {
	ParseName string `json:"parseName"`
	RouteRule string `json:"routeRule"`
	ParseRule string `json:"parseRule"`
}
