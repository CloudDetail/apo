// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	pd "github.com/PagerDuty/go-pagerduty"
	"github.com/go-faster/errors"
)

var PagerDutyProviderType = ProviderType{
	Name: "pagerduty",
	ParamSpec: ParamSpec{
		Name: "root",
		Type: JSONTypeObject,
		Children: []ParamSpec{
			{
				Name:     "api_key",
				Type:     JSONTypeString,
				Optional: false,
				Children: []ParamSpec{},
				Desc:     "API Key",
				DescEN:   "API Key",
			},
		},
	},
	factory:               NewPagerDutyProvider,
	SupportPull:           false,
	SupportWebhookInstall: true,
}

type PagerDutyProvider struct {
	client *pdClientEx

	source alert.AlertSource
}

func NewPagerDutyProvider(source alert.AlertSource, params alert.AlertSourceParams) Provider {
	apiKey := params.GetString("api_key")
	c := pd.NewClient(apiKey)
	cEx := &pdClientEx{Client: c}

	return &PagerDutyProvider{
		client: cEx,
		source: source,
	}
}

func (f *PagerDutyProvider) GetAlertSource() alert.AlertSource {
	return f.source
}

func (f *PagerDutyProvider) SetAlertSource(source alert.AlertSource) {
	apiKey := source.Params.Obj.GetString("api_key")
	c := pd.NewClient(apiKey)
	cEx := &pdClientEx{Client: c}

	f.client = cEx
	f.source = source
}

func (f *PagerDutyProvider) SetupWebhook(ctx core.Context, webhookURL string) error {
	subscriptions, err := f.client.ListWebhookSubscription(ctx)
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		if strings.HasSuffix(subscription.DeliveryMethod.URL, f.source.SourceID) {
			err = f.client.DeleteWebhookSubscription(ctx, subscription.Type)
			if err != nil {
				log.Printf("delete existed pagerduty webhook subscription failed, err: %v", err)
			}
			break
		}
	}

	err = f.client.CreateWebhookSubscriptions(ctx, webhookURL, []PDCustomHeader{})
	if err != nil {
		return err
	}
	return nil
}

func (f *PagerDutyProvider) PullAlerts(args GetAlertParams) ([]alert.AlertEvent, error) {
	// TODO unsupported pull incident from pagerDuty and transform into alerts
	return nil, nil
}

func (f *PagerDutyProvider) ClearUP(ctx core.Context) {
	subscriptions, err := f.client.ListWebhookSubscription(ctx)
	if err != nil {
		return
	}

	for _, subscription := range subscriptions {
		if strings.HasSuffix(subscription.DeliveryMethod.URL, f.source.SourceID) {
			err = f.client.DeleteWebhookSubscription(ctx, subscription.Type)
			if err != nil {
				log.Printf("delete existed pagerduty webhook subscription failed, err: %v", err)
			}
			break
		}
	}
}

type pdClientEx struct {
	*pd.Client
}

type PDCustomHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type pdDeliveryMethod struct {
	Type          string           `json:"type"`
	URL           string           `json:"url"`
	CustomHeaders []PDCustomHeader `json:"custom_headers"`
}

type pdFilter struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type pdWebhookSubscription struct {
	DeliveryMethod pdDeliveryMethod `json:"delivery_method"`
	Description    string           `json:"description"`
	Events         []string         `json:"events"`
	Filter         pdFilter         `json:"filter"`
	// OAuthClientID  string         `json:"oauth_client_id"`
	Type string `json:"type"`
}

type pdCreateWebhookSubscriptionRequest struct {
	WebhookSubscription pdWebhookSubscription `json:"webhook_subscription"`
}

const (
	PagerDutyWebhookSubscriptionType   = "webhook_subscription"
	PagerDutyBaseURL                   = "https://api.pagerduty.com"
	PagerDutyWebhookSubScriptionAPIURL = PagerDutyBaseURL + "/webhook_subscriptions"
)

func (c *pdClientEx) CreateWebhookSubscriptions(ctx context.Context, webhookAddress string, customHeaders []PDCustomHeader) error {
	reqBody := &pdCreateWebhookSubscriptionRequest{
		WebhookSubscription: pdWebhookSubscription{
			DeliveryMethod: pdDeliveryMethod{
				Type:          "http_delivery_method",
				URL:           webhookAddress,
				CustomHeaders: customHeaders,
			},
			Description: "APO Incident Webhook Subscription",
			Events: []string{
				// "incident.acknowledged",
				// "incident.annotated",
				// "incident.delegated",
				// "incident.escalated",
				"incident.priority_updated",
				// "incident.reassigned",
				"incident.reopened",
				"incident.resolved",
				// "incident.responder.added",
				// "incident.responder.replied",
				"incident.triggered",
				// "incident.unacknowledged",
			},
			Filter: pdFilter{
				Type: "account_reference",
			},
			Type: PagerDutyWebhookSubscriptionType,
		},
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", PagerDutyWebhookSubScriptionAPIURL, bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		return err
	}

	response, err := c.Do(req, true)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return errors.Errorf("read response body failed: %w", err)
		}
		return errors.Errorf("unacceptable response code %d, response: %s", response.StatusCode, string(body))
	}
	return nil
}

func (c *pdClientEx) ListWebhookSubscription(ctx context.Context) ([]pdWebhookSubscription, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", PagerDutyWebhookSubScriptionAPIURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unacceptable response code %d, response: %s", resp.StatusCode, string(body))
	}

	type subscriptions struct {
		WebhookSubscriptions []pdWebhookSubscription `json:"webhook_subscriptions"`
	}
	var respBody subscriptions
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		return nil, errors.Errorf("unmarshal response body failed: %w", err)
	}
	return respBody.WebhookSubscriptions, nil
}

func (c *pdClientEx) DeleteWebhookSubscription(ctx context.Context, subscriptionID string) error {
	deleteURL := PagerDutyWebhookSubScriptionAPIURL + "/" + subscriptionID

	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Errorf("read response body failed: %w", err)
		}
		return errors.Errorf("unacceptable response code %d, response: %s", resp.StatusCode, string(body))
	}
	return nil
}
