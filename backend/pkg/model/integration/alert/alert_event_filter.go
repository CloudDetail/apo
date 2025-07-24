// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
)

type AlertEventFilter struct {
	Source       string
	Group        string
	Name         string
	EventID      string
	Severity     string
	Status       string
	WithMutation bool

	GroupIDs []string

	*AlertTagsFilter
}

// TagsFilter using field:tags to filter alert
// Use OR to connect different conditions
type AlertTagsFilter struct {
	ServiceEndpoints []model.EndpointKey
	model.RelatedInstances
}
