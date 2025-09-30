// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/provider/lifecycle"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

var DatadogProviderType = ProviderType{
	Name: "Datadog",
	ParamSpec: ParamSpec{
		Name: "root",
		Type: JSONTypeObject,
		Children: []ParamSpec{
			{
				Name:   "site",
				Type:   JSONTypeString,
				Desc:   "DataDog地址,示例: datadoghq.com",
				DescEN: "DataDog site, example: datadoghq.com",
			},
			{
				Name:   "apiKey",
				Type:   JSONTypeString,
				Desc:   "DataDog API Key",
				DescEN: "DataDog API key",
			},
			{
				Name:   "appKey",
				Type:   JSONTypeString,
				Desc:   "DataDog APP Key",
				DescEN: "DataDog APP key",
			},
		},
	},
	factory: NewDatadogProvider,

	SupportPull:           true,
	SupportWebhookInstall: true,
}

type DatadogProvider struct {
	client *datadog.APIClient

	authCtx context.Context
	source  alert.AlertSource
}

func NewDatadogProvider(source alert.AlertSource, params alert.AlertSourceParams) Provider {
	configuration := datadog.NewConfiguration()
	client := datadog.NewAPIClient(configuration)

	ctx := context.WithValue(context.Background(),
		datadog.ContextServerVariables,
		map[string]string{
			"site": params.GetString("site"),
		},
	)

	ctx = context.WithValue(ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {Key: params.GetString("apiKey")},
			"appKeyAuth": {Key: params.GetString("appKey")},
		},
	)

	return &DatadogProvider{
		client:  client,
		authCtx: ctx,
		source:  source,
	}
}

func (f *DatadogProvider) GetAlertSource() alert.AlertSource {
	return f.source
}

func (f *DatadogProvider) SetAlertSource(source alert.AlertSource) {
	params := source.Params.Obj

	ctx := context.WithValue(context.Background(),
		datadog.ContextServerVariables,
		map[string]string{
			"site": params.GetString("site"),
		},
	)

	ctx = context.WithValue(ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {Key: params.GetString("apiKey")},
			"appKeyAuth": {Key: params.GetString("appKey")},
		},
	)

	f.authCtx = ctx
	f.source = source
}

const DDWebhookPayload = `
{
	"alert_id": "$ALERT_ID",
	"agg_reg_key": "$AGGREG_KEY",
	"alert_cycle_key": "$ALERT_CYCLE_KEY",
	"alert_query": "$ALERT_QUERY",
	"alert_title": "$ALERT_TITLE",
	"alert_transition": "$ALERT_TRANSITION",
	"alert_priority": "$ALERT_PRIORITY",
	"id": "$ID",
	"event_title": "$EVENT_TITLE",
	"event_msg": "$TEXT_ONLY_MSG",
	"last_updated": "$LAST_UPDATED",
	"tags": "$TAGS",
	"date": "$DATE",
	"org_id": "$ORG_ID",
	"org_name": "$ORG_NAME"
}`

func (f *DatadogProvider) SetupWebhook(ctx core.Context, webhookURL string) error {
	cCtx := newCContext(ctx, f.authCtx) // Combine req.Done and f.authCtx

	webhookName := fmt.Sprintf("webhook-apo-%s", f.source.SourceID[:8])

	// step1 setupWebhook
	if err := f.setupWebhook(cCtx, webhookName, webhookURL); err != nil {
		return err
	}
	// step2 update monitor message
	return f.updateMonitor(cCtx, webhookName)
}

func (f *DatadogProvider) updateMonitor(ctx CContext, webhookName string) error {
	monitorAPI := datadogV1.NewMonitorsApi(f.client)

	monitors, resp, err := monitorAPI.ListMonitors(ctx)
	if err != nil {
		log.Printf("list monitor failed, err: %v, full response: %v", err, resp)
		return err
	}

	for _, monitor := range monitors {
		message := monitor.GetMessage()
		if strings.Contains(message, "@"+webhookName) {
			continue
		}

		monitorReq := datadogV1.NewMonitorUpdateRequest()
		monitorReq.SetMessage(message + " @" + webhookName)

		_, resp, err := monitorAPI.UpdateMonitor(ctx, monitor.GetId(), *monitorReq)
		if err != nil {
			log.Printf("update monitor failed, monitor: %d/%s, err: %v, full response: %v", monitor.GetId(), monitor.GetName(), err, resp)
			// Ignore error
			//  <custom_check> monitor can not update since group is not defined
		}
	}
	return nil
}

func (f *DatadogProvider) clearupMonitor(ctx context.Context, webhookName string) error {
	monitorAPI := datadogV1.NewMonitorsApi(f.client)
	monitors, resp, err := monitorAPI.ListMonitors(ctx)
	if err != nil {
		log.Printf("list monitor failed, err: %v, full response: %v", err, resp)
		return err
	}

	for _, monitor := range monitors {
		message := monitor.GetMessage()
		if !strings.Contains(message, "@"+webhookName) {
			continue
		}

		message = strings.ReplaceAll(message, "@"+webhookName, "")
		monitorReq := datadogV1.NewMonitorUpdateRequest()
		monitorReq.SetMessage(message)
		_, resp, err := monitorAPI.UpdateMonitor(ctx, monitor.GetId(), *monitorReq)
		if err != nil {
			log.Printf("clearup webhook in monitor failed, monitor: %d/%s, err: %v, full response: %v", monitor.GetId(), monitor.GetName(), err, resp)
			// Ignore error
			//  <custom_check> monitor can not update since group is not defined
		}
	}
	return nil
}

// Combine Context
type CContext struct {
	context.Context
	valueCtx context.Context
}

func newCContext(ctx context.Context, valueCtx context.Context) CContext {
	return CContext{
		Context:  ctx,
		valueCtx: valueCtx,
	}
}

func (c CContext) Value(key any) any {
	if v := c.valueCtx.Value(key); v != nil {
		return v
	}
	return c.Context.Value(key)
}

func (f *DatadogProvider) setupWebhook(ctx CContext, webhookName string, webhookURL string) error {
	webhookAPI := datadogV1.NewWebhooksIntegrationApi(f.client)

	// Check history
	_, resp, err := webhookAPI.GetWebhooksIntegration(ctx, webhookName)
	if resp.StatusCode == 404 {
		// create new
		integration := datadogV1.NewWebhooksIntegration(webhookName, webhookURL)
		integration.SetPayload(DDWebhookPayload)
		_, resp, err = webhookAPI.CreateWebhooksIntegration(ctx, *integration)
		if err != nil {
			log.Printf("create webhook failed, err: %v, full response: %v", err, resp)
			return err
		}
		return err
	}

	if err != nil {
		return err
	}

	// update
	updateReq := datadogV1.NewWebhooksIntegrationUpdateRequest()
	updateReq.SetPayload(DDWebhookPayload)
	updateReq.SetUrl(webhookURL)

	_, resp, err = webhookAPI.UpdateWebhooksIntegration(ctx, webhookName, *updateReq)
	if err != nil {
		log.Printf("update webhook failed, err: %v, full response: %v", err, resp)
		return err
	}
	return err
}

func (f *DatadogProvider) PullAlerts(args GetAlertParams) ([]alert.AlertEvent, error) {
	eventAPI := datadogV2.NewEventsApi(f.client)

	start := args.From
	end := args.To

	resChan, cancel := eventAPI.ListEventsWithPagination(f.authCtx, *datadogV2.NewListEventsOptionalParameters().
		WithPageLimit(1000).
		WithSort(datadogV2.EVENTSSORT_TIMESTAMP_ASCENDING).
		WithFilterFrom(strconv.FormatInt(start.UnixMilli(), 10)).
		WithFilterTo(strconv.FormatInt(end.UnixMilli(), 10)).
		WithFilterQuery("source:alert"),
	)
	defer cancel()

	receivedTime := time.Now()

	var events []alert.AlertEvent
	var err error
	for item := range resChan {
		if item.Error != nil {
			err = item.Error
			break
		}

		if event := f.parseEventItem(item, receivedTime); event != nil {
			events = append(events, *event)
		}
	}

	return events, err
}

func (f *DatadogProvider) parseEventItem(item datadog.PaginationResult[datadogV2.EventResponse], receivedTime time.Time) *alert.AlertEvent {
	attrs := item.Item.Attributes
	nestedAttrs := item.Item.Attributes.Attributes

	alertID := fmt.Sprintf("%s-%s", f.source.SourceID[:8], nestedAttrs.GetAggregationKey())
	eventID := fmt.Sprintf("%s-%s", f.source.SourceID[:8], item.Item.GetId())

	if lifecycle.AlertLifeCycle.CheckEventSeen(eventID) {
		// repeated with webhook event, ignore
		return nil
	}

	monitor := nestedAttrs.GetMonitor()
	severity := f.getSeverityFromMonitor(monitor)

	createTime := f.calculateCreateTime(nestedAttrs)
	status, endTime := f.getStatusAndEndTimeFromMonitor(monitor, nestedAttrs)

	// Only cache for webhook
	lifecycle.AlertLifeCycle.CacheEventStatus(alertID, status, createTime)

	tags := buildDDTags(nestedAttrs, attrs.GetTags())
	group := getGroup(nestedAttrs, attrs.GetTags())

	return &alert.AlertEvent{
		Alert: alert.Alert{
			Source:     f.source.SourceName,
			SourceID:   f.source.SourceID,
			AlertID:    alertID,
			Group:      group,
			Name:       nestedAttrs.GetTitle(),
			EnrichTags: make(map[string]string),
			Tags:       tags,
		},
		EventID:      eventID,
		Detail:       buildDDDetail(attrs.GetMessage(), tags),
		CreateTime:   createTime,
		UpdateTime:   time.UnixMilli(nestedAttrs.GetTimestamp()),
		EndTime:      endTime,
		ReceivedTime: receivedTime,
		Severity:     severity,
		Status:       status,
	}
}

func (f *DatadogProvider) getSeverityFromMonitor(monitor datadogV2.MonitorType) string {
	var priority = alert.SeverityUnknownLevel
	if priorityLevel, find := monitor.AdditionalProperties["priority"]; find {
		priority, ok := priorityLevel.(float64)
		if !ok {
			return alert.SeverityUnknownLevel
		}
		if p, find := DDPriorityMap[priority]; find {
			return p
		}
		return alert.SeverityUnknownLevel
	}
	return priority
}

func (f *DatadogProvider) getStatusAndEndTimeFromMonitor(monitor datadogV2.MonitorType, nestedAttrs *datadogV2.EventAttributes) (string, time.Time) {
	var status = alert.StatusFiring
	var endTime time.Time
	if transition, find := monitor.AdditionalProperties["transition"]; find {
		status = getDDStatus(transition)

		if status == alert.StatusResolved {
			endTime = time.UnixMilli(nestedAttrs.GetTimestamp())
		}
	}
	return status, endTime
}

func (f *DatadogProvider) calculateCreateTime(nestedAttrs *datadogV2.EventAttributes) time.Time {
	if nestedAttrs.GetDuration() > 0 {
		return time.UnixMilli(nestedAttrs.GetTimestamp() - nestedAttrs.GetDuration()/1e6)
	}
	return time.UnixMilli(nestedAttrs.GetTimestamp())
}

func (f *DatadogProvider) ClearUP(ctx core.Context) {
	cCtx := newCContext(ctx, f.authCtx) // Combine req.Done and f.authCtx
	webhookName := fmt.Sprintf("webhook-apo-%s", f.source.SourceID[:8])
	// update monitor

	err := f.clearupMonitor(cCtx, webhookName)
	if err != nil {
		log.Printf("clearup monitor failed, err: %v", err)
		// Ignored
	}
	// remove webhook
	err = f.removeWebhook(cCtx, webhookName)
	if err != nil {
		log.Printf("remove webhook failed, err: %v", err)
		// Ignored
	}
}

func (f *DatadogProvider) removeWebhook(cCtx CContext, webhookName string) error {
	webhookAPI := datadogV1.NewWebhooksIntegrationApi(f.client)
	_, err := webhookAPI.DeleteWebhooksIntegration(cCtx, webhookName)
	return err
}

/*
*

	{
		... // resp tags to map
		"attr": {
			... // attr tags to map
			"title": attrs.GetTitle(),
			"service": attrs.GetService(),
			"monitor": {
				... // monitor tags to map
				"id": attrs.GetMonitorId(),
				"name": monitor.GetName(),
				"query": monitor.GetQuery(),
			},
		},
	}

*
*/
func buildDDTags(attrs *datadogV2.EventAttributes, eventTags []string) map[string]any {
	tags := make(map[string]any)

	for _, tagStr := range eventTags {
		tag := strings.Split(tagStr, ":")
		if len(tag) == 2 {
			tags[tag[0]] = tag[1]
		} else {
			tags[tagStr] = ""
		}
	}

	for _, tagStr := range attrs.GetTags() {
		tag := strings.Split(tagStr, ":")
		if len(tag) == 2 {
			tags[tag[0]] = tag[1]
		} else {
			tags[tagStr] = ""
		}
	}

	monitor := attrs.GetMonitor()
	for _, tagStr := range monitor.GetTags() {
		tag := strings.Split(tagStr, ":")
		if len(tag) == 2 {
			tags[tag[0]] = tag[1]
		} else {
			tags[tagStr] = ""
		}
	}
	return tags
}

var (
	DDPriorityMap = map[float64]string{
		1: alert.SeverityCriticalLevel,
		2: alert.SeverityErrorLevel,
		3: alert.SeverityWarnLevel,
		4: alert.SeverityInfoLevel,
		5: alert.SeverityInfoLevel,
	}
)

func getDDStatus(transition any) string {
	if transition == nil {
		return alert.StatusFiring
	}

	transitionMap, ok := transition.(map[string]any)
	if !ok {
		log.Printf("get datadog alert status failed, transition is not expected map, default set to firing")
		return alert.StatusFiring
	}
	if status, find := transitionMap["transition_type"]; find && status == "alert recovery" {
		return alert.StatusResolved
	}
	return alert.StatusFiring
}

func buildDDDetail(message string, tags map[string]any) string {
	detail := make(map[string]any)
	detail["summary"] = message
	detail["description"] = tags
	detailBytes, err := json.Marshal(detail)
	if err != nil {
		return ""
	}
	return string(detailBytes)
}

func getGroup(attrs *datadogV2.EventAttributes, eventTags []string) string {
	const prefix = "apo_group:"
	const plen = len(prefix)

	for _, tag := range eventTags {
		if strings.HasPrefix(tag, prefix) {
			return tag[plen:]
		}
	}

	tags := attrs.GetTags()
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			return tag[plen:]
		}
	}

	monitorTags := attrs.GetMonitor().Tags
	for _, tag := range monitorTags {
		if strings.HasPrefix(tag, prefix) {
			return tag[plen:]
		}
	}

	if len(attrs.GetService()) > 0 && attrs.GetService() != "undefined" {
		return string(clickhouse.APP_GROUP)
	}
	return ""
}
