// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"sort"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// Fetch the endpoint overview information based on the input endpoints.
func (s *service) GetAlertRelatedEntryData(ctx_core core.Context,
	startTime, endTime time.Time, namespaces []string,
	entry []response.AlertRelatedEntry,
) (res []response.AlertRelatedEntry, err error) {
	filters, entryMap := splitEntries(entry)

	var filteredEntryIdx []int
	for i := 0; i < len(filters); i++ {
		endpointsMap := s.EndpointsREDMetric(startTime, endTime, filters[i])
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

// Split the entry for querying data into multiple requests to avoid overly long requests
func splitEntries(entries []response.AlertRelatedEntry) ([][]string, map[prom.EndpointKey]int) {
	var querySize int = 0

	var filters [][]string = make([][]string, 0)
	var services []string
	var contentKeys []string

	var entryMap = map[prom.EndpointKey]int{}
	for idx, entry := range entries {
		entryMap[prom.EndpointKey{
			SvcName:	entry.ServiceName,
			ContentKey:	entry.Endpoint,
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
