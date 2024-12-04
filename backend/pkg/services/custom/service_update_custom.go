package custom

import (
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) UpdateCustomQuery(req *request.UpdateCustomRequest) error {
	return s.dbRepo.UpdateCustomQuery(req.UserName, req.GroupName, req.CustomName, strings.Join(req.Type, ","))
}
