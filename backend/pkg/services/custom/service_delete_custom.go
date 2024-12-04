package custom

import "github.com/CloudDetail/apo/backend/pkg/model/request"

func (s *service) DeleteCustomQuery(req *request.DeleteCustomRequest) error {
	return s.dbRepo.DeleteCustomQuery(req.UserName, req.GroupName, req.CustomName)
}
