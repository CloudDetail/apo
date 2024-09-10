package serviceoverview

import (
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// RES_MAX_VALUE 返回前端的最大值，同比为该值时表示最大值
const RES_MAX_VALUE = 9999999

type EndpointMetrics struct {
	prom.EndpointKey

	DelaySource   *float64 //延时主要来源
	AlertCount    int
	NamespaceList []string // 包含该端点的Namespace

	// TODO DelaySource值为nil和值为0是两种场景。
	//  nil表示没有查询到数据（可能没有这个指标），显示未知；0表示无网络占比，显示自身
	IsLatencyExceeded   bool
	IsErrorRateExceeded bool
	IsTPSExceeded       bool

	Avg1MinLatencyMutationRate float64 //延时突变率
	Avg1MinErrorMutationRate   float64 //错误率突变率

	prom.REDMetrics

	LatencyData   []prom.Points // 延时时间段的数据
	ErrorRateData []prom.Points // 错误率时间段的数据
	TPMData       []prom.Points // TPM 时间段的数据
}
type ServiceDetail struct {
	ServiceName          string
	EndpointCount        int
	ServiceSize          int
	Endpoints            []*EndpointMetrics
	Instances            []Instance
	LogData              []prom.Points // 日志告警次数 30min的数据
	InfrastructureStatus string
	NetStatus            string
	K8sStatus            string
	Timestamp            int64
}

const (
	POD = iota
	CONTAINER
	VM
)

type Instance struct {
	InstanceName           string //实例名
	ContentKey             string // URL
	ConvertName            string
	SvcName                string //url所属的服务名
	Count                  int
	InstanceType           int
	ContainerId            string
	Pid                    string
	IsLatencyDODExceeded   bool
	IsErrorRateDODExceeded bool
	IsTPSDODExceeded       bool
	IsLatencyWOWExceeded   bool
	IsErrorRateWOWExceeded bool
	IsTPSWOWExceeded       bool
	Pod                    string
	Namespace              string
	NodeName               string
	AvgLatency             *float64      // 30min内的平均延时时间
	LatencyDayOverDay      *float64      // 延时日同比
	LatencyWeekOverWeek    *float64      // 延时周同比
	LatencyData            []prom.Points // 延时30min的数据

	AvgErrorRate          *float64      // 30min内的平均错误率
	ErrorRateDayOverDay   *float64      // 错误率日同比
	ErrorRateWeekOverWeek *float64      // 错误率周同比
	ErrorRateData         []prom.Points // 错误率30min的数据

	AvgTPS          *float64      // 30min内的平均TPS
	TPSDayOverDay   *float64      // TPS日同比
	TPSWeekOverWeek *float64      // TPS周同比
	TPSData         []prom.Points // TPS 30min的数据

	AvgLog          *float64      // 30min内的日志告警次数
	LogDayOverDay   *float64      // 日志告警次数日同比
	LogWeekOverWeek *float64      // 日志告警次数周同比
	LogData         []prom.Points // 日志告警次数 30min的数据
}
