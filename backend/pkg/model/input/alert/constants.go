package alert

const (
	JSONType       string = "json"
	PrometheusType string = "prometheus"
	ZabbixType     string = "zabbix"
)

const (
	StatusFiring   = "firing"
	StatusResolved = "resolved"
)

const (
	SeverityCriticalLevel = "critical"
	SeverityErrorLevel    = "error"
	SeverityWarnLevel     = "warn"
	SeverityInfoLevel     = "info"
	SeverityUnknownLevel  = "unknown"
)

const (
	ZabbixSeverityDisaster = "Disaster"
	ZabbixSeverityHigh     = "High"
	ZabbixSeverityAverage  = "Average"
	ZabbixSeverityWarning  = "Warning"
	ZabbixSeverityInfo     = "Information"
	ZabbixSeverityUnknown  = "Not classified"
)

const (
	ZabbixStatusOK      = "OK"
	ZabbixStatusProblem = "PROBLEM"
)
