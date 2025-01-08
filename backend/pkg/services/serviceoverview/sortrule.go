// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"sort"

	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// Sort according to DoD/WoW Growth Rate threshold
func sortByDODThreshold(endpoints []*prom.EndpointMetrics) {
	sort.SliceStable(endpoints, func(i, j int) bool {
		// Sort by count first
		if endpoints[i].AlertCount != endpoints[j].AlertCount {
			return endpoints[i].AlertCount > endpoints[j].AlertCount
		}
		// Sort by error rate when equal to 3
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
		// count = 2 comparison method
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

// sort by mutation
func sortByMutation(endpoints []*prom.EndpointMetrics) {
	for i, _ := range endpoints {
		// The average error rate and 1m error rate cannot be found, and the mutation rate is 0
		if endpoints[i].REDMetrics.Avg.ErrorRate == nil && endpoints[i].REDMetrics.Realtime.ErrorRate == nil {
			endpoints[i].Avg1MinErrorMutationRate = 0
		}
		// The average error rate can be found, but the error rate of 1m cannot be found.
		if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[i].REDMetrics.Realtime.ErrorRate == nil {
			// Average error rate is 0: mutation rate is 0
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = 0
			}
			// Average error rate is not 0, mutation rate is -1
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = -1
			}
		}
		// The average error rate cannot be found, and the error rate of 1m cannot be found.
		if endpoints[i].REDMetrics.Avg.ErrorRate == nil && endpoints[i].REDMetrics.Realtime.ErrorRate != nil {
			// 1m error rate is 0, mutation rate is 0
			if endpoints[i].REDMetrics.Realtime.ErrorRate != nil && *endpoints[i].REDMetrics.Realtime.ErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = 0
			}
			// 1m error rate is not 0, mutation rate is max
			if endpoints[i].REDMetrics.Realtime.ErrorRate != nil && *endpoints[i].REDMetrics.Realtime.ErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
		}
		// The average error rate cannot be found, and the error rate of 1m cannot be found.
		if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[i].REDMetrics.Realtime.ErrorRate != nil {
			// 1m error rate is 0, mutation rate is 0
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate == 0 {
				endpoints[i].Avg1MinErrorMutationRate = RES_MAX_VALUE
			}
			// 1m error rate is not 0, mutation rate is max
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[i].REDMetrics.Realtime.ErrorRate != nil && *endpoints[i].REDMetrics.Avg.ErrorRate != 0 {
				endpoints[i].Avg1MinErrorMutationRate = *endpoints[i].REDMetrics.Realtime.ErrorRate / *endpoints[i].REDMetrics.Avg.ErrorRate
			}
		}
		//latency
		// The average delay and 1m delay cannot be found, and the mutation rate is 0 (impossible)
		if endpoints[i].REDMetrics.Avg.Latency == nil && endpoints[i].REDMetrics.Realtime.Latency == nil {
			endpoints[i].Avg1MinLatencyMutationRate = 0
		}
		// The average delay can be found, but the 1m delay cannot be found.
		if endpoints[i].REDMetrics.Avg.Latency != nil && endpoints[i].REDMetrics.Realtime.Latency == nil {
			// average delay is 0: mutation rate is 0
			if endpoints[i].REDMetrics.Avg.Latency != nil && *endpoints[i].REDMetrics.Avg.Latency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = 0
			}
			// average delay is not 0, mutation rate is -1
			if endpoints[i].REDMetrics.Avg.Latency != nil && *endpoints[i].REDMetrics.Avg.Latency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = -1
			}
		}
		// The average error rate cannot be found, and the error rate of 1m cannot be found.
		if endpoints[i].REDMetrics.Avg.Latency == nil && endpoints[i].REDMetrics.Realtime.Latency != nil {
			// 1m delay is 0, mutation rate is 0
			if endpoints[i].REDMetrics.Realtime.Latency != nil && *endpoints[i].REDMetrics.Realtime.Latency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = 0
			}
			// 1m delay is not 0, mutation rate is max
			if endpoints[i].REDMetrics.Realtime.Latency != nil && *endpoints[i].REDMetrics.Realtime.Latency != 0 {
				endpoints[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
		}
		// The average error rate cannot be found, and the error rate of 1m cannot be found.
		if endpoints[i].REDMetrics.Avg.Latency != nil && endpoints[i].REDMetrics.Realtime.Latency != nil {
			// average delay is 0, mutation rate is max
			if endpoints[i].REDMetrics.Avg.Latency != nil && *endpoints[i].REDMetrics.Avg.Latency == 0 {
				endpoints[i].Avg1MinLatencyMutationRate = RES_MAX_VALUE
			}
			// average delay is not 0, mutation rate is 1m delay/average delay
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
		// Case 1: If there is an error rate mutation rate greater than 1 (error rate increases)
		if endpoints[i].Avg1MinErrorMutationRate > 1 || endpoints[j].Avg1MinErrorMutationRate > 1 {
			return endpoints[i].Avg1MinErrorMutationRate > endpoints[j].Avg1MinErrorMutationRate
		}

		// Case 2: If the error rate mutation rate is less than or equal to 1
		if endpoints[i].Avg1MinErrorMutationRate <= 1 && endpoints[j].Avg1MinErrorMutationRate <= 1 {
			// Sort by delayed mutation rate first, with larger ones first
			if endpoints[i].Avg1MinLatencyMutationRate != endpoints[j].Avg1MinLatencyMutationRate {
				return endpoints[i].Avg1MinLatencyMutationRate > endpoints[j].Avg1MinLatencyMutationRate
			}

			// Sort by error rate if the delay mutation rate is the same
			if endpoints[i].REDMetrics.Realtime.ErrorRate != nil && endpoints[j].REDMetrics.Realtime.ErrorRate != nil {
				return *endpoints[i].REDMetrics.Realtime.ErrorRate > *endpoints[j].REDMetrics.Realtime.ErrorRate
			}
			// If one error rate is nil and the other is not nil, the error rate is not nil.
			if endpoints[i].REDMetrics.Realtime.ErrorRate != nil && endpoints[j].REDMetrics.Realtime.ErrorRate == nil {
				return true
			}
			if endpoints[i].REDMetrics.Realtime.ErrorRate == nil && endpoints[j].REDMetrics.Realtime.ErrorRate != nil {
				return false
			}
			// Sort by error rate if the delay mutation rate is the same
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[j].REDMetrics.Avg.ErrorRate != nil {
				return *endpoints[i].REDMetrics.Avg.ErrorRate > *endpoints[j].REDMetrics.Avg.ErrorRate
			}
			// If one error rate is nil and the other is not nil, the error rate is not nil.
			if endpoints[i].REDMetrics.Avg.ErrorRate != nil && endpoints[j].REDMetrics.Avg.ErrorRate == nil {
				return true
			}
			if endpoints[i].REDMetrics.Avg.ErrorRate == nil && endpoints[j].REDMetrics.Avg.ErrorRate != nil {
				return false
			}
			// if error rates are the same or nil, sort by latency
			if endpoints[i].REDMetrics.Realtime.Latency != nil && endpoints[j].REDMetrics.Realtime.Latency != nil {
				return *endpoints[i].REDMetrics.Realtime.Latency > *endpoints[j].REDMetrics.Realtime.Latency
			}
			if endpoints[i].REDMetrics.Realtime.Latency != nil && endpoints[j].REDMetrics.Realtime.Latency == nil {
				return true
			}
			if endpoints[i].REDMetrics.Realtime.Latency == nil && endpoints[j].REDMetrics.Realtime.Latency != nil {
				return false
			}
			// if error rates are the same or nil, sort by latency
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
		// No return if there is no data
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
