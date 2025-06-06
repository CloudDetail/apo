package custom

import (
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) CreateCustomQuery(ctx core.Context, req *request.CreateCustomRequest) error {
	serviceType := &database.QueryServiceType{
		CustomName: req.CustomName,
		QueryType:  strings.Join(req.Type, ","),
	}
	return s.dbRepo.CreateCustomQuery(ctx, serviceType)
}
