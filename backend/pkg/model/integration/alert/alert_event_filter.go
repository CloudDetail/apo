// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/CloudDetail/apo/backend/pkg/model"

type AlertEventFilter struct {
	Source       string `form:"source"`
	Group        string `form:"group"`
	Name         string `form:"name"`
	ID           string `form:"id"`
	Severity     string `form:"severity"`
	Status       string `form:"status"`
	WithMutation bool   `form:"withMutation"`

	*AlertTagsFilter
}

// TagsFilter using field:tags to filter alert
// Use OR to connect different conditions
type AlertTagsFilter struct {
	ServiceEndpoints []model.EndpointKey
	model.RelatedInstances
}
