package custom

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) DeleteCustomQuery(ctx core.Context, req *request.DeleteCustomRequest) error {
	return s.dbRepo.DeleteCustomQuery(ctx, req.CustomName)
}
