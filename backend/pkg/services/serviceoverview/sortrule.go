// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"sort"

	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// 跟据日同比阈值进行排序,错误率同比相同，比较错误率值
func sortByDODThreshold(endpoints []*prom.EndpointMetrics) {
	sort.SliceStable(endpoints, func(i, j int) bool {
		//先按照count排序
		if endpoints[i].AlertCount != endpoints[j].AlertCount {
			return endpoints[i].AlertCount > endpoints[j].AlertCount
		}
		//等于3时按照错误率排序
		if endpoints[i].AlertCount == endpoints[j].AlertCount && endpoints[i].AlertCount == 3 {
			if endpoints[i].REDMetrics.DOD.ErrorRate != nil && endpoints[j].REDMetrics.DOD.ErrorRate != nil && endpoints[i].REDMetrics.DOD.ErrorRate != endpoints[j].REDMetrics.DOD.ErrorRate {
				if *endpoints[i].REDMetrics.DOD.ErrorRate != *endpoints[j].REDMetrics.DOD.ErrorRate {
					return *endpoints[i].REDMetrics.DOD.ErrorRate > *endpoints[j].REDMetrics.DOD.ErrorRate
				}
				if *endpoints[i].REDMetrics.DOD.ErrorRate == *endpoints[j].REDMetrics.DOD.ErrorRate && endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[j].REDMetrics.Avg.ErrorRate != nil {
					return *endpoints[i].REDMetrics.Avg.ErrorRate > *endpoints[j].REDMetrics.Avg.ErrorRate
				}
			}
			if endpoints[i].REDMetrics.DOD.Latency != nil && endpoints[j].REDMetrics.DOD.Latency != nil && endpoints[i].REDMetrics.DOD.Latency != endpoints[j].REDMetrics.DOD.Latency {
				return *endpoints[i].REDMetrics.DOD.Latency > *endpoints[j].REDMetrics.DOD.Latency
			}
			if endpoints[i].REDMetrics.DOD.TPM != nil && endpoints[j].REDMetrics.DOD.TPM != nil {
				return *endpoints[i].REDMetrics.DOD.TPM > *endpoints[j].REDMetrics.DOD.TPM
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
				if endpoints[i].REDMetrics.DOD.Latency != nil && endpoints[j].REDMetrics.DOD.Latency != nil && endpoints[i].REDMetrics.DOD.Latency != endpoints[j].REDMetrics.DOD.Latency {
					return *endpoints[i].REDMetrics.DOD.Latency > *endpoints[j].REDMetrics.DOD.Latency
				}
				if endpoints[i].REDMetrics.DOD.TPM != nil && endpoints[j].REDMetrics.DOD.TPM != nil {
					return *endpoints[i].REDMetrics.DOD.TPM > *endpoints[j].REDMetrics.DOD.TPM
				}
			}

			if endpoints[i].IsLatencyExceeded == endpoints[j].IsLatencyExceeded && endpoints[j].IsLatencyExceeded == false {
				if endpoints[i].REDMetrics.DOD.ErrorRate != nil && endpoints[j].REDMetrics.DOD.ErrorRate != nil && endpoints[i].REDMetrics.DOD.ErrorRate != endpoints[j].REDMetrics.DOD.ErrorRate {
					if *endpoints[i].REDMetrics.DOD.ErrorRate != *endpoints[j].REDMetrics.DOD.ErrorRate {
						return *endpoints[i].REDMetrics.DOD.ErrorRate > *endpoints[j].REDMetrics.DOD.ErrorRate
					}
					if *endpoints[i].REDMetrics.DOD.ErrorRate == *endpoints[j].REDMetrics.DOD.ErrorRate && endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[j].REDMetrics.Avg.ErrorRate != nil {
						return *endpoints[i].REDMetrics.Avg.ErrorRate > *endpoints[j].REDMetrics.Avg.ErrorRate
					}
				}
				if endpoints[i].REDMetrics.DOD.TPM != nil && endpoints[j].REDMetrics.DOD.TPM != nil {
					return *endpoints[i].REDMetrics.DOD.TPM > *endpoints[j].REDMetrics.DOD.TPM
				}
			}
			if endpoints[i].IsTPSExceeded == endpoints[j].IsTPSExceeded && endpoints[j].IsTPSExceeded == false {
				if endpoints[i].REDMetrics.DOD.ErrorRate != nil && endpoints[j].REDMetrics.DOD.ErrorRate != nil && endpoints[i].REDMetrics.DOD.ErrorRate != endpoints[j].REDMetrics.DOD.ErrorRate {
					if *endpoints[i].REDMetrics.DOD.ErrorRate != *endpoints[j].REDMetrics.DOD.ErrorRate {
						return *endpoints[i].REDMetrics.DOD.ErrorRate > *endpoints[j].REDMetrics.DOD.ErrorRate
					}
					if *endpoints[i].REDMetrics.DOD.ErrorRate == *endpoints[j].REDMetrics.DOD.ErrorRate && endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[j].REDMetrics.Avg.ErrorRate != nil {
						return *endpoints[i].REDMetrics.Avg.ErrorRate > *endpoints[j].REDMetrics.Avg.ErrorRate
					}
				}
				if endpoints[i].REDMetrics.DOD.Latency != nil && endpoints[j].REDMetrics.DOD.Latency != nil && endpoints[i].REDMetrics.DOD.Latency != endpoints[j].REDMetrics.DOD.Latency {
					return *endpoints[i].REDMetrics.DOD.Latency > *endpoints[j].REDMetrics.DOD.Latency
				}

				if endpoints[i].REDMetrics.DOD.TPM != nil && endpoints[j].REDMetrics.DOD.TPM != nil {
					return *endpoints[i].REDMetrics.DOD.TPM > *endpoints[j].REDMetrics.DOD.TPM
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
			if endpoints[i].REDMetrics.DOD.ErrorRate != nil && endpoints[j].REDMetrics.DOD.ErrorRate != nil && endpoints[i].IsErrorRateExceeded == endpoints[j].IsErrorRateExceeded && endpoints[j].IsErrorRateExceeded == true {
				if *endpoints[i].REDMetrics.DOD.ErrorRate != *endpoints[j].REDMetrics.DOD.ErrorRate {
					return *endpoints[i].REDMetrics.DOD.ErrorRate > *endpoints[j].REDMetrics.DOD.ErrorRate
				}
				if *endpoints[i].REDMetrics.DOD.ErrorRate == *endpoints[j].REDMetrics.DOD.ErrorRate && endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[j].REDMetrics.Avg.ErrorRate != nil {
					return *endpoints[i].REDMetrics.Avg.ErrorRate > *endpoints[j].REDMetrics.Avg.ErrorRate
				}

			}
			if endpoints[i].REDMetrics.DOD.Latency != nil && endpoints[j].REDMetrics.DOD.Latency != nil && endpoints[i].IsLatencyExceeded == endpoints[j].IsLatencyExceeded && endpoints[j].IsLatencyExceeded == true {
				return *endpoints[i].REDMetrics.DOD.Latency > *endpoints[j].REDMetrics.DOD.Latency
			}
			if endpoints[i].REDMetrics.DOD.TPM != nil && endpoints[j].REDMetrics.DOD.TPM != nil && endpoints[i].IsTPSExceeded == endpoints[j].IsTPSExceeded && endpoints[j].IsTPSExceeded == true {
				return *endpoints[i].REDMetrics.DOD.TPM > *endpoints[j].REDMetrics.DOD.TPM
			}
		}
		if endpoints[i].AlertCount == endpoints[j].AlertCount && endpoints[i].AlertCount == 0 {
			if endpoints[i].REDMetrics.DOD.ErrorRate != nil && endpoints[j].REDMetrics.DOD.ErrorRate == nil {
				return true
			}
			if endpoints[i].REDMetrics.DOD.ErrorRate == endpoints[j].REDMetrics.DOD.ErrorRate && endpoints[i].REDMetrics.DOD.Latency != nil && endpoints[j].REDMetrics.DOD.Latency == nil {
				return true
			}
			if endpoints[i].REDMetrics.DOD.ErrorRate == endpoints[j].REDMetrics.DOD.ErrorRate && endpoints[i].REDMetrics.DOD.Latency == endpoints[j].REDMetrics.DOD.Latency && endpoints[i].REDMetrics.DOD.TPM != nil && endpoints[j].REDMetrics.DOD.TPM == nil {
				return true
			}
			if endpoints[i].REDMetrics.DOD.ErrorRate != nil && endpoints[j].REDMetrics.DOD.ErrorRate != nil && endpoints[i].REDMetrics.DOD.ErrorRate != endpoints[j].REDMetrics.DOD.ErrorRate {
				if *endpoints[i].REDMetrics.DOD.ErrorRate != *endpoints[j].REDMetrics.DOD.ErrorRate {
					return *endpoints[i].REDMetrics.DOD.ErrorRate > *endpoints[j].REDMetrics.DOD.ErrorRate
				}
				if *endpoints[i].REDMetrics.DOD.ErrorRate == *endpoints[j].REDMetrics.DOD.ErrorRate && endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[j].REDMetrics.Avg.ErrorRate != nil {
					return *endpoints[i].REDMetrics.Avg.ErrorRate > *endpoints[j].REDMetrics.Avg.ErrorRate
				}
			}
			if endpoints[i].REDMetrics.DOD.Latency != nil && endpoints[j].REDMetrics.DOD.Latency != nil && endpoints[i].REDMetrics.DOD.Latency != endpoints[j].REDMetrics.DOD.Latency {
				return *endpoints[i].REDMetrics.DOD.Latency > *endpoints[j].REDMetrics.DOD.Latency
			}
			if endpoints[i].REDMetrics.DOD.TPM != nil && endpoints[j].REDMetrics.DOD.TPM != nil && endpoints[i].REDMetrics.DOD.TPM != endpoints[j].REDMetrics.DOD.TPM {
				return *endpoints[i].REDMetrics.DOD.TPM > *endpoints[j].REDMetrics.DOD.TPM
			}

		}

		return endpoints[i].AlertCount > endpoints[j].AlertCount
	})
}

// 突变排序
func sortByMutation(endpoints []*prom.EndpointMetrics) {
	for i, _ := range endpoints {
		//平均错误率和1m错误率都查不出来，突变率为0
		if endpoints[i].REDMetrics.Avg.ErrorRate == nil && endpoints[i].REDMetrics.Realtime.ErrorRate == nil {
			endpoints[i].Avg1MinErrorMutationRate = 0
		}
		//平均错误率查的出来，1m错误率查不出来
		if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[i].REDMetrics.Realtime.ErrorRate == nil {
			//平均错误率为0 ：突变率为0
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = 0
			}
			//平均错误率不为0，突变率为-1
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = -1
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].REDMetrics.Avg.ErrorRate == nil && endpoints[i].REDMetrics.Realtime.ErrorRate != nil {
			//1m错误率为0，突变率为0
			if endpoints[i].REDMetrics.Realtime.ErrorRate != nil && *endpoints[i].REDMetrics.Realtime.ErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = 0
			}
			//1m错误率不为0，突变率为max
			if endpoints[i].REDMetrics.Realtime.ErrorRate != nil && *endpoints[i].REDMetrics.Realtime.ErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[i].REDMetrics.Realtime.ErrorRate != nil {
			//1m错误率为0，突变率为0
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
			//1m错误率不为0，突变率为max
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[i].REDMetrics.Realtime.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = *endpoints[i].REDMetrics.Realtime.ErrorRate / *endpoints[i].REDMetrics.Avg.ErrorRate
			}
		}
		//latency
		//平均延时和1m延时都查不出来，突变率为0(不可能的情况)
		if endpoints[i].REDMetrics.Avg.Latency == nil && endpoints[i].REDMetrics.Realtime.Latency == nil {
			endpoints[i].Avg1MinLatencyMutationRate = 0
		}
		//平均延时查的出来，1m延时查不出来
		if endpoints[i].REDMetrics.Avg.Latency != nil && endpoints[i].REDMetrics.Realtime.Latency == nil {
			//平均延时为0 ：突变率为0
			if endpoints[i].REDMetrics.Avg.Latency != nil && *endpoints[i].REDMetrics.Avg.Latency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = 0
			}
			//平均延时不为0，突变率为-1
			if endpoints[i].REDMetrics.Avg.Latency != nil && *endpoints[i].REDMetrics.Avg.Latency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = -1
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].REDMetrics.Avg.Latency == nil && endpoints[i].REDMetrics.Realtime.Latency != nil {
			//1m延时为0，突变率为0
			if endpoints[i].REDMetrics.Realtime.Latency != nil && *endpoints[i].REDMetrics.Realtime.Latency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = 0
			}
			//1m延时不为0，突变率为max
			if endpoints[i].REDMetrics.Realtime.Latency != nil && *endpoints[i].REDMetrics.Realtime.Latency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].REDMetrics.Avg.Latency != nil && endpoints[i].REDMetrics.Realtime.Latency != nil {
			//平均延时为0，突变率为max
			if endpoints[i].REDMetrics.Avg.Latency != nil && *endpoints[i].REDMetrics.Avg.Latency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
			//平均延时不为0，突变率为1m延时/平均延时
			if endpoints[i].REDMetrics.Avg.Latency != nil && endpoints[i].REDMetrics.Realtime.Latency != nil && *endpoints[i].REDMetrics.Avg.Latency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = *endpoints[i].REDMetrics.Realtime.Latency / *endpoints[i].REDMetrics.Avg.Latency
			}
		}
		//if urls[i].REDMetrics.Realtime.ErrorRate != nil {
		//	a := *urls[i].REDMetrics.Realtime.ErrorRate
		//	fmt.Printf("1minError:%v,iminLatency:%v\n", a, urls[i].Avg1minLatency)
		//} else {
		//	a := 10
		//	fmt.Printf("1minError:%v,iminLatency:%v\n", a, urls[i].Avg1minLatency)
		//}
		//if urls[i].REDMetrics.Avg.ErrorRate != nil {
		//	a := *urls[i].REDMetrics.Avg.ErrorRate
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
			if endpoints[i].REDMetrics.Realtime.ErrorRate != nil && endpoints[j].REDMetrics.Realtime.ErrorRate != nil {
				return *endpoints[i].REDMetrics.Realtime.ErrorRate > *endpoints[j].REDMetrics.Realtime.ErrorRate
			}
			// 如果一个错误率为nil，另一个不为nil，错误率不为nil的排在前面
			if endpoints[i].REDMetrics.Realtime.ErrorRate != nil && endpoints[j].REDMetrics.Realtime.ErrorRate == nil {
				return true
			}
			if endpoints[i].REDMetrics.Realtime.ErrorRate == nil && endpoints[j].REDMetrics.Realtime.ErrorRate != nil {
				return false
			}
			// 如果延迟突变率相同，按错误率排序
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[j].REDMetrics.Avg.ErrorRate != nil {
				return *endpoints[i].REDMetrics.Avg.ErrorRate > *endpoints[j].REDMetrics.Avg.ErrorRate
			}
			// 如果一个错误率为nil，另一个不为nil，错误率不为nil的排在前面
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[j].REDMetrics.Avg.ErrorRate == nil {
				return true
			}
			if endpoints[i].REDMetrics.Avg.ErrorRate == nil && endpoints[j].REDMetrics.Avg.ErrorRate != nil {
				return false
			}
			// 如果错误率相同或都为nil，按延迟排序
			if endpoints[i].REDMetrics.Realtime.Latency != nil && endpoints[j].REDMetrics.Realtime.Latency != nil {
				return *endpoints[i].REDMetrics.Realtime.Latency > *endpoints[j].REDMetrics.Realtime.Latency
			}
			if endpoints[i].REDMetrics.Realtime.Latency != nil && endpoints[j].REDMetrics.Realtime.Latency == nil {
				return true
			}
			if endpoints[i].REDMetrics.Realtime.Latency == nil && endpoints[j].REDMetrics.Realtime.Latency != nil {
				return false
			}
			// 如果错误率相同或都为nil，按延迟排序
			if endpoints[i].REDMetrics.Avg.Latency != nil && endpoints[j].REDMetrics.Avg.Latency != nil {
				return *endpoints[i].REDMetrics.Avg.Latency > *endpoints[j].REDMetrics.Avg.Latency
			}
			if endpoints[i].REDMetrics.Avg.Latency != nil && endpoints[j].REDMetrics.Avg.Latency == nil {
				return true
			}
			if endpoints[i].REDMetrics.Avg.Latency == nil && endpoints[j].REDMetrics.Avg.Latency != nil {
				return false
			}
		}
		return false
	})

}

func fillServices(endpoints []*prom.EndpointMetrics) []ServiceDetail {
	var services []ServiceDetail
	for _, url := range endpoints {
		//如果没有数据则不返回
		if (url.REDMetrics.Avg.Latency == nil && url.REDMetrics.Avg.ErrorRate == nil) || (url.REDMetrics.Avg.Latency == nil && url.REDMetrics.Avg.ErrorRate != nil && *url.REDMetrics.Avg.ErrorRate == 0 && url.REDMetrics.Avg.TPM == nil) {
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
				Endpoints:     []*prom.EndpointMetrics{url},
			}
			services = append(services, newService)
		}
	}

	return services
}

func fillOneService(endpoints []*prom.EndpointMetrics) []ServiceDetail {
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
				Endpoints:     []*prom.EndpointMetrics{url},
			}
			service = append(service, newService)
		}
	}

	return service
}
