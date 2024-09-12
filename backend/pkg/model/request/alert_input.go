package request

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

type GetAlertRuleRequest struct {
	AlertRuleFile string `form:"alertRuleFile" json:"alertRuleFile"`
}

type UpdateAlertRuleRequest struct {
	AlertRules map[string]string `json:"alertRules"`
}
