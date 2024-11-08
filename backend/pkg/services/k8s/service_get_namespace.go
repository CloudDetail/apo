package k8s

import (
	"encoding/json"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s service) GetNamespaceList() (*response.GetNamespaceListResponse, error) {
	list, err := s.k8sRepo.GetNamespaceList()
	if err != nil {
		return nil, err
	}
	listJosn, _ := json.Marshal(list)
	return &response.GetNamespaceListResponse{
		NamespaceList: string(listJosn),
	}, nil
}

func (s service) GetNamespaceInfo(req *request.GetNamespaceInfoRequest) (*response.GetNamespaceInfoResponse, error) {
	info, err := s.k8sRepo.GetNamespaceInfo(req.Namespace)
	if err != nil {
		return nil, err
	}
	infoJson, _ := json.Marshal(info)
	return &response.GetNamespaceInfoResponse{
		NamespaceInfo: string(infoJson),
	}, nil
}
