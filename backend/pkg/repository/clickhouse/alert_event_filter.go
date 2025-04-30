// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func extractAlertEventFilter(filter *alert.AlertEventFilter) *whereSQL {
	if filter == nil {
		return ALWAYS_TRUE
	}

	var basicFilters []*whereSQL
	basicFilters = append(basicFilters,
		equalsIfNotEmpty("source", filter.Source),
		// EqualsIfNotEmpty("group", filter.Group),
		equalsIfNotEmpty("name", filter.Name),
		equalsIfNotEmpty("id", filter.EventID),
		equalsIfNotEmpty("severity", filter.Severity),
		equalsIfNotEmpty("status", filter.Status),
	)

	if len(filter.Group) > 0 && filter.WithMutation {
		basicFilters = append(basicFilters,
			in("group", clickhouse.ArraySet{
				filter.Group,
				"mutation-" + filter.Group,
			}))
	} else if len(filter.Group) > 0 {
		basicFilters = append(basicFilters, equals("group", filter.Group))
	}

	if !filter.WithMutation {
		basicFilters = append(basicFilters, notLike("group", "mutation%"))
	}

	basicSQL := mergeWheres(AndSep, basicFilters...)

	if filter.AlertTagsFilter == nil {
		return basicSQL
	}

	// TODO use tagFilter to decrease events
	// return MergeWheres(AndSep, basicSQL, extractAlertTagsFilter(filter.AlertTagsFilter))
	return basicSQL
}
