package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetAMConfigReceivers(req *request.GetAlertManagerConfigReceverRequest) response.GetAlertManagerConfigReceiverResponse {
	if req.PageParam == nil {
		req.PageParam = &request.PageParam{
			CurrentPage: 1,
			PageSize:    999,
		}
	}
	receivers, totalCount := s.k8sApi.GetAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiverFilter, req.PageParam, req.RefreshCache)
	return response.GetAlertManagerConfigReceiverResponse{
		AMConfigReceivers: receivers,
		Pagination: &model.Pagination{
			Total:       int64(totalCount),
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
		},
	}
}

func (s *service) AddAMConfigReceiver(req *request.AddAlertManagerConfigReceiver) error {
	return s.k8sApi.AddAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiver)
}

func (s *service) UpdateAMConfigReceiver(req *request.UpdateAlertManagerConfigReceiver) error {
	return s.k8sApi.UpdateAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiver, req.OldName)
}

func (s *service) DeleteAMConfigReceiver(req *request.DeleteAlertManagerConfigReceiverRequest) error {
	return s.k8sApi.DeleteAMConfigReceiver(req.AMConfigFile, req.Name)
}
