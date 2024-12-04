package custom

import (
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetCustomQuery(req *request.GetCustomRequest) (*response.GetCustomResponse, error) {
	serviecType, err := s.dbRepo.GetCustomQuery(req.UserName, req.GroupName)
	if err != nil {
		return nil, err
	}
	var customType []response.CustonType
	for _, v := range serviecType {
		customType = append(customType, response.CustonType{
			Name: v.CustomName,
			Type: strings.Split(v.QueryType, ","),
		})
	}
	return &response.GetCustomResponse{
		Customs: customType,
	}, nil
}
