package request

type RemoveAlertSlienceConfigRequest struct {
	AlertID string `json:"alertId" form:"alertId"`
}

type GetAlertSlienceConfigRequest struct {
	AlertID string `json:"alertId" form:"alertId"`
}

type SetAlertSlienceConfigRequest struct {
	AlertID     string `json:"alertId" form:"alertId"`
	ForDuration string `json:"forDuration" form:"forDuration"`
}
