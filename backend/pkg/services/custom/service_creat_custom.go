package custom

import (
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) CreateCustomQuery(req *request.CreateCustomRequest) error {
	serviceType := &database.QueryServiceType{
		CustomName: req.CustomName,
		GroupName:  req.GroupName,
		QueryType:  strings.Join(req.Type, ","),
		UserName:   req.UserName,
	}
	return s.dbRepo.CreateCustomQuery(serviceType)
}
