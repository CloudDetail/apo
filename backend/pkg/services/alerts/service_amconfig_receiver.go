package alerts

import (
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	uuid2 "github.com/google/uuid"
	"net/url"
)

func (s *service) GetAMConfigReceivers(req *request.GetAlertManagerConfigReceverRequest) response.GetAlertManagerConfigReceiverResponse {
	if req.PageParam == nil {
		req.PageParam = &request.PageParam{
			CurrentPage: 1,
			PageSize:    999,
		}
	}
	// 从内存中获取am的配置
	receivers, totalCount := s.k8sApi.GetAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiverFilter, req.PageParam, req.RefreshCache)
	resp := response.GetAlertManagerConfigReceiverResponse{
		AMConfigReceivers: receivers,
		Pagination: &model.Pagination{
			Total:       int64(totalCount),
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
		},
	}

	if req.AMConfigReceiverFilter != nil && req.AMConfigReceiverFilter.RType != "" && req.AMConfigReceiverFilter.RType != "dingtalk" {
		return resp
	}
	// 计算db的分页参数
	page := req.PageParam.CurrentPage - totalCount/req.PageParam.PageSize
	pageSize := req.PageParam.PageSize - len(receivers)
	name := ""
	if req.AMConfigReceiverFilter != nil {
		name = req.AMConfigReceiverFilter.Name
	}
	dingTalkReceivers, dingTalkCount, err := s.dbRepo.GetDingTalkReceiverByAlertName(req.AMConfigFile, name, page, pageSize)
	if err != nil {
		return resp
	}

	for i := range dingTalkReceivers {
		receiver := amconfig.Receiver{
			Name: dingTalkReceivers[i].AlertName,
			DingTalkConfigs: []*amconfig.DingTalkConfig{
				dingTalkReceivers[i]}}
		resp.AMConfigReceivers = append(resp.AMConfigReceivers, receiver)
	}
	resp.Pagination.Total += dingTalkCount
	return resp
}

func (s *service) AddAMConfigReceiver(req *request.AddAlertManagerConfigReceiver) error {
	if req.Type != "dingtalk" {
		return s.k8sApi.AddAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiver)
	}

	if req.AMConfigReceiver.DingTalkConfigs == nil || len(req.AMConfigReceiver.DingTalkConfigs) == 0 {
		return model.NewErrWithMessage(fmt.Errorf("receiver is empty"), code.AlertManagerEmptyReceiver)
	}

	for i := range req.AMConfigReceiver.DingTalkConfigs {
		uuid, err := s.addDingTalkWebhook(req.AMConfigReceiver.Name)
		if err != nil {
			return err
		}
		req.AMConfigReceiver.DingTalkConfigs[i].UUID = uuid
		req.AMConfigReceiver.DingTalkConfigs[i].AlertName = req.AMConfigReceiver.Name
		req.AMConfigReceiver.DingTalkConfigs[i].ConfigFile = req.AMConfigFile
		err = s.dbRepo.CreateDingTalkReceiver(req.AMConfigReceiver.DingTalkConfigs[i])
		if err != nil {
			s.k8sApi.DeleteAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiver.Name)
			return err
		}
	}
	return nil
}

func (s *service) UpdateAMConfigReceiver(req *request.UpdateAlertManagerConfigReceiver) error {
	if req.Type != "dingtalk" {
		return s.k8sApi.UpdateAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiver, req.OldName)
	}

	if req.AMConfigReceiver.DingTalkConfigs == nil || len(req.AMConfigReceiver.DingTalkConfigs) == 0 {
		return model.NewErrWithMessage(fmt.Errorf("receiver is empty"), code.AlertManagerEmptyReceiver)
	}

	for i := range req.AMConfigReceiver.DingTalkConfigs {
		// regard as a create option
		uuid, err := s.addDingTalkWebhook(req.AMConfigReceiver.Name)
		req.AMConfigReceiver.DingTalkConfigs[i].UUID = uuid
		req.AMConfigReceiver.DingTalkConfigs[i].AlertName = req.AMConfigReceiver.Name
		req.AMConfigReceiver.DingTalkConfigs[i].ConfigFile = req.AMConfigFile

		err = s.dbRepo.UpdateDingTalkReceiver(req.AMConfigReceiver.DingTalkConfigs[i], req.OldName)
		if err != nil {
			// redo
			s.k8sApi.DeleteAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiver.Name)
			return err
		}
		// remove old config
		s.k8sApi.DeleteAMConfigReceiver(req.AMConfigFile, req.OldName)
	}
	return nil
}

func (s *service) DeleteAMConfigReceiver(req *request.DeleteAlertManagerConfigReceiverRequest) error {
	err := s.k8sApi.DeleteAMConfigReceiver(req.AMConfigFile, req.Name)
	if err != nil {
		return err
	}
	if req.Type != "dingtalk" {
		return nil
	}
	return s.dbRepo.DeleteDingTalkReceiver(req.AMConfigFile, req.Name)
}

func (s *service) addDingTalkWebhook(name string) (string, error) {
	uuid := uuid2.New()
	escapedUUID := url.PathEscape(uuid.String())
	webhookURL := fmt.Sprintf(`http://apo-backend-svc:8080/api/alerts/outputs/dingtalk/%s`, escapedUUID)
	webhookConfig := amconfig.NewWebhookConfig(webhookURL)
	req := request.AddAlertManagerConfigReceiver{}
	req.AMConfigReceiver.WebhookConfigs = []*amconfig.WebhookConfig{webhookConfig}
	req.AMConfigReceiver.Name = name
	err := s.k8sApi.AddAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiver)
	if err != nil {
		return "", err
	}
	return escapedUUID, nil
}
