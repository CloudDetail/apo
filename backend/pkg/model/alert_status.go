package model

type AlertStatus struct {
	InfrastructureStatus string `json:"infrastructureStatus"` // 基础设施告警
	NetStatus            string `json:"netStatus"`            // 网络告警
	K8sStatus            string `json:"k8sStatus"`            // K8s状态告警
}

type AlertReason map[string]string

func (r AlertReason) Add(key, value string) {
	if len(value) == 0 || len(key) == 0 {
		return
	}

	if r == nil {
		return
	}

	reason, find := r[key]
	if !find {
		r[key] = value
		return
	}

	r[key] = reason + "\n" + value
}
