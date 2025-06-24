package model

import (
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

type ReportOverView struct {
	TimeStamp int64             `json:"timestamp"`
	Detail    string            `json:"detail"`
	Reason    string            `json:"reason"`
	Tags      map[string]string `json:"tags"`
}

type ReportRootCause struct {
	Cause string           `json:"cause"`
	Other []map[string]any `json:"other"`
}
