package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

var (
	// DefaultFields 默认展示字段
	defaultFields = []string{
		"timestamp",
		"content",
		"source",
		"container_id",
		"pid",
		"container_name",
		"host_ip",
		"host_name",
		"k8s_namespace_name",
		"k8s_pod_name",
	}
)

func (s *service) QueryLog(req *request.LogQueryRequest) (*response.LogQueryResponse, error) {
	logs, sql, err := s.chRepo.QueryAllLogs(req)
	if err != nil {
		return nil, err
	}
	hiddenFields := []string{}
	// model := &database.LogTableInfo{
	// 	DataBase: req.DataBase,
	// 	Table:    req.TableName,
	// }
	// err = s.dbRepo.OperateLogTableInfo(model, database.QUERY)
	// if err != nil {
	// 	return nil, err
	// }
	// var fields []request.Field
	// err = json.Unmarshal([]byte(model.Fields), &fields)
	// if err != nil {
	// 	return nil, err
	// }

	// for _, field := range fields {
	// 	hiddenFields = append(hiddenFields, field.Name)
	// }
	res := &response.LogQueryResponse{
		Limited:       req.PageSize,
		HiddenFields:  hiddenFields,
		DefaultFields: defaultFields,
		Logs:          logs,
		Query:         sql,
	}

	return res, nil
}
