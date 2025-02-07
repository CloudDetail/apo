// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package enrich

import (
	"sync"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type AlertEnricher struct {
	*alert.SourceFrom
	Enrichers []Enricher

	enrichersMutex sync.RWMutex

	// RuleID -> Enricher
	EnricherMap sync.Map
}

func (e *AlertEnricher) Enrich(events []alert.AlertEvent) error {
	e.enrichersMutex.RLock()
	defer e.enrichersMutex.RUnlock()
	for i := 0; i < len(events); i++ {
		for _, enricher := range e.Enrichers {
			enricher.Enrich(&events[i])
		}
	}
	return nil
}

func (e *AlertEnricher) RemoveRuleByDeletedSchema(schema string) {
	e.enrichersMutex.Lock()
	defer e.enrichersMutex.Unlock()

	e.removeEnricherBySchema(schema)
}

func (e *AlertEnricher) removeEnricherBySchema(schema string) {
	result := make([]Enricher, 0, len(e.Enrichers)-1)
	for _, enricher := range e.Enrichers {
		if tagEnricher, ok := enricher.(*TagEnricher); ok {
			if tagEnricher.Schema == schema {
				e.EnricherMap.Delete(tagEnricher.RuleID())
				continue
			}
		}
		result = append(result, enricher)
	}
	e.Enrichers = result
}

type Enricher interface {
	RuleID() string
	RuleOrder() int
	Enrich(*alert.AlertEvent)
}
