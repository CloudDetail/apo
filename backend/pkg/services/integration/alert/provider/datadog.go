// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

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
			"site": params["site"].(string),
		},
	)

	ctx = context.WithValue(ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {Key: params["apiKey"].(string)},
			"appKeyAuth": {Key: params["appKey"].(string)},
		},
	)

	return &DatadogProvider{
		api:        datadogV2.NewEventsApi(client),
		ctx:        ctx,
		sourceFrom: sourceFrom,
	}
}

func (f *DatadogProvider) GetAlerts(args map[string]any) ([]alert.AlertEvent, error) {
	lastPullTs := args["from"].(time.Time)
	now := args["to"].(time.Time)

	resChan, cancel := f.api.ListEventsWithPagination(f.ctx, *datadogV2.NewListEventsOptionalParameters().
		WithPageLimit(1000).
		WithSort(datadogV2.EVENTSSORT_TIMESTAMP_ASCENDING).
		WithFilterFrom(strconv.FormatInt(lastPullTs.UnixMilli(), 10)).
		WithFilterTo(strconv.FormatInt(now.UnixMilli(), 10)).
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

		attrs := item.Item.Attributes
		nestedAttrs := item.Item.Attributes.Attributes

		monitor := nestedAttrs.GetMonitor()
		var priority = alert.SeverityUnknownLevel
		if priorityLevel, find := monitor.AdditionalProperties["priority"]; find {
			priority = getDDPriority(priorityLevel.(float64))
		}

		var status = alert.StatusFiring
		var createTime, endTime time.Time
		if transition, find := monitor.AdditionalProperties["transition"]; find {
			status = getDDStatus(transition)

			if status == alert.StatusResolved {
				endTime = time.UnixMilli(nestedAttrs.GetTimestamp())
			}
		}
		if nestedAttrs.GetDuration() > 0 {
			createTime = time.UnixMilli(nestedAttrs.GetTimestamp() - nestedAttrs.GetDuration()/1e6)
		} else {
			createTime = time.UnixMilli(nestedAttrs.GetTimestamp())
		}

		tags := buildDDTags(nestedAttrs, attrs.GetTags())
		group := getGroup(nestedAttrs, attrs.GetTags())

		events = append(events, alert.AlertEvent{
			Alert: alert.Alert{
				Source:     f.sourceFrom.SourceName,
				SourceID:   f.sourceFrom.SourceID,
				AlertID:    nestedAttrs.GetAggregationKey(),
				Group:      group,
				Name:       nestedAttrs.GetTitle(),
				EnrichTags: make(map[string]string),
				Tags:       tags,
			},
			// ID:           uuid.UUID{},
			EventID:      item.Item.GetId(),
			Detail:       buildDDDetail(attrs.GetMessage(), tags),
			CreateTime:   createTime,
			UpdateTime:   time.UnixMilli(nestedAttrs.GetTimestamp()),
			EndTime:      endTime,
			ReceivedTime: receivedTime,
			Severity:     priority,
			Status:       status,
		})
	}

	return events, err
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

func getDDPriority(priority float64) string {
	switch priority {
	case 1:
		return alert.SeverityCriticalLevel
	case 2:
		return alert.SeverityErrorLevel
	case 3:
		return alert.SeverityWarnLevel
	case 4, 5:
		return alert.SeverityInfoLevel
	default:
		return alert.SeverityUnknownLevel
	}
}

func getDDStatus(transition any) string {
	if transition == nil {
		return alert.StatusFiring
	}

	transitionMap := transition.(map[string]any)
	if status, find := transitionMap["transition_type"]; find {
		if status == "alert recovery" {
			return alert.StatusResolved
		}
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
	for _, tag := range eventTags {
		if strings.HasPrefix(tag, "apo_group:") {
			return tag[6:]
		}
	}

	tags := attrs.GetTags()

	for _, tag := range tags {
		if strings.HasPrefix(tag, "apo_group:") {
			return tag[6:]
		}
	}

	monitorTags := attrs.GetTags()
	if len(monitorTags) > 0 {
		for _, tag := range monitorTags {
			if strings.HasPrefix(tag, "apo_group:") {
				return tag[6:]
			}
		}
	}

	if len(attrs.GetService()) > 0 && attrs.GetService() != "undefined" {
		return string(clickhouse.APP_GROUP)
	}
	return ""
}
