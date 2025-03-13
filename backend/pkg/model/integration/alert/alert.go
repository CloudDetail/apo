// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

// Each alert has a unique alertId, and usually contains multi events
type Alert struct {
	Source   string `json:"source,omitempty" ch:"source"`
	SourceID string `json:"sourceId,omitempty" ch:"source_id"`

	AlertID string `json:"alertId" ch:"alert_id"`
	Group   string `ch:"group" json:"group,omitempty"`
	Name    string `ch:"name" json:"name,omitempty"`

	EnrichTags map[string]string `json:"tags" ch:"tags" mapstructure:"enrich_tags"`
	// HACK the existing clickhouse query uses `tags` as the filter field
	// so enrichTags in ch is named as 'tags' to filter new alertInput
	Tags RawTags `json:"rawTags" ch:"raw_tags" mapstructure:"tags"`
}

type RawTags map[string]any

// Impl clickhouse OrderedMap interface to accept map[string]string
func (t *RawTags) Get(key any) (any, bool) {
	if *t == nil {
		return nil, false
	}
	strKey, ok := key.(string)
	if !ok {
		return nil, false
	}

	value, exists := (*t)[strKey]
	return value, exists
}

func (t *RawTags) Put(key any, value any) {
	if *t == nil {
		*t = make(RawTags)
	}

	strKey, ok := key.(string)
	if !ok {
		return
	}

	(*t)[strKey] = value
}

func (t *RawTags) Keys() <-chan any {
	ch := make(chan any)

	go func() {
		defer close(ch)

		for key := range *t {
			ch <- key
		}
	}()

	return ch
}

type AlertWithEventCount struct {
	Alert

	Count uint64 `json:"count" ch:"count"`
}
