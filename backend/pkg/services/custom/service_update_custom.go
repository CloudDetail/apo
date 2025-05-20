package custom

import (
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) UpdateCustomQuery(ctx core.Context, req *request.UpdateCustomRequest) error {
	return s.dbRepo.UpdateCustomQuery(ctx, req.CustomName, strings.Join(req.Type, ","))
}
