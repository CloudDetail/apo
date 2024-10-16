package response

type LogParseResponse struct {
	ParseName string            `json:"parseName"`
	RouteRule map[string]string `json:"routeRule"`
	ParseRule string            `json:"parseRule"`
	Err       string            `json:"error"`
}
