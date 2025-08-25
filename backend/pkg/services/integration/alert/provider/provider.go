// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type ProviderType struct {
	Name      string
	ParamSpec ParamSpec
	New       ProviderFactory

	WithPullOptions bool
}

type ProviderFactory func(sourceFrom alert.SourceFrom, params alert.AlertSourceParams) Provider

type Provider interface {
	PullAlerts(args GetAlertParams) ([]alert.AlertEvent, error)

	// InstallWebhook() error
}

type GetAlertParams struct {
	From time.Time
	To   time.Time
	// UnstructuredParams map[string]any
}

var ProviderRegistry = map[string]ProviderType{
	"datadog": {
		Name:      "Datadog",
		ParamSpec: DatadogParamSpec,
		New:       NewDatadogProvider,
	},
}
