package request

type RemoveAlertSlienceConfigRequest struct {
	AlertID string `json:"alertId"`
}

type GetAlertSlienceConfigRequest struct {
	AlertID string `json:"alertId"`
}

type SetAlertSlienceConfigRequest struct {
	AlertID     string `json:"alertId"`
	ForDuration string `json:"forDuration"`
}
