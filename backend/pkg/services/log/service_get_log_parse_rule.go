package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetLogParseRule(req *request.QueryLogParseRequest) (*response.LogParseResponse, error) {
	model := &database.LogTableInfo{
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	err := s.dbRepo.OperateLogTableInfo(model, database.QUERY)
	if err != nil {
		return &response.LogParseResponse{
			ParseName: defaultParseName,
			ParseRule: defaultParseRule,
			RouteRule: defaultRouteRule,
		}, err
	}
	return &response.LogParseResponse{
		ParseName: model.ParseName,
		ParseRule: model.ParseRule,
		RouteRule: model.RouteRule,
	}, nil
}
