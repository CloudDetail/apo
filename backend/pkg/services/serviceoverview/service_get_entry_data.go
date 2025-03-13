// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"sort"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// GetServicesEndpointDataByEndpoints 根据输入的Endpoint填充端点总览信息
// 不按Service分组,并保留指标为空的Endpoint
// 当前只在告警分析时使用
func (s *service) GetAlertRelatedEntryData(
	startTime, endTime time.Time, namespaces []string,
	entry []response.AlertRelatedEntry,
) (res []response.AlertRelatedEntry, err error) {
	filters, entryMap := splitEntries(entry)

	var filteredEntryIdx []int
	for i := 0; i < len(filters); i++ {
		// step1 查询满足Filter的Endpoint,并返回对应的RED指标
		// RED指标包含了选定时间段内的平均值,日同比变化率和周同比变化率
		endpointsMap := s.EndpointsREDMetric(startTime, endTime, filters[i])
		// step2.. 填充Namespace信息
		_ = s.EndpointsNamespaceInfo(endpointsMap, startTime, endTime, filters[i])

		for k, metrics := range endpointsMap.MetricGroupMap {
			idx, find := entryMap[k]
			if !find || metrics == nil {
				continue
			}

			tmpSet := make(map[string]struct{})
			nsList := make([]string, 0)

			if len(nsList) > 0 {
				var isFiltered = true
				for _, ns := range metrics.NamespaceList {
					if _, find := tmpSet[ns]; find {
						continue
					}

					for _, namespace := range namespaces {
						if ns == namespace {
							isFiltered = false
						}
					}
					tmpSet[ns] = struct{}{}
					nsList = append(nsList, ns)
				}

				if isFiltered {
					filteredEntryIdx = append(filteredEntryIdx, idx)
					continue
				}
			}

			entry[idx].Namespaces = nsList

			entry[idx].Latency.Value = metrics.REDMetrics.Avg.Latency
			entry[idx].Latency.Ratio.DayOverDay = metrics.REDMetrics.DOD.Latency
			entry[idx].Latency.Ratio.WeekOverDay = metrics.REDMetrics.WOW.Latency

			entry[idx].ErrorRate.Value = metrics.REDMetrics.Avg.ErrorRate
			entry[idx].ErrorRate.Ratio.DayOverDay = metrics.REDMetrics.DOD.ErrorRate
			entry[idx].ErrorRate.Ratio.WeekOverDay = metrics.REDMetrics.WOW.ErrorRate

			entry[idx].Tps.Value = metrics.REDMetrics.Avg.TPM
			entry[idx].Tps.Ratio.DayOverDay = metrics.REDMetrics.DOD.TPM
			entry[idx].Tps.Ratio.WeekOverDay = metrics.REDMetrics.WOW.TPM
		}
	}

	if len(filteredEntryIdx) > 0 {
		sort.Ints(filteredEntryIdx)
		var res = make([]response.AlertRelatedEntry, 0, len(entry)-len(filteredEntryIdx))
		var nextIdx = 0
		for i := 0; i < len(entry); i++ {
			if i == filteredEntryIdx[nextIdx] {
				nextIdx++
				continue
			}
			res = append(res, entry[i])
		}
		entry = res
	}

	return entry, err
}

// 将需要查询数据的入口分割成多次请求,避免请求过长
func splitEntries(entries []response.AlertRelatedEntry) ([][]string, map[prom.EndpointKey]int) {
	var querySize int = 0

	var filters [][]string = make([][]string, 0)
	var services []string
	var contentKeys []string

	var entryMap = map[prom.EndpointKey]int{}
	for idx, entry := range entries {
		entryMap[prom.EndpointKey{
			SvcName:    entry.ServiceName,
			ContentKey: entry.Endpoint,
		}] = idx

		services = append(services, entry.ServiceName)
		contentKeys = append(contentKeys, entry.Endpoint)

		if querySize > 8000 || idx+1 == len(entries) {
			filters = append(filters, []string{
				prom.ServiceRegexPQLFilter, prom.RegexMultipleValue(services...),
				prom.ContentKeyRegexPQLFilter, prom.RegexMultipleValue(contentKeys...),
			})

			querySize = 0
			services = []string{}
			contentKeys = []string{}
		}
	}

	return filters, entryMap
}
