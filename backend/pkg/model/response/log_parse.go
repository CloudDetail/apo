package response

type LogParseResponse struct {
	Service   string            `json:"serviceName"`
	ParseName string            `json:"parseName"`
	RouteRule map[string]string `json:"routeRule"`
	ParseRule string            `json:"parseRule"`
}

type GetServiceRouteResponse struct {
	RouteRule map[string]string `json:"routeRule"`
}
