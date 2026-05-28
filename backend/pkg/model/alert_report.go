package model

import (
	"encoding/json"
	"time"
)

type AlertEventReport struct {
	ReportType        string            `ch:"report_type"`
	OverviewTimestamp time.Time         `ch:"overview_timestamp"`
	OverviewDetail    string            `ch:"overview_detail"`
	OverviewReason    string            `ch:"overview_reason"`
	OverviewTags      map[string]string `ch:"overview_tags"`
	Tags              map[string]string `ch:"tags"`
	Topology          string            `ch:"topology"`
	RootCause         string            `ch:"rootcause"`
	RootCauseOther    []string          `ch:"rootcause_other"`
	Suggest           []string          `ch:"suggest"`
	Data              []string          `ch:"data"`
}

const (
	ReportOverviewHeadlineTagKey = "__overview_headline"
	ReportViewTagKey             = "__report_view"
)

type ReportOverView struct {
	TimeStamp int64             `json:"timestamp"`
	Headline  string            `json:"headline,omitempty"`
	Detail    string            `json:"detail"`
	Reason    string            `json:"reason"`
	Tags      map[string]string `json:"tags"`
}

type ReportRootCause struct {
	Cause string           `json:"cause"`
	Other []map[string]any `json:"other"`
}

func ReportOverviewHeadlineFromTags(tags map[string]string) string {
	if tags == nil {
		return ""
	}
	return tags[ReportOverviewHeadlineTagKey]
}

func ReportOverviewTagsForStorage(tags map[string]string, headline string, reportView ...map[string]any) map[string]string {
	result := make(map[string]string, len(tags)+1)
	for key, value := range tags {
		result[key] = value
	}
	if headline != "" {
		result[ReportOverviewHeadlineTagKey] = headline
	}
	if len(reportView) > 0 && len(reportView[0]) > 0 {
		if raw, err := json.Marshal(reportView[0]); err == nil {
			result[ReportViewTagKey] = string(raw)
		}
	}
	return result
}

func ReportOverviewPublicTags(tags map[string]string) map[string]string {
	result := make(map[string]string, len(tags))
	for key, value := range tags {
		if key == ReportOverviewHeadlineTagKey || key == ReportViewTagKey {
			continue
		}
		result[key] = value
	}
	return result
}

func ReportViewFromTags(tags map[string]string) map[string]any {
	if tags == nil || tags[ReportViewTagKey] == "" {
		return nil
	}
	var reportView map[string]any
	if err := json.Unmarshal([]byte(tags[ReportViewTagKey]), &reportView); err != nil {
		return nil
	}
	return reportView
}
