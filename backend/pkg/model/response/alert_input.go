package response

type GetAlertRuleResponse struct {
	AlertRules map[string]string `json:"alertRules"`
}
