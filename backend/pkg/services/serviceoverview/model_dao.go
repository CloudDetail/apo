package serviceoverview

import "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"

// RES_MAX_VALUE 返回前端的最大值，同比为该值时表示最大值
const RES_MAX_VALUE = 9999999

type Endpoint struct {
	ContentKey          string   // URL
	SvcName             string   // url所属的服务名
	NamespaceList       []string // 包含该端点的Namespace
	Count               int
	DelaySource         *float64 //延时主要来源
	IsLatencyExceeded   bool
	IsErrorRateExceeded bool
	IsTPSExceeded       bool

	Avg1minLatency             *float64 //最近一分钟之内的平均延时
	Avg1minErrorRate           *float64 //最近一分钟之内的错误率
	Avg1MinLatencyMutationRate float64  //延时突变率
	Avg1MinErrorMutationRate   float64  //错误率突变率

	AvgLatency          *float64            // 时间段内的平均延时时间
	LatencyDayOverDay   *float64            // 延时日同比
	LatencyWeekOverWeek *float64            // 延时周同比
	LatencyData         []prometheus.Points // 延时时间段的数据

	AvgErrorRate          *float64            // 时间段内的平均错误率
	ErrorRateDayOverDay   *float64            // 错误率日同比
	ErrorRateWeekOverWeek *float64            // 错误率周同比
	ErrorRateData         []prometheus.Points // 错误率时间段的数据

	AvgTPM          *float64            // 时间段内的平均每分钟请求数
	TPMDayOverDay   *float64            // TPM日同比
	TPMWeekOverWeek *float64            // TPM周同比
	TPMData         []prometheus.Points // TPM 时间段的数据

}
type serviceDetail struct {
	ServiceName          string
	EndpointCount        int
	ServiceSize          int
	Endpoints            []*Endpoint
	Instances            []Instance
	LogData              []prometheus.Points // 日志告警次数 30min的数据
	InfrastructureStatus string
	NetStatus            string
	K8sStatus            string
	Timestamp            int64
}

const (
	POD = iota
	NODE
	VM
)

type Instance struct {
	InstanceName           string //实例名
	ContentKey             string // URL
	ConvertName            string
	SvcName                string //url所属的服务名
	Count                  int
	InstanceType           int
	ServerAddress          string
	ContainerId            string
	Pid                    string
	DelaySource            *float64 //延时主要来源
	IsLatencyDODExceeded   bool
	IsErrorRateDODExceeded bool
	IsTPSDODExceeded       bool
	IsLatencyWOWExceeded   bool
	IsErrorRateWOWExceeded bool
	IsTPSWOWExceeded       bool
	Pod                    string
	Namespace              string
	NodeName               string
	AvgLatency             *float64            // 30min内的平均延时时间
	LatencyDayOverDay      *float64            // 延时日同比
	LatencyWeekOverWeek    *float64            // 延时周同比
	LatencyData            []prometheus.Points // 延时30min的数据

	AvgErrorRate          *float64            // 30min内的平均错误率
	ErrorRateDayOverDay   *float64            // 错误率日同比
	ErrorRateWeekOverWeek *float64            // 错误率周同比
	ErrorRateData         []prometheus.Points // 错误率30min的数据

	AvgTPS          *float64            // 30min内的平均TPS
	TPSDayOverDay   *float64            // TPS日同比
	TPSWeekOverWeek *float64            // TPS周同比
	TPSData         []prometheus.Points // TPS 30min的数据

	AvgLog          *float64            // 30min内的日志告警次数
	LogDayOverDay   *float64            // 日志告警次数日同比
	LogWeekOverWeek *float64            // 日志告警次数周同比
	LogData         []prometheus.Points // 日志告警次数 30min的数据
}
