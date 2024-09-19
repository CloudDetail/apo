package model

const (
	DelaySourceAlert    = "delaySource"
	InfrastructureAlert = "infrastructureStatus"
	NetAlert            = "netStatus"
	K8sEventAlert       = "k8sStatus"
	REDMetricsAlert     = "REDStatus"
	LogMetricsAlert     = "logsStatus"
	AppAlert            = "appStatus"
	ContainerAlert      = "containerStaus"

	UndefinedAlert = "undefinedAlert"
)

type AlertStatusPROM struct {
	LogMetricsStatus string `json:"logsStatus"` // 日志指标告警
}

func (s *AlertStatusPROM) IsAllNormal() bool {
	return s.LogMetricsStatus == STATUS_NORMAL
}

type AlertStatusCH struct {
	InfrastructureStatus string `json:"infrastructureStatus"` // 基础设施告警
	NetStatus            string `json:"netStatus"`            // 网络告警
	K8sStatus            string `json:"k8sStatus"`            // K8s状态告警

	AppStatus       string `json:"appStatus"`       // 应用告警
	ContainerStatus string `json:"containerStatus"` // 容器告警
}

func (s *AlertStatusCH) IsAllNormal() bool {
	return s.InfrastructureStatus == STATUS_NORMAL &&
		s.NetStatus == STATUS_NORMAL &&
		s.K8sStatus == STATUS_NORMAL &&
		s.AppStatus == STATUS_NORMAL &&
		s.ContainerStatus == STATUS_NORMAL
}

var NORMAL_ALERT_STATUS = AlertStatus{
	AlertStatusCH: AlertStatusCH{
		InfrastructureStatus: STATUS_NORMAL,
		NetStatus:            STATUS_NORMAL,
		K8sStatus:            STATUS_NORMAL,
		AppStatus:            STATUS_NORMAL,
		ContainerStatus:      STATUS_NORMAL,
	},
	AlertStatusPROM: AlertStatusPROM{
		LogMetricsStatus: STATUS_NORMAL,
	},
}

type AlertStatus struct {
	AlertStatusCH
	AlertStatusPROM
}

func (s *AlertStatus) IsAllNormal() bool {
	return s.AlertStatusCH.IsAllNormal() && s.AlertStatusPROM.IsAllNormal()
}

type AlertReason map[string][]AlertDetail

type AlertDetail struct {
	Timestamp    int64  `json:"timestamp"`
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

type AlertEventLevelCountMap map[string]AlertEventLevelCount

type AlertEventLevelCount map[SeverityLevel]int

func (m AlertEventLevelCountMap) Add(key string, level SeverityLevel) {
	counts, find := m[key]
	if !find {
		counts := make(AlertEventLevelCount)
		counts[level] = 1
		m[key] = counts
		return
	}

	count, find := counts[level]
	if !find {
		counts[level] = 1
		m[key] = counts
		return
	}

	counts[level] = count + 1
	m[key] = counts
}
