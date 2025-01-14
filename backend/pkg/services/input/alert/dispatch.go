// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"sync"

	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
	"github.com/CloudDetail/apo/backend/pkg/services/input/alert/enrich"
)

type Dispatcher struct {
	// SourceID -> *AlertSourceEnricher
	EnricherMap sync.Map
	// SourceName -> *AlertSourceEnricher
	SourceName2EnricherMap sync.Map
}

func (d *Dispatcher) DispatchEvents(
	source alert.SourceFrom, events []alert.AlertEvent,
) error {
	var enricher *enrich.AlertEnricher
	if len(source.SourceID) > 0 {
		enricherPtr, find := d.EnricherMap.Load(source.SourceID)
		if find {
			enricher = enricherPtr.(*enrich.AlertEnricher)
		}
	} else {
		enricherPtr, find := d.SourceName2EnricherMap.Load(source.SourceName)
		if find {
			enricher = enricherPtr.(*enrich.AlertEnricher)
		}
	}

	if enricher == nil {
		// alertsource not existed
		return alert.ErrAlertSourceNotExist{}
	}
	return enricher.Enrich(events)
}

// AddOrUpdateAlertSourceRule
func (d *Dispatcher) AddOrUpdateAlertSourceRule(
	source alert.SourceFrom, enricher *enrich.AlertEnricher,
) {
	oldEnricherPtr, loaded := d.EnricherMap.Swap(source.SourceID, enricher)
	if loaded {
		oldEnricher := oldEnricherPtr.(*enrich.AlertEnricher)
		d.SourceName2EnricherMap.Store(oldEnricher.SourceName, enricher)
		return
	}

	if len(source.SourceName) > 0 && len(source.SourceType) > 0 {
		d.SourceName2EnricherMap.Store(source.SourceName, enricher)
	}
}

func (d *Dispatcher) AddAlertSource(
	source *alert.SourceFrom, enricher *enrich.AlertEnricher,
) {
	d.EnricherMap.Store(source.SourceID, enricher)
	if len(source.SourceName) > 0 && len(source.SourceType) > 0 {
		d.SourceName2EnricherMap.Store(source.SourceName, enricher)
	}
}

func (d *Dispatcher) DeleteAlertSource(
	source *alert.AlertSource,
) {
	enricher, loaded := d.EnricherMap.LoadAndDelete(source.SourceID)
	if !loaded {
		return
	}

	d.SourceName2EnricherMap.Delete(enricher.(*enrich.AlertEnricher).SourceName)
}
