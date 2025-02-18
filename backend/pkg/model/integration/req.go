package integration

type GetCInstallRequest struct {
	ClusterID string `json:"clusterId" form:"clusterId"`
}

type TriggerAdapterUpdateRequest struct {
	LastUpdateTS int64 `form:"lastUpdateTS" json:"lastUpdateTS"`
}
