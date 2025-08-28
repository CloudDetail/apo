// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type ProviderType struct {
	Name      string
	ParamSpec ParamSpec
	factory   func(source alert.AlertSource, params alert.AlertSourceParams) Provider

	SupportPull           bool
	SupportWebhookInstall bool
}

func (p *ProviderType) New(source alert.AlertSource, params alert.AlertSourceParams) Provider {
	if p.factory == nil {
		return nil
	}
	return p.factory(source, params)
}

type Provider interface {
	GetAlertSource() alert.AlertSource
	SetAlertSource(source alert.AlertSource)

	PullAlerts(args GetAlertParams) ([]alert.AlertEvent, error)
	// Install or update webhook
	SetupWebhook(ctx core.Context, webhookURL string) error

	ClearUP(ctx core.Context)
}

type GetAlertParams struct {
	From time.Time
	To   time.Time
	// UnstructuredParams map[string]any
}

var ProviderRegistry = map[string]ProviderType{
	"datadog":   DatadogProviderType,
	"pagerduty": PagerDutyProviderType, // TODO pull support

	// Do not support auto install webhook or pull alerts now
	"prometheus": BasicEncoder("Prometheus"), // TODO pull support
	"zabbix":     BasicEncoder("Zabbix"),
	"json":       BasicEncoder("JSON"),
}

func BasicEncoder(name string) ProviderType {
	return ProviderType{
		Name: name,
		ParamSpec: ParamSpec{
			Name: "root",
			Type: JSONTypeObject,
		},
		factory: nil,

		SupportPull:           false,
		SupportWebhookInstall: false,
	}
}
