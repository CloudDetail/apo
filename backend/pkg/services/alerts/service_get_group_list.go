package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
)

func (s *service) GetGroupList() response.GetGroupListResponse {
	resp := response.GetGroupListResponse{
		GroupsLabel: kubernetes.GroupsLabel,
	}

	return resp
}
