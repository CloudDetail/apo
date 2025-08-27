// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

type ProviderType struct {
	Name      string
	ParamSpec ParamSpec
	factory   func(sourceFrom alert.SourceFrom, params alert.AlertSourceParams) Provider

	SupportPull           bool
	SupportWebhookInstall bool
}

func (p *ProviderType) New(sourceFrom alert.SourceFrom, params alert.AlertSourceParams) Provider {
	if p.factory == nil {
		return nil
	}
	return p.factory(sourceFrom, params)
}

type Provider interface {
	PullAlerts(args GetAlertParams) ([]alert.AlertEvent, error)
	// Install or update webhook
	SetupWebhook(ctx core.Context, webhookURL string, params alert.AlertSourceParams) error
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

func getWebhookAddress(baseURL string, sourceID string) string {
	return fmt.Sprintf("%s/api/alertinput/event/source?sourceId=%s", baseURL, sourceID)
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
