package mock

import (
	"github.com/CloudDetail/apo/backend/internal/model/request"
	"github.com/CloudDetail/apo/backend/internal/model/response"
)

func (s *service) Detail(req *request.DetailRequest) (info *response.DetailResponse, err error) {
	model, err := s.dbRepo.GetMockById(req.Id)
	if err == nil {
		info = &response.DetailResponse{
			Id:   model.ID,
			Name: model.Name,
		}
	}
	return
}
