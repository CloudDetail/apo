package serviceoverview

import (
	"sort"
)

// 跟据日同比阈值进行排序,错误率同比相同，比较错误率值
func sortByDODThreshold(urls []Url) {
	sort.SliceStable(urls, func(i, j int) bool {
		//先按照count排序
		if urls[i].Count != urls[j].Count {
			return urls[i].Count > urls[j].Count
		}
		//等于3时按照错误率排序
		if urls[i].Count == urls[j].Count && urls[i].Count == 3 {
			if urls[i].ErrorRateDayOverDay != nil && urls[j].ErrorRateDayOverDay != nil && urls[i].ErrorRateDayOverDay != urls[j].ErrorRateDayOverDay {
				if *urls[i].ErrorRateDayOverDay != *urls[j].ErrorRateDayOverDay {
					return *urls[i].ErrorRateDayOverDay > *urls[j].ErrorRateDayOverDay
				}
				if *urls[i].ErrorRateDayOverDay == *urls[j].ErrorRateDayOverDay && urls[i].AvgErrorRate != nil && urls[j].AvgErrorRate != nil {
					return *urls[i].AvgErrorRate > *urls[j].AvgErrorRate
				}
			}
			if urls[i].LatencyDayOverDay != nil && urls[j].LatencyDayOverDay != nil && urls[i].LatencyDayOverDay != urls[j].LatencyDayOverDay {
				return *urls[i].LatencyDayOverDay > *urls[j].LatencyDayOverDay
			}
			if urls[i].TPSDayOverDay != nil && urls[j].TPSDayOverDay != nil {
				return *urls[i].TPSDayOverDay > *urls[j].TPSDayOverDay
			}
		}
		//count = 2的比较方式
		if urls[i].Count == urls[j].Count && urls[i].Count == 2 {
			if urls[i].IsErrorRateExceeded == true && urls[j].IsErrorRateExceeded == false {
				return true
			}
			if urls[i].IsLatencyExceeded == true && urls[j].IsLatencyExceeded == false && urls[i].IsErrorRateExceeded == urls[j].IsErrorRateExceeded {
				return true
			}
			if urls[i].IsErrorRateExceeded == urls[j].IsErrorRateExceeded && urls[j].IsErrorRateExceeded == false {
				if urls[i].LatencyDayOverDay != nil && urls[j].LatencyDayOverDay != nil && urls[i].LatencyDayOverDay != urls[j].LatencyDayOverDay {
					return *urls[i].LatencyDayOverDay > *urls[j].LatencyDayOverDay
				}
				if urls[i].TPSDayOverDay != nil && urls[j].TPSDayOverDay != nil {
					return *urls[i].TPSDayOverDay > *urls[j].TPSDayOverDay
				}
			}

			if urls[i].IsLatencyExceeded == urls[j].IsLatencyExceeded && urls[j].IsLatencyExceeded == false {
				if urls[i].ErrorRateDayOverDay != nil && urls[j].ErrorRateDayOverDay != nil && urls[i].ErrorRateDayOverDay != urls[j].ErrorRateDayOverDay {
					if *urls[i].ErrorRateDayOverDay != *urls[j].ErrorRateDayOverDay {
						return *urls[i].ErrorRateDayOverDay > *urls[j].ErrorRateDayOverDay
					}
					if *urls[i].ErrorRateDayOverDay == *urls[j].ErrorRateDayOverDay && urls[i].AvgErrorRate != nil && urls[j].AvgErrorRate != nil {
						return *urls[i].AvgErrorRate > *urls[j].AvgErrorRate
					}
				}
				if urls[i].TPSDayOverDay != nil && urls[j].TPSDayOverDay != nil {
					return *urls[i].TPSDayOverDay > *urls[j].TPSDayOverDay
				}
			}
			if urls[i].IsTPSExceeded == urls[j].IsTPSExceeded && urls[j].IsTPSExceeded == false {
				if urls[i].ErrorRateDayOverDay != nil && urls[j].ErrorRateDayOverDay != nil && urls[i].ErrorRateDayOverDay != urls[j].ErrorRateDayOverDay {
					if *urls[i].ErrorRateDayOverDay != *urls[j].ErrorRateDayOverDay {
						return *urls[i].ErrorRateDayOverDay > *urls[j].ErrorRateDayOverDay
					}
					if *urls[i].ErrorRateDayOverDay == *urls[j].ErrorRateDayOverDay && urls[i].AvgErrorRate != nil && urls[j].AvgErrorRate != nil {
						return *urls[i].AvgErrorRate > *urls[j].AvgErrorRate
					}
				}
				if urls[i].LatencyDayOverDay != nil && urls[j].LatencyDayOverDay != nil && urls[i].LatencyDayOverDay != urls[j].LatencyDayOverDay {
					return *urls[i].LatencyDayOverDay > *urls[j].LatencyDayOverDay
				}

				if urls[i].TPSDayOverDay != nil && urls[j].TPSDayOverDay != nil {
					return *urls[i].TPSDayOverDay > *urls[j].TPSDayOverDay
				}
			}

		}
		if urls[i].Count == urls[j].Count && urls[i].Count == 1 {
			if urls[i].IsErrorRateExceeded == true && urls[j].IsErrorRateExceeded == false {
				return true
			}
			if urls[i].IsLatencyExceeded == true && urls[j].IsLatencyExceeded == false && urls[i].IsErrorRateExceeded == urls[j].IsErrorRateExceeded {
				return true
			}
			if urls[i].ErrorRateDayOverDay != nil && urls[j].ErrorRateDayOverDay != nil && urls[i].IsErrorRateExceeded == urls[j].IsErrorRateExceeded && urls[j].IsErrorRateExceeded == true {
				if *urls[i].ErrorRateDayOverDay != *urls[j].ErrorRateDayOverDay {
					return *urls[i].ErrorRateDayOverDay > *urls[j].ErrorRateDayOverDay
				}
				if *urls[i].ErrorRateDayOverDay == *urls[j].ErrorRateDayOverDay && urls[i].AvgErrorRate != nil && urls[j].AvgErrorRate != nil {
					return *urls[i].AvgErrorRate > *urls[j].AvgErrorRate
				}

			}
			if urls[i].LatencyDayOverDay != nil && urls[j].LatencyDayOverDay != nil && urls[i].IsLatencyExceeded == urls[j].IsLatencyExceeded && urls[j].IsLatencyExceeded == true {
				return *urls[i].LatencyDayOverDay > *urls[j].LatencyDayOverDay
			}
			if urls[i].TPSDayOverDay != nil && urls[j].TPSDayOverDay != nil && urls[i].IsTPSExceeded == urls[j].IsTPSExceeded && urls[j].IsTPSExceeded == true {
				return *urls[i].TPSDayOverDay > *urls[j].TPSDayOverDay
			}
		}
		if urls[i].Count == urls[j].Count && urls[i].Count == 0 {
			if urls[i].ErrorRateDayOverDay != nil && urls[j].ErrorRateDayOverDay == nil {
				return true
			}
			if urls[i].ErrorRateDayOverDay == urls[j].ErrorRateDayOverDay && urls[i].LatencyDayOverDay != nil && urls[j].LatencyDayOverDay == nil {
				return true
			}
			if urls[i].ErrorRateDayOverDay == urls[j].ErrorRateDayOverDay && urls[i].LatencyDayOverDay == urls[j].LatencyDayOverDay && urls[i].TPSDayOverDay != nil && urls[j].TPSDayOverDay == nil {
				return true
			}
			if urls[i].ErrorRateDayOverDay != nil && urls[j].ErrorRateDayOverDay != nil && urls[i].ErrorRateDayOverDay != urls[j].ErrorRateDayOverDay {
				if *urls[i].ErrorRateDayOverDay != *urls[j].ErrorRateDayOverDay {
					return *urls[i].ErrorRateDayOverDay > *urls[j].ErrorRateDayOverDay
				}
				if *urls[i].ErrorRateDayOverDay == *urls[j].ErrorRateDayOverDay && urls[i].AvgErrorRate != nil && urls[j].AvgErrorRate != nil {
					return *urls[i].AvgErrorRate > *urls[j].AvgErrorRate
				}
			}
			if urls[i].LatencyDayOverDay != nil && urls[j].LatencyDayOverDay != nil && urls[i].LatencyDayOverDay != urls[j].LatencyDayOverDay {
				return *urls[i].LatencyDayOverDay > *urls[j].LatencyDayOverDay
			}
			if urls[i].TPSDayOverDay != nil && urls[j].TPSDayOverDay != nil && urls[i].TPSDayOverDay != urls[j].TPSDayOverDay {
				return *urls[i].TPSDayOverDay > *urls[j].TPSDayOverDay
			}

		}

		return urls[i].Count > urls[j].Count
	})
}

// 突变排序
func sortByMutation(urls []Url) {
	for i, _ := range urls {
		//平均错误率和1m错误率都查不出来，突变率为0
		if urls[i].AvgErrorRate == nil && urls[i].Avg1minErrorRate == nil {
			urls[i].Avg1MinErrorMutationRate = 0
		}
		//平均错误率查的出来，1m错误率查不出来
		if urls[i].AvgErrorRate != nil && urls[i].Avg1minErrorRate == nil {
			//平均错误率为0 ：突变率为0
			if urls[i].AvgErrorRate != nil && *urls[i].AvgErrorRate == 0 {
				urls[i].Avg1MinErrorMutationRate = 0
			}
			//平均错误率不为0，突变率为-1
			if urls[i].AvgErrorRate != nil && *urls[i].AvgErrorRate != 0 {
				urls[i].Avg1MinErrorMutationRate = -1
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if urls[i].AvgErrorRate == nil && urls[i].Avg1minErrorRate != nil {
			//1m错误率为0，突变率为0
			if urls[i].Avg1minErrorRate != nil && *urls[i].Avg1minErrorRate == 0 {
				urls[i].Avg1MinErrorMutationRate = 0
			}
			//1m错误率不为0，突变率为max
			if urls[i].Avg1minErrorRate != nil && *urls[i].Avg1minErrorRate != 0 {
				urls[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if urls[i].AvgErrorRate != nil && urls[i].Avg1minErrorRate != nil {
			//1m错误率为0，突变率为0
			if urls[i].AvgErrorRate != nil && *urls[i].AvgErrorRate == 0 {
				urls[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
			//1m错误率不为0，突变率为max
			if urls[i].AvgErrorRate != nil && urls[i].Avg1minErrorRate != nil && *urls[i].AvgErrorRate != 0 {
				urls[i].Avg1MinErrorMutationRate = *urls[i].Avg1minErrorRate / *urls[i].AvgErrorRate
			}
		}
		//latency
		//平均延时和1m延时都查不出来，突变率为0(不可能的情况)
		if urls[i].AvgLatency == nil && urls[i].Avg1minLatency == nil {
			urls[i].Avg1MinLatencyMutationRate = 0
		}
		//平均延时查的出来，1m延时查不出来
		if urls[i].AvgLatency != nil && urls[i].Avg1minLatency == nil {
			//平均延时为0 ：突变率为0
			if urls[i].AvgLatency != nil && *urls[i].AvgLatency == 0 {
				urls[i].Avg1MinLatencyMutationRate = 0
			}
			//平均延时不为0，突变率为-1
			if urls[i].AvgLatency != nil && *urls[i].AvgLatency != 0 {
				urls[i].Avg1MinLatencyMutationRate = -1
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if urls[i].AvgLatency == nil && urls[i].Avg1minLatency != nil {
			//1m延时为0，突变率为0
			if urls[i].Avg1minLatency != nil && *urls[i].Avg1minLatency == 0 {
				urls[i].Avg1MinLatencyMutationRate = 0
			}
			//1m延时不为0，突变率为max
			if urls[i].Avg1minLatency != nil && *urls[i].Avg1minLatency != 0 {
				urls[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
		}
		//平均错误率查不出来，1m错误率查的出来
		if urls[i].AvgLatency != nil && urls[i].Avg1minLatency != nil {
			//平均延时为0，突变率为max
			if urls[i].AvgLatency != nil && *urls[i].AvgLatency == 0 {
				urls[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
			//平均延时不为0，突变率为1m延时/平均延时
			if urls[i].AvgLatency != nil && urls[i].Avg1minLatency != nil && *urls[i].AvgLatency != 0 {
				urls[i].Avg1MinLatencyMutationRate = *urls[i].Avg1minLatency / *urls[i].AvgLatency
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
	sort.SliceStable(urls, func(i, j int) bool {
		// Case 1: 如果有一个错误率突变率大于1（错误率上升）
		if urls[i].Avg1MinErrorMutationRate > 1 || urls[j].Avg1MinErrorMutationRate > 1 {
			return urls[i].Avg1MinErrorMutationRate > urls[j].Avg1MinErrorMutationRate
		}

		// Case 2: 如果错误率突变率都小于等于1
		if urls[i].Avg1MinErrorMutationRate <= 1 && urls[j].Avg1MinErrorMutationRate <= 1 {
			// 优先按延迟突变率排序，较大的排在前面
			if urls[i].Avg1MinLatencyMutationRate != urls[j].Avg1MinLatencyMutationRate {
				return urls[i].Avg1MinLatencyMutationRate > urls[j].Avg1MinLatencyMutationRate
			}

			// 如果延迟突变率相同，按错误率排序
			if urls[i].Avg1minErrorRate != nil && urls[j].Avg1minErrorRate != nil {
				return *urls[i].Avg1minErrorRate > *urls[j].Avg1minErrorRate
			}
			// 如果一个错误率为nil，另一个不为nil，错误率不为nil的排在前面
			if urls[i].Avg1minErrorRate != nil && urls[j].Avg1minErrorRate == nil {
				return true
			}
			if urls[i].Avg1minErrorRate == nil && urls[j].Avg1minErrorRate != nil {
				return false
			}
			// 如果延迟突变率相同，按错误率排序
			if urls[i].AvgErrorRate != nil && urls[j].AvgErrorRate != nil {
				return *urls[i].AvgErrorRate > *urls[j].AvgErrorRate
			}
			// 如果一个错误率为nil，另一个不为nil，错误率不为nil的排在前面
			if urls[i].AvgErrorRate != nil && urls[j].AvgErrorRate == nil {
				return true
			}
			if urls[i].AvgErrorRate == nil && urls[j].AvgErrorRate != nil {
				return false
			}
			// 如果错误率相同或都为nil，按延迟排序
			if urls[i].Avg1minLatency != nil && urls[j].Avg1minLatency != nil {
				return *urls[i].Avg1minLatency > *urls[j].Avg1minLatency
			}
			if urls[i].Avg1minLatency != nil && urls[j].Avg1minLatency == nil {
				return true
			}
			if urls[i].Avg1minLatency == nil && urls[j].Avg1minLatency != nil {
				return false
			}
			// 如果错误率相同或都为nil，按延迟排序
			if urls[i].AvgLatency != nil && urls[j].AvgLatency != nil {
				return *urls[i].AvgLatency > *urls[j].AvgLatency
			}
			if urls[i].AvgLatency != nil && urls[j].AvgLatency == nil {
				return true
			}
			if urls[i].AvgLatency == nil && urls[j].AvgLatency != nil {
				return false
			}
		}
		return false
	})

}

func fillServices(Urls []Url) []serviceDetail {
	var Services []serviceDetail
	for _, url := range Urls {
		//如果没有数据则不返回
		if (url.AvgLatency == nil && url.AvgErrorRate == nil) || (url.AvgLatency == nil && url.AvgErrorRate != nil && *url.AvgErrorRate == 0 && url.AvgTPS == nil) {
			continue
		}
		serviceName := url.SvcName
		found := false
		for j, _ := range Services {
			if Services[j].ServiceName == serviceName {
				found = true
				Services[j].EndpointCount++
				if Services[j].ServiceSize < 3 {
					Services[j].Urls = append(Services[j].Urls, url)
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
				Urls:          []Url{url},
			}
			Services = append(Services, newService)
		}
	}

	return Services
}

func fillOneService(Urls []Url) []serviceDetail {
	var Services []serviceDetail
	for _, url := range Urls {
		serviceName := url.SvcName
		found := false
		for j, _ := range Services {
			if Services[j].ServiceName == serviceName {
				found = true
				Services[j].EndpointCount++
				Services[j].Urls = append(Services[j].Urls, url)
				Services[j].ServiceSize++
				break
			}
		}
		if !found {
			newService := serviceDetail{
				ServiceName:   serviceName,
				ServiceSize:   1,
				EndpointCount: 1,
				Urls:          []Url{url},
			}
			Services = append(Services, newService)
		}
	}

	return Services
}
