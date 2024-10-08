package request

import "github.com/CloudDetail/apo/backend/pkg/model/amconfig"

type InputAlertManagerRequest struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type GetAlertRuleConfigRequest struct {
	AlertRuleFile string `form:"alertRuleFile" json:"alertRuleFile"`
}

type GetAlertRuleRequest struct {
	AlertRuleFile string `form:"alertRuleFile" json:"alertRuleFile"`
	RefreshCache  bool   `form:"refreshCache" json:"refreshCache"`

	*AlertRuleFilter `json:",inline"`
	*PageParam       `json:",inline"`
}

type GetAlertManagerConfigReceverRequest struct {
	AMConfigFile string `form:"amConfigFile" json:"amConfigFile"`
	RefreshCache bool   `form:"refreshCache" json:"refreshCache"`

	*AMConfigReceiverFilter
	*PageParam
}

type AlertRuleFilter struct {
	Group    string   `form:"group" json:"group"`
	Alert    string   `form:"alert" json:"alert"`
	Severity []string `form:"severity" json:"severity"` // 告警级别 info warning ...
	Keyword  string   `form:"keyword" json:"keyword"`
}

type AMConfigReceiverFilter struct {
	Name  string `form:"name" json:"name"`
	RType string `form:"rType" json:"rType"`
}

type UpdateAlertRuleConfigRequest struct {
	AlertRuleFile string `json:"alertRuleFile"`
	Content       string `json:"content"`
}

type UpdateAlertRuleRequest struct {
	AlertRuleFile string `json:"alertRuleFile"`

	OldGroup  string    `json:"oldGroup" binding:"required"`
	OldAlert  string    `json:"oldAlert" binding:"required"`
	AlertRule AlertRule `json:"alertRule"`
}

type AddAlertManagerConfigReceiver UpdateAlertManagerConfigReceiver

type UpdateAlertManagerConfigReceiver struct {
	AMConfigFile string `form:"amConfigFile" json:"amConfigFile"`

	OldName          string            `form:"oldName" json:"oldName"`
	AMConfigReceiver amconfig.Receiver `form:"amConfigReceiver" json:"amConfigReceiver"`
}

type DeleteAlertRuleRequest struct {
	AlertRuleFile string `form:"alertRuleFile" json:"alertRuleFile"`

	Group string `form:"group" json:"group" binding:"required"`
	Alert string `form:"alert" json:"alert" binding:"required"`
}

type DeleteAlertManagerConfigReceiverRequest struct {
	AMConfigFile string `form:"amConfigFile" json:"amConfigFile"`

	Name string `form:"name" json:"name" binding:"required"`
}

type AlertRule struct {
	Group string `json:"group" binding:"required"`

	Record        string            `json:"record"`
	Alert         string            `json:"alert" binding:"required"`
	Expr          string            `json:"expr"`
	For           string            `json:"for,omitempty"`
	KeepFiringFor string            `json:"keepFiringFor,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

type AddAlertRuleRequest struct {
	AlertRuleFile string `json:"alertRuleFile"`

	AlertRule AlertRule `json:"alertRule"`
}

type CheckAlertRuleRequest struct {
	AlertRuleFile string `form:"alertRuleFile,omitempty"`
	Group         string `form:"group" binding:"required"`
	Alert         string `form:"alert" binding:"required"`
}
