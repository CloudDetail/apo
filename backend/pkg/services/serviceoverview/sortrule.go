package serviceoverview

import (
	"sort"
)

// 跟据日同比阈值进行排序,错误率同比相同，比较错误率值
func sortByDODThreshold(endpoints []*EndpointMetrics) {
	sort.SliceStable(endpoints, func(i, j int) bool {
		//先按照count排序
		if endpoints[i].AlertCount != endpoints[j].AlertCount {
			return endpoints[i].AlertCount > endpoints[j].AlertCount
		}
		//等于3时按照错误率排序
		if endpoints[i].AlertCount == endpoints[j].AlertCount && endpoints[i].AlertCount == 3 {
			if endpoints[i].DOD.ErrorRate != nil && endpoints[j].DOD.ErrorRate != nil && endpoints[i].DOD.ErrorRate != endpoints[j].DOD.ErrorRate {
				if *endpoints[i].DOD.ErrorRate != *endpoints[j].DOD.ErrorRate {
					return *endpoints[i].DOD.ErrorRate > *endpoints[j].DOD.ErrorRate
				}
				if *endpoints[i].DOD.ErrorRate == *endpoints[j].DOD.ErrorRate && endpoints[i].Avg.ErrorRate != nil && endpoints[j].Avg.ErrorRate != nil {
					return *endpoints[i].Avg.ErrorRate > *endpoints[j].Avg.ErrorRate
				}
			}
			if endpoints[i].DOD.Latency != nil && endpoints[j].DOD.Latency != nil && endpoints[i].DOD.Latency != endpoints[j].DOD.Latency {
				return *endpoints[i].DOD.Latency > *endpoints[j].DOD.Latency
			}
			if endpoints[i].DOD.TPM != nil && endpoints[j].DOD.TPM != nil {
				return *endpoints[i].DOD.TPM > *endpoints[j].DOD.TPM
			}
		}
		//count = 2的比较方式
		if endpoints[i].AlertCount == endpoints[j].AlertCount && endpoints[i].AlertCount == 2 {
			if endpoints[i].IsErrorRateExceeded == true && endpoints[j].IsErrorRateExceeded == false {
				return true
			}
			if endpoints[i].IsLatencyExceeded == true && endpoints[j].IsLatencyExceeded == false && endpoints[i].IsErrorRateExceeded == endpoints[j].IsErrorRateExceeded {
				return true
			}
			if endpoints[i].IsErrorRateExceeded == endpoints[j].IsErrorRateExceeded && endpoints[j].IsErrorRateExceeded == false {
				if endpoints[i].DOD.Latency != nil && endpoints[j].DOD.Latency != nil && endpoints[i].DOD.Latency != endpoints[j].DOD.Latency {
					return *endpoints[i].DOD.Latency > *endpoints[j].DOD.Latency
				}
				if endpoints[i].DOD.TPM != nil && endpoints[j].DOD.TPM != nil {
					return *endpoints[i].DOD.TPM > *endpoints[j].DOD.TPM
				}
			}

			if endpoints[i].IsLatencyExceeded == endpoints[j].IsLatencyExceeded && endpoints[j].IsLatencyExceeded == false {
				if endpoints[i].DOD.ErrorRate != nil && endpoints[j].DOD.ErrorRate != nil && endpoints[i].DOD.ErrorRate != endpoints[j].DOD.ErrorRate {
					if *endpoints[i].DOD.ErrorRate != *endpoints[j].DOD.ErrorRate {
						return *endpoints[i].DOD.ErrorRate > *endpoints[j].DOD.ErrorRate
					}
					if *endpoints[i].DOD.ErrorRate == *endpoints[j].DOD.ErrorRate && endpoints[i].Avg.ErrorRate != nil && endpoints[j].Avg.ErrorRate != nil {
						return *endpoints[i].Avg.ErrorRate > *endpoints[j].Avg.ErrorRate
					}
				}
				if endpoints[i].DOD.TPM != nil && endpoints[j].DOD.TPM != nil {
					return *endpoints[i].DOD.TPM > *endpoints[j].DOD.TPM
				}
			}
			if endpoints[i].IsTPSExceeded == endpoints[j].IsTPSExceeded && endpoints[j].IsTPSExceeded == false {
				if endpoints[i].DOD.ErrorRate != nil && endpoints[j].DOD.ErrorRate != nil && endpoints[i].DOD.ErrorRate != endpoints[j].DOD.ErrorRate {
					if *endpoints[i].DOD.ErrorRate != *endpoints[j].DOD.ErrorRate {
						return *endpoints[i].DOD.ErrorRate > *endpoints[j].DOD.ErrorRate
					}
					if *endpoints[i].DOD.ErrorRate == *endpoints[j].DOD.ErrorRate && endpoints[i].Avg.ErrorRate != nil && endpoints[j].Avg.ErrorRate != nil {
						return *endpoints[i].Avg.ErrorRate > *endpoints[j].Avg.ErrorRate
					}
				}
				if endpoints[i].DOD.Latency != nil && endpoints[j].DOD.Latency != nil && endpoints[i].DOD.Latency != endpoints[j].DOD.Latency {
					return *endpoints[i].DOD.Latency > *endpoints[j].DOD.Latency
				}

				if endpoints[i].DOD.TPM != nil && endpoints[j].DOD.TPM != nil {
					return *endpoints[i].DOD.TPM > *endpoints[j].DOD.TPM
				}
			}

		}
		if endpoints[i].AlertCount == endpoints[j].AlertCount && endpoints[i].AlertCount == 1 {
			if endpoints[i].IsErrorRateExceeded == true && endpoints[j].IsErrorRateExceeded == false {
				return true
			}
			if endpoints[i].IsLatencyExceeded == true && endpoints[j].IsLatencyExceeded == false && endpoints[i].IsErrorRateExceeded == endpoints[j].IsErrorRateExceeded {
				return true
			}
			if endpoints[i].DOD.ErrorRate != nil && endpoints[j].DOD.ErrorRate != nil && endpoints[i].IsErrorRateExceeded == endpoints[j].IsErrorRateExceeded && endpoints[j].IsErrorRateExceeded == true {
				if *endpoints[i].DOD.ErrorRate != *endpoints[j].DOD.ErrorRate {
					return *endpoints[i].DOD.ErrorRate > *endpoints[j].DOD.ErrorRate
				}
				if *endpoints[i].DOD.ErrorRate == *endpoints[j].DOD.ErrorRate && endpoints[i].Avg.ErrorRate != nil && endpoints[j].Avg.ErrorRate != nil {
					return *endpoints[i].Avg.ErrorRate > *endpoints[j].Avg.ErrorRate
				}

			}
			if endpoints[i].DOD.Latency != nil && endpoints[j].DOD.Latency != nil && endpoints[i].IsLatencyExceeded == endpoints[j].IsLatencyExceeded && endpoints[j].IsLatencyExceeded == true {
				return *endpoints[i].DOD.Latency > *endpoints[j].DOD.Latency
			}
			if endpoints[i].DOD.TPM != nil && endpoints[j].DOD.TPM != nil && endpoints[i].IsTPSExceeded == endpoints[j].IsTPSExceeded && endpoints[j].IsTPSExceeded == true {
				return *endpoints[i].DOD.TPM > *endpoints[j].DOD.TPM
			}
		}
		if endpoints[i].AlertCount == endpoints[j].AlertCount && endpoints[i].AlertCount == 0 {
			if endpoints[i].DOD.ErrorRate != nil && endpoints[j].DOD.ErrorRate == nil {
				return true
			}
			if endpoints[i].DOD.ErrorRate == endpoints[j].DOD.ErrorRate && endpoints[i].DOD.Latency != nil && endpoints[j].DOD.Latency == nil {
				return true
			}
			if endpoints[i].DOD.ErrorRate == endpoints[j].DOD.ErrorRate && endpoints[i].DOD.Latency == endpoints[j].DOD.Latency && endpoints[i].DOD.TPM != nil && endpoints[j].DOD.TPM == nil {
				return true
			}
			if endpoints[i].DOD.ErrorRate != nil && endpoints[j].DOD.ErrorRate != nil && endpoints[i].DOD.ErrorRate != endpoints[j].DOD.ErrorRate {
				if *endpoints[i].DOD.ErrorRate != *endpoints[j].DOD.ErrorRate {
					return *endpoints[i].DOD.ErrorRate > *endpoints[j].DOD.ErrorRate
				}
				if *endpoints[i].DOD.ErrorRate == *endpoints[j].DOD.ErrorRate && endpoints[i].Avg.ErrorRate != nil && endpoints[j].Avg.ErrorRate != nil {
					return *endpoints[i].Avg.ErrorRate > *endpoints[j].Avg.ErrorRate
				}
			}
			if endpoints[i].DOD.Latency != nil && endpoints[j].DOD.Latency != nil && endpoints[i].DOD.Latency != endpoints[j].DOD.Latency {
				return *endpoints[i].DOD.Latency > *endpoints[j].DOD.Latency
			}
			if endpoints[i].DOD.TPM != nil && endpoints[j].DOD.TPM != nil && endpoints[i].DOD.TPM != endpoints[j].DOD.TPM {
				return *endpoints[i].DOD.TPM > *endpoints[j].DOD.TPM
			}

		}

		return endpoints[i].AlertCount > endpoints[j].AlertCount
	})
}

// 突变排序
func sortByMutation(endpoints []*EndpointMetrics) {
	for i, _ := range endpoints {
		//平均错误率和1m错误率都查不出来，突变率为0
		if endpoints[i].Avg.ErrorRate == nil && endpoints[i].Realtime.ErrorRate == nil {
			endpoints[i].Avg1MinErrorMutationRate = 0
		}
		//平均错误率查的出来，1m错误率查不出来
		if endpoints[i].Avg.ErrorRate != nil && endpoints[i].Realtime.ErrorRate == nil {
			//平均错误率为0 ：突变率为0
			if endpoints[i].Avg.ErrorRate != nil && *endpoints[i].Avg.ErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = 0
			}
			//平均错误率不为0，突变率为-1
			if endpoints[i].Avg.ErrorRate != nil && *endpoints[i].Avg.ErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = -1
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].Avg.ErrorRate == nil && endpoints[i].Realtime.ErrorRate != nil {
			//1m错误率为0，突变率为0
			if endpoints[i].Realtime.ErrorRate != nil && *endpoints[i].Realtime.ErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = 0
			}
			//1m错误率不为0，突变率为max
			if endpoints[i].Realtime.ErrorRate != nil && *endpoints[i].Realtime.ErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].Avg.ErrorRate != nil && endpoints[i].Realtime.ErrorRate != nil {
			//1m错误率为0，突变率为0
			if endpoints[i].Avg.ErrorRate != nil && *endpoints[i].Avg.ErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
			//1m错误率不为0，突变率为max
			if endpoints[i].Avg.ErrorRate != nil && endpoints[i].Realtime.ErrorRate != nil && *endpoints[i].Avg.ErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = *endpoints[i].Realtime.ErrorRate / *endpoints[i].Avg.ErrorRate
			}
		}
		//latency
		//平均延时和1m延时都查不出来，突变率为0(不可能的情况)
		if endpoints[i].Avg.Latency == nil && endpoints[i].Realtime.Latency == nil {
			endpoints[i].Avg1MinLatencyMutationRate = 0
		}
		//平均延时查的出来，1m延时查不出来
		if endpoints[i].Avg.Latency != nil && endpoints[i].Realtime.Latency == nil {
			//平均延时为0 ：突变率为0
			if endpoints[i].Avg.Latency != nil && *endpoints[i].Avg.Latency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = 0
			}
			//平均延时不为0，突变率为-1
			if endpoints[i].Avg.Latency != nil && *endpoints[i].Avg.Latency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = -1
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].Avg.Latency == nil && endpoints[i].Realtime.Latency != nil {
			//1m延时为0，突变率为0
			if endpoints[i].Realtime.Latency != nil && *endpoints[i].Realtime.Latency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = 0
			}
			//1m延时不为0，突变率为max
			if endpoints[i].Realtime.Latency != nil && *endpoints[i].Realtime.Latency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].Avg.Latency != nil && endpoints[i].Realtime.Latency != nil {
			//平均延时为0，突变率为max
			if endpoints[i].Avg.Latency != nil && *endpoints[i].Avg.Latency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
			//平均延时不为0，突变率为1m延时/平均延时
			if endpoints[i].Avg.Latency != nil && endpoints[i].Realtime.Latency != nil && *endpoints[i].Avg.Latency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = *endpoints[i].Realtime.Latency / *endpoints[i].Avg.Latency
			}
		}
		//if urls[i].Realtime.ErrorRate != nil {
		//	a := *urls[i].Realtime.ErrorRate
		//	fmt.Printf("1minError:%v,iminLatency:%v\n", a, urls[i].Avg1minLatency)
		//} else {
		//	a := 10
		//	fmt.Printf("1minError:%v,iminLatency:%v\n", a, urls[i].Avg1minLatency)
		//}
		//if urls[i].Avg.ErrorRate != nil {
		//	a := *urls[i].Avg.ErrorRate
		//	fmt.Printf("Error:%v\n", a)
		//} else {
		//	a := 10
		//	fmt.Printf("Error:%v\n", a)
		//}
	}
	sort.SliceStable(endpoints, func(i, j int) bool {
		// Case 1: 如果有一个错误率突变率大于1（错误率上升）
		if endpoints[i].Avg1MinErrorMutationRate > 1 || endpoints[j].Avg1MinErrorMutationRate > 1 {
			return endpoints[i].Avg1MinErrorMutationRate > endpoints[j].Avg1MinErrorMutationRate
		}

		// Case 2: 如果错误率突变率都小于等于1
		if endpoints[i].Avg1MinErrorMutationRate <= 1 && endpoints[j].Avg1MinErrorMutationRate <= 1 {
			// 优先按延迟突变率排序，较大的排在前面
			if endpoints[i].Avg1MinLatencyMutationRate != endpoints[j].Avg1MinLatencyMutationRate {
				return endpoints[i].Avg1MinLatencyMutationRate > endpoints[j].Avg1MinLatencyMutationRate
			}

			// 如果延迟突变率相同，按错误率排序
			if endpoints[i].Realtime.ErrorRate != nil && endpoints[j].Realtime.ErrorRate != nil {
				return *endpoints[i].Realtime.ErrorRate > *endpoints[j].Realtime.ErrorRate
			}
			// 如果一个错误率为nil，另一个不为nil，错误率不为nil的排在前面
			if endpoints[i].Realtime.ErrorRate != nil && endpoints[j].Realtime.ErrorRate == nil {
				return true
			}
			if endpoints[i].Realtime.ErrorRate == nil && endpoints[j].Realtime.ErrorRate != nil {
				return false
			}
			// 如果延迟突变率相同，按错误率排序
			if endpoints[i].Avg.ErrorRate != nil && endpoints[j].Avg.ErrorRate != nil {
				return *endpoints[i].Avg.ErrorRate > *endpoints[j].Avg.ErrorRate
			}
			// 如果一个错误率为nil，另一个不为nil，错误率不为nil的排在前面
			if endpoints[i].Avg.ErrorRate != nil && endpoints[j].Avg.ErrorRate == nil {
				return true
			}
			if endpoints[i].Avg.ErrorRate == nil && endpoints[j].Avg.ErrorRate != nil {
				return false
			}
			// 如果错误率相同或都为nil，按延迟排序
			if endpoints[i].Realtime.Latency != nil && endpoints[j].Realtime.Latency != nil {
				return *endpoints[i].Realtime.Latency > *endpoints[j].Realtime.Latency
			}
			if endpoints[i].Realtime.Latency != nil && endpoints[j].Realtime.Latency == nil {
				return true
			}
			if endpoints[i].Realtime.Latency == nil && endpoints[j].Realtime.Latency != nil {
				return false
			}
			// 如果错误率相同或都为nil，按延迟排序
			if endpoints[i].Avg.Latency != nil && endpoints[j].Avg.Latency != nil {
				return *endpoints[i].Avg.Latency > *endpoints[j].Avg.Latency
			}
			if endpoints[i].Avg.Latency != nil && endpoints[j].Avg.Latency == nil {
				return true
			}
			if endpoints[i].Avg.Latency == nil && endpoints[j].Avg.Latency != nil {
				return false
			}
		}
		return false
	})

}

func fillServices(endpoints []*EndpointMetrics) []ServiceDetail {
	var services []ServiceDetail
	for _, url := range endpoints {
		//如果没有数据则不返回
		if (url.Avg.Latency == nil && url.Avg.ErrorRate == nil) || (url.Avg.Latency == nil && url.Avg.ErrorRate != nil && *url.Avg.ErrorRate == 0 && url.Avg.TPM == nil) {
			continue
		}
		serviceName := url.SvcName
		found := false
		for j, _ := range services {
			if services[j].ServiceName == serviceName {
				found = true
				services[j].EndpointCount++
				if services[j].ServiceSize < 3 {
					services[j].Endpoints = append(services[j].Endpoints, url)
					services[j].ServiceSize++
					break
				}
			}
		}
		if !found {
			newService := ServiceDetail{
				ServiceName:   serviceName,
				ServiceSize:   1,
				EndpointCount: 1,
				Endpoints:     []*EndpointMetrics{url},
			}
			services = append(services, newService)
		}
	}

	return services
}

func fillOneService(endpoints []*EndpointMetrics) []ServiceDetail {
	var service []ServiceDetail
	for _, url := range endpoints {
		serviceName := url.SvcName
		found := false
		for j, _ := range service {
			if service[j].ServiceName == serviceName {
				found = true
				service[j].EndpointCount++
				service[j].Endpoints = append(service[j].Endpoints, url)
				service[j].ServiceSize++
				break
			}
		}
		if !found {
			newService := ServiceDetail{
				ServiceName:   serviceName,
				ServiceSize:   1,
				EndpointCount: 1,
				Endpoints:     []*EndpointMetrics{url},
			}
			service = append(service, newService)
		}
	}

	return service
}
