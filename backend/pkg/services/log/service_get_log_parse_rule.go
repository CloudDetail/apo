package log

import (
	"errors"
	"gorm.io/gorm"
	"regexp"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

var routeReg = regexp.MustCompile(`\"(.*?)\"`)

func getRouteRuleMap(routeRule string) map[string]string {
	res := make(map[string][]string)
	lines := strings.Split(routeRule, "||")
	for _, line := range lines {
		if line == "" {
			continue
		}
		matches := routeReg.FindAllStringSubmatch(line, -1)
		if len(matches) == 2 {
			key := matches[0][1]
			value := matches[1][1]
			// if key == "k8s.pod.name" {
			// 	continue
			// }
			res[key] = append(res[key], value)
		}
	}
	rc := make(map[string]string)
	for k, v := range res {
		rc[k] = strings.Join(v, ",")
	}
	return rc
}

func (s *service) GetLogParseRule(req *request.QueryLogParseRequest) (*response.LogParseResponse, error) {
	model := &database.LogTableInfo{
		DataBase: req.DataBase,
		Table:    req.TableName,
	}
	err := s.dbRepo.OperateLogTableInfo(model, database.QUERY)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.LogParseResponse{
			ParseName: defaultParseName,
			ParseRule: defaultParseRule,
			RouteRule: defaultRouteRuleMap,
		}, nil
	} else if err != nil {
		return nil, err
	}

	return &response.LogParseResponse{
		Service:   strings.Split(model.Service, ","),
		ParseName: model.ParseName,
		ParseRule: model.ParseRule,
		ParseInfo: model.ParseInfo,
		RouteRule: getRouteRuleMap(model.RouteRule),
	}, nil
}
