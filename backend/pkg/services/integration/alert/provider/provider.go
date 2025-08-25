// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package provider

import "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"

type Provider interface {
	GetAlerts(args map[string]any) ([]alert.AlertEvent, error)
}
