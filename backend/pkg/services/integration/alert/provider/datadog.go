// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
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
	SupportWebhookInstall: false, // TODO support webhook installation in datadog
}

type DatadogProvider struct {
	api *datadogV2.EventsApi

	ctx        context.Context
	sourceFrom alert.SourceFrom
}

func NewDatadogProvider(sourceFrom alert.SourceFrom, params alert.AlertSourceParams) Provider {
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
		api:        datadogV2.NewEventsApi(client),
		ctx:        ctx,
		sourceFrom: sourceFrom,
	}
}

func (f *DatadogProvider) SetupWebhook(ctx core.Context, webhookURL string, params alert.AlertSourceParams) error {
	// TODO support webhook installation in datadog
	panic("unimplemented")
}

func (f *DatadogProvider) PullAlerts(args GetAlertParams) ([]alert.AlertEvent, error) {
	start := args.From
	end := args.To

	resChan, cancel := f.api.ListEventsWithPagination(f.ctx, *datadogV2.NewListEventsOptionalParameters().
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

		event := f.parseEventItem(item, receivedTime)
		events = append(events, event)
	}

	return events, err
}

func (f *DatadogProvider) parseEventItem(item datadog.PaginationResult[datadogV2.EventResponse], receivedTime time.Time) alert.AlertEvent {
	attrs := item.Item.Attributes
	nestedAttrs := item.Item.Attributes.Attributes

	monitor := nestedAttrs.GetMonitor()
	severity := f.getSeverityFromMonitor(monitor)

	createTime := f.calculateCreateTime(nestedAttrs)
	status, endTime := f.getStatusAndEndTimeFromMonitor(monitor, nestedAttrs)

	tags := buildDDTags(nestedAttrs, attrs.GetTags())
	group := getGroup(nestedAttrs, attrs.GetTags())

	return alert.AlertEvent{
		Alert: alert.Alert{
			Source:     f.sourceFrom.SourceName,
			SourceID:   f.sourceFrom.SourceID,
			AlertID:    nestedAttrs.GetAggregationKey(),
			Group:      group,
			Name:       nestedAttrs.GetTitle(),
			EnrichTags: make(map[string]string),
			Tags:       tags,
		},
		EventID:      item.Item.GetId(),
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

	attrTags := make(map[string]any)
	attrTags["title"] = attrs.GetTitle()
	attrTags["service"] = attrs.GetService()
	attrTags["hostname"] = attrs.GetHostname()
	for _, tagStr := range attrs.GetTags() {
		tag := strings.Split(tagStr, ":")
		if len(tag) == 2 {
			attrTags[tag[0]] = tag[1]
		} else {
			attrTags[tagStr] = ""
		}
	}

	monitor := attrs.GetMonitor()

	monitorTags := make(map[string]any)
	monitorTags["id"] = attrs.GetMonitorId()
	monitorTags["name"] = monitor.GetName()
	monitorTags["query"] = monitor.GetQuery()
	for _, tagStr := range monitor.GetTags() {
		tag := strings.Split(tagStr, ":")
		if len(tag) == 2 {
			monitorTags[tag[0]] = tag[1]
		} else {
			monitorTags[tagStr] = ""
		}
	}

	tags["attr"] = attrTags
	attrTags["monitor"] = monitorTags

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
