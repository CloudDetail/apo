package serviceoverview

import (
	"sort"
)

// 跟据日同比阈值进行排序,错误率同比相同，比较错误率值
func sortByDODThreshold(endpoints []*Endpoint) {
	sort.SliceStable(endpoints, func(i, j int) bool {
		//先按照count排序
		if endpoints[i].Count != endpoints[j].Count {
			return endpoints[i].Count > endpoints[j].Count
		}
		//等于3时按照错误率排序
		if endpoints[i].Count == endpoints[j].Count && endpoints[i].Count == 3 {
			if endpoints[i].ErrorRateDayOverDay != nil && endpoints[j].ErrorRateDayOverDay != nil && endpoints[i].ErrorRateDayOverDay != endpoints[j].ErrorRateDayOverDay {
				if *endpoints[i].ErrorRateDayOverDay != *endpoints[j].ErrorRateDayOverDay {
					return *endpoints[i].ErrorRateDayOverDay > *endpoints[j].ErrorRateDayOverDay
				}
				if *endpoints[i].ErrorRateDayOverDay == *endpoints[j].ErrorRateDayOverDay && endpoints[i].AvgErrorRate != nil && endpoints[j].AvgErrorRate != nil {
					return *endpoints[i].AvgErrorRate > *endpoints[j].AvgErrorRate
				}
			}
			if endpoints[i].LatencyDayOverDay != nil && endpoints[j].LatencyDayOverDay != nil && endpoints[i].LatencyDayOverDay != endpoints[j].LatencyDayOverDay {
				return *endpoints[i].LatencyDayOverDay > *endpoints[j].LatencyDayOverDay
			}
			if endpoints[i].TPMDayOverDay != nil && endpoints[j].TPMDayOverDay != nil {
				return *endpoints[i].TPMDayOverDay > *endpoints[j].TPMDayOverDay
			}
		}
		//count = 2的比较方式
		if endpoints[i].Count == endpoints[j].Count && endpoints[i].Count == 2 {
			if endpoints[i].IsErrorRateExceeded == true && endpoints[j].IsErrorRateExceeded == false {
				return true
			}
			if endpoints[i].IsLatencyExceeded == true && endpoints[j].IsLatencyExceeded == false && endpoints[i].IsErrorRateExceeded == endpoints[j].IsErrorRateExceeded {
				return true
			}
			if endpoints[i].IsErrorRateExceeded == endpoints[j].IsErrorRateExceeded && endpoints[j].IsErrorRateExceeded == false {
				if endpoints[i].LatencyDayOverDay != nil && endpoints[j].LatencyDayOverDay != nil && endpoints[i].LatencyDayOverDay != endpoints[j].LatencyDayOverDay {
					return *endpoints[i].LatencyDayOverDay > *endpoints[j].LatencyDayOverDay
				}
				if endpoints[i].TPMDayOverDay != nil && endpoints[j].TPMDayOverDay != nil {
					return *endpoints[i].TPMDayOverDay > *endpoints[j].TPMDayOverDay
				}
			}

			if endpoints[i].IsLatencyExceeded == endpoints[j].IsLatencyExceeded && endpoints[j].IsLatencyExceeded == false {
				if endpoints[i].ErrorRateDayOverDay != nil && endpoints[j].ErrorRateDayOverDay != nil && endpoints[i].ErrorRateDayOverDay != endpoints[j].ErrorRateDayOverDay {
					if *endpoints[i].ErrorRateDayOverDay != *endpoints[j].ErrorRateDayOverDay {
						return *endpoints[i].ErrorRateDayOverDay > *endpoints[j].ErrorRateDayOverDay
					}
					if *endpoints[i].ErrorRateDayOverDay == *endpoints[j].ErrorRateDayOverDay && endpoints[i].AvgErrorRate != nil && endpoints[j].AvgErrorRate != nil {
						return *endpoints[i].AvgErrorRate > *endpoints[j].AvgErrorRate
					}
				}
				if endpoints[i].TPMDayOverDay != nil && endpoints[j].TPMDayOverDay != nil {
					return *endpoints[i].TPMDayOverDay > *endpoints[j].TPMDayOverDay
				}
			}
			if endpoints[i].IsTPSExceeded == endpoints[j].IsTPSExceeded && endpoints[j].IsTPSExceeded == false {
				if endpoints[i].ErrorRateDayOverDay != nil && endpoints[j].ErrorRateDayOverDay != nil && endpoints[i].ErrorRateDayOverDay != endpoints[j].ErrorRateDayOverDay {
					if *endpoints[i].ErrorRateDayOverDay != *endpoints[j].ErrorRateDayOverDay {
						return *endpoints[i].ErrorRateDayOverDay > *endpoints[j].ErrorRateDayOverDay
					}
					if *endpoints[i].ErrorRateDayOverDay == *endpoints[j].ErrorRateDayOverDay && endpoints[i].AvgErrorRate != nil && endpoints[j].AvgErrorRate != nil {
						return *endpoints[i].AvgErrorRate > *endpoints[j].AvgErrorRate
					}
				}
				if endpoints[i].LatencyDayOverDay != nil && endpoints[j].LatencyDayOverDay != nil && endpoints[i].LatencyDayOverDay != endpoints[j].LatencyDayOverDay {
					return *endpoints[i].LatencyDayOverDay > *endpoints[j].LatencyDayOverDay
				}

				if endpoints[i].TPMDayOverDay != nil && endpoints[j].TPMDayOverDay != nil {
					return *endpoints[i].TPMDayOverDay > *endpoints[j].TPMDayOverDay
				}
			}

		}
		if endpoints[i].Count == endpoints[j].Count && endpoints[i].Count == 1 {
			if endpoints[i].IsErrorRateExceeded == true && endpoints[j].IsErrorRateExceeded == false {
				return true
			}
			if endpoints[i].IsLatencyExceeded == true && endpoints[j].IsLatencyExceeded == false && endpoints[i].IsErrorRateExceeded == endpoints[j].IsErrorRateExceeded {
				return true
			}
			if endpoints[i].ErrorRateDayOverDay != nil && endpoints[j].ErrorRateDayOverDay != nil && endpoints[i].IsErrorRateExceeded == endpoints[j].IsErrorRateExceeded && endpoints[j].IsErrorRateExceeded == true {
				if *endpoints[i].ErrorRateDayOverDay != *endpoints[j].ErrorRateDayOverDay {
					return *endpoints[i].ErrorRateDayOverDay > *endpoints[j].ErrorRateDayOverDay
				}
				if *endpoints[i].ErrorRateDayOverDay == *endpoints[j].ErrorRateDayOverDay && endpoints[i].AvgErrorRate != nil && endpoints[j].AvgErrorRate != nil {
					return *endpoints[i].AvgErrorRate > *endpoints[j].AvgErrorRate
				}

			}
			if endpoints[i].LatencyDayOverDay != nil && endpoints[j].LatencyDayOverDay != nil && endpoints[i].IsLatencyExceeded == endpoints[j].IsLatencyExceeded && endpoints[j].IsLatencyExceeded == true {
				return *endpoints[i].LatencyDayOverDay > *endpoints[j].LatencyDayOverDay
			}
			if endpoints[i].TPMDayOverDay != nil && endpoints[j].TPMDayOverDay != nil && endpoints[i].IsTPSExceeded == endpoints[j].IsTPSExceeded && endpoints[j].IsTPSExceeded == true {
				return *endpoints[i].TPMDayOverDay > *endpoints[j].TPMDayOverDay
			}
		}
		if endpoints[i].Count == endpoints[j].Count && endpoints[i].Count == 0 {
			if endpoints[i].ErrorRateDayOverDay != nil && endpoints[j].ErrorRateDayOverDay == nil {
				return true
			}
			if endpoints[i].ErrorRateDayOverDay == endpoints[j].ErrorRateDayOverDay && endpoints[i].LatencyDayOverDay != nil && endpoints[j].LatencyDayOverDay == nil {
				return true
			}
			if endpoints[i].ErrorRateDayOverDay == endpoints[j].ErrorRateDayOverDay && endpoints[i].LatencyDayOverDay == endpoints[j].LatencyDayOverDay && endpoints[i].TPMDayOverDay != nil && endpoints[j].TPMDayOverDay == nil {
				return true
			}
			if endpoints[i].ErrorRateDayOverDay != nil && endpoints[j].ErrorRateDayOverDay != nil && endpoints[i].ErrorRateDayOverDay != endpoints[j].ErrorRateDayOverDay {
				if *endpoints[i].ErrorRateDayOverDay != *endpoints[j].ErrorRateDayOverDay {
					return *endpoints[i].ErrorRateDayOverDay > *endpoints[j].ErrorRateDayOverDay
				}
				if *endpoints[i].ErrorRateDayOverDay == *endpoints[j].ErrorRateDayOverDay && endpoints[i].AvgErrorRate != nil && endpoints[j].AvgErrorRate != nil {
					return *endpoints[i].AvgErrorRate > *endpoints[j].AvgErrorRate
				}
			}
			if endpoints[i].LatencyDayOverDay != nil && endpoints[j].LatencyDayOverDay != nil && endpoints[i].LatencyDayOverDay != endpoints[j].LatencyDayOverDay {
				return *endpoints[i].LatencyDayOverDay > *endpoints[j].LatencyDayOverDay
			}
			if endpoints[i].TPMDayOverDay != nil && endpoints[j].TPMDayOverDay != nil && endpoints[i].TPMDayOverDay != endpoints[j].TPMDayOverDay {
				return *endpoints[i].TPMDayOverDay > *endpoints[j].TPMDayOverDay
			}

		}

		return endpoints[i].Count > endpoints[j].Count
	})
}

// 突变排序
func sortByMutation(endpoints []*Endpoint) {
	for i, _ := range endpoints {
		//平均错误率和1m错误率都查不出来，突变率为0
		if endpoints[i].AvgErrorRate == nil && endpoints[i].Avg1minErrorRate == nil {
			endpoints[i].Avg1MinErrorMutationRate = 0
		}
		//平均错误率查的出来，1m错误率查不出来
		if endpoints[i].AvgErrorRate != nil && endpoints[i].Avg1minErrorRate == nil {
			//平均错误率为0 ：突变率为0
			if endpoints[i].AvgErrorRate != nil && *endpoints[i].AvgErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = 0
			}
			//平均错误率不为0，突变率为-1
			if endpoints[i].AvgErrorRate != nil && *endpoints[i].AvgErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = -1
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].AvgErrorRate == nil && endpoints[i].Avg1minErrorRate != nil {
			//1m错误率为0，突变率为0
			if endpoints[i].Avg1minErrorRate != nil && *endpoints[i].Avg1minErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = 0
			}
			//1m错误率不为0，突变率为max
			if endpoints[i].Avg1minErrorRate != nil && *endpoints[i].Avg1minErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].AvgErrorRate != nil && endpoints[i].Avg1minErrorRate != nil {
			//1m错误率为0，突变率为0
			if endpoints[i].AvgErrorRate != nil && *endpoints[i].AvgErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
			//1m错误率不为0，突变率为max
			if endpoints[i].AvgErrorRate != nil && endpoints[i].Avg1minErrorRate != nil && *endpoints[i].AvgErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = *endpoints[i].Avg1minErrorRate / *endpoints[i].AvgErrorRate
			}
		}
		//latency
		//平均延时和1m延时都查不出来，突变率为0(不可能的情况)
		if endpoints[i].AvgLatency == nil && endpoints[i].Avg1minLatency == nil {
			endpoints[i].Avg1MinLatencyMutationRate = 0
		}
		//平均延时查的出来，1m延时查不出来
		if endpoints[i].AvgLatency != nil && endpoints[i].Avg1minLatency == nil {
			//平均延时为0 ：突变率为0
			if endpoints[i].AvgLatency != nil && *endpoints[i].AvgLatency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = 0
			}
			//平均延时不为0，突变率为-1
			if endpoints[i].AvgLatency != nil && *endpoints[i].AvgLatency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = -1
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].AvgLatency == nil && endpoints[i].Avg1minLatency != nil {
			//1m延时为0，突变率为0
			if endpoints[i].Avg1minLatency != nil && *endpoints[i].Avg1minLatency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = 0
			}
			//1m延时不为0，突变率为max
			if endpoints[i].Avg1minLatency != nil && *endpoints[i].Avg1minLatency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if endpoints[i].AvgLatency != nil && endpoints[i].Avg1minLatency != nil {
			//平均延时为0，突变率为max
			if endpoints[i].AvgLatency != nil && *endpoints[i].AvgLatency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
			//平均延时不为0，突变率为1m延时/平均延时
			if endpoints[i].AvgLatency != nil && endpoints[i].Avg1minLatency != nil && *endpoints[i].AvgLatency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = *endpoints[i].Avg1minLatency / *endpoints[i].AvgLatency
			}
		}
		//if urls[i].Avg1minErrorRate != nil {
		//	a := *urls[i].Avg1minErrorRate
		//	fmt.Printf("1minError:%v,iminLatency:%v\n", a, urls[i].Avg1minLatency)
		//} else {
		//	a := 10
		//	fmt.Printf("1minError:%v,iminLatency:%v\n", a, urls[i].Avg1minLatency)
		//}
		//if urls[i].AvgErrorRate != nil {
		//	a := *urls[i].AvgErrorRate
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
			if endpoints[i].Avg1minErrorRate != nil && endpoints[j].Avg1minErrorRate != nil {
				return *endpoints[i].Avg1minErrorRate > *endpoints[j].Avg1minErrorRate
			}
			// 如果一个错误率为nil，另一个不为nil，错误率不为nil的排在前面
			if endpoints[i].Avg1minErrorRate != nil && endpoints[j].Avg1minErrorRate == nil {
				return true
			}
			if endpoints[i].Avg1minErrorRate == nil && endpoints[j].Avg1minErrorRate != nil {
				return false
			}
			// 如果延迟突变率相同，按错误率排序
			if endpoints[i].AvgErrorRate != nil && endpoints[j].AvgErrorRate != nil {
				return *endpoints[i].AvgErrorRate > *endpoints[j].AvgErrorRate
			}
			// 如果一个错误率为nil，另一个不为nil，错误率不为nil的排在前面
			if endpoints[i].AvgErrorRate != nil && endpoints[j].AvgErrorRate == nil {
				return true
			}
			if endpoints[i].AvgErrorRate == nil && endpoints[j].AvgErrorRate != nil {
				return false
			}
			// 如果错误率相同或都为nil，按延迟排序
			if endpoints[i].Avg1minLatency != nil && endpoints[j].Avg1minLatency != nil {
				return *endpoints[i].Avg1minLatency > *endpoints[j].Avg1minLatency
			}
			if endpoints[i].Avg1minLatency != nil && endpoints[j].Avg1minLatency == nil {
				return true
			}
			if endpoints[i].Avg1minLatency == nil && endpoints[j].Avg1minLatency != nil {
				return false
			}
			// 如果错误率相同或都为nil，按延迟排序
			if endpoints[i].AvgLatency != nil && endpoints[j].AvgLatency != nil {
				return *endpoints[i].AvgLatency > *endpoints[j].AvgLatency
			}
			if endpoints[i].AvgLatency != nil && endpoints[j].AvgLatency == nil {
				return true
			}
			if endpoints[i].AvgLatency == nil && endpoints[j].AvgLatency != nil {
				return false
			}
		}
		return false
	})

}

func fillServices(endpoints []*Endpoint) []serviceDetail {
	var Services []serviceDetail
	for _, url := range endpoints {
		//如果没有数据则不返回
		if (url.AvgLatency == nil && url.AvgErrorRate == nil) || (url.AvgLatency == nil && url.AvgErrorRate != nil && *url.AvgErrorRate == 0 && url.AvgTPM == nil) {
			continue
		}
		serviceName := url.SvcName
		found := false
		for j, _ := range Services {
			if Services[j].ServiceName == serviceName {
				found = true
				Services[j].EndpointCount++
				if Services[j].ServiceSize < 3 {
					Services[j].Endpoints = append(Services[j].Endpoints, url)
					Services[j].ServiceSize++
					break
				}
			}
		}
		if !found {
			newService := serviceDetail{
				ServiceName:   serviceName,
				ServiceSize:   1,
				EndpointCount: 1,
				Endpoints:     []*Endpoint{url},
			}
			Services = append(Services, newService)
		}
	}

	return Services
}

func fillOneService(endpoints []*Endpoint) []serviceDetail {
	var Services []serviceDetail
	for _, url := range endpoints {
		serviceName := url.SvcName
		found := false
		for j, _ := range Services {
			if Services[j].ServiceName == serviceName {
				found = true
				Services[j].EndpointCount++
				Services[j].Endpoints = append(Services[j].Endpoints, url)
				Services[j].ServiceSize++
				break
			}
		}
		if !found {
			newService := serviceDetail{
				ServiceName:   serviceName,
				ServiceSize:   1,
				EndpointCount: 1,
				Endpoints:     []*Endpoint{url},
			}
			Services = append(Services, newService)
		}
	}

	return Services
}
