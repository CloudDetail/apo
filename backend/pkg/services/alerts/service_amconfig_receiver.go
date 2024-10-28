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
	// TODO 根据filter筛选
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
	resp.AMConfigReceivers = append(resp.AMConfigReceivers, amconfig.Receiver{DingTalkConfigs: dingTalkReceivers})
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
		uuid := uuid2.New()
		escapedUUID := url.PathEscape(uuid.String())
		webhookURL := fmt.Sprintf(`http://apo-backend-svc:8080/inputs/dingtalk/%s`, escapedUUID)
		webhookConfig := amconfig.NewWebhookConfig(webhookURL, true)
		req.AMConfigReceiver.WebhookConfigs = []*amconfig.WebhookConfig{webhookConfig}
		err := s.k8sApi.AddAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiver)
		if err != nil {
			return err
		}
		req.AMConfigReceiver.DingTalkConfigs[i].UUID = escapedUUID
		req.AMConfigReceiver.DingTalkConfigs[i].AlertName = req.AMConfigReceiver.Name
		req.AMConfigReceiver.DingTalkConfigs[i].ConfigFile = req.AMConfigFile
		// TODO consistency
		err = s.dbRepo.CreateDingTalkReceiver(req.AMConfigReceiver.DingTalkConfigs[i])
		if err != nil {
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
		uuid := uuid2.New()
		escapedUUID := url.PathEscape(uuid.String())
		webhookURL := fmt.Sprintf(`http://apo-backend-svc:8080/inputs/dingtalk/%s`, escapedUUID)
		webhookConfig := amconfig.NewWebhookConfig(webhookURL, true)
		req.AMConfigReceiver.WebhookConfigs = []*amconfig.WebhookConfig{webhookConfig}
		err := s.k8sApi.UpdateAMConfigReceiver(req.AMConfigFile, req.AMConfigReceiver, req.OldName)
		if err != nil {
			return err
		}
		req.AMConfigReceiver.DingTalkConfigs[i].UUID = escapedUUID
		req.AMConfigReceiver.DingTalkConfigs[i].AlertName = req.AMConfigReceiver.Name
		req.AMConfigReceiver.DingTalkConfigs[i].ConfigFile = req.AMConfigFile
		// TODO consistency
		err = s.dbRepo.UpdateDingTalkReceiver(req.AMConfigReceiver.DingTalkConfigs[i], req.OldName)
		if err != nil {
			return err
		}
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
