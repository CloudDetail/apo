package k8s

import (
	"encoding/json"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s service) GetPodList(req *request.GetPodListRequest) (*response.GetPodListResponse, error) {
	list, err := s.k8sRepo.GetPodList(req.Namespace)
	if err != nil {
		return nil, err
	}
	listJson, _ := json.Marshal(list)
	return &response.GetPodListResponse{
		PodList: string(listJson),
	}, nil
}

func (s service) GetPodInfo(req *request.GetPodInfoRequest) (*response.GetPodInfoResponse, error) {
	info, err := s.k8sRepo.GetPodInfo(req.Namespace, req.Pod)
	if err != nil {
		return nil, err
	}
	infoJson, _ := json.Marshal(info)
	return &response.GetPodInfoResponse{
		PodInfo: string(infoJson),
	}, nil
}
