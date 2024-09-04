package model

const (
	DelaySourceAlert    = "delaySource"
	InfrastructureAlert = "infrastructureStatus"
	NetAlert            = "netStatus"
	K8sEventAlert       = "k8sStatus"
	REDMetricsAlert     = "REDStatus"
	LogMetricsAlert     = "logsStatus"
)

type AlertStatus struct {
	InfrastructureStatus string `json:"infrastructureStatus"` // 基础设施告警
	NetStatus            string `json:"netStatus"`            // 网络告警
	K8sStatus            string `json:"k8sStatus"`            // K8s状态告警
}

type AlertReason map[string][]AlertDetail

type AlertDetail struct {
	Timestamp    int64  `json:"time"`
	AlertObject  string `json:"alertObject"`
	AlertReason  string `json:"alertReason"`
	AlertMessage string `json:"alertMessage"`
}

func (r AlertReason) Add(key string, detail AlertDetail) {
	if len(key) == 0 {
		return
	}

	if r == nil {
		return
	}

	details, find := r[key]
	if !find {
		r[key] = []AlertDetail{detail}
		return
	}

	r[key] = append(details, detail)
}
