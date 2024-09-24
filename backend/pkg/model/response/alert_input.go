package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

type GetAlertRuleFileResponse struct {
	AlertRules map[string]string `json:"alertRules"`
}

type GetAlertRulesResponse struct {
	AlertRules []*request.AlertRule `json:"alertRules"`

	Pagination *model.Pagination `json:"pagination"`
}

type GetAlertManagerConfigReceiverResponse struct {
	AMConfigReceivers []amconfig.Receiver `json:"amConfigReceivers"`

	Pagination *model.Pagination `json:"pagination"`
}
