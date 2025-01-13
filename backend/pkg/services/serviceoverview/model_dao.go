// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// RES_MAX_VALUE returns the maximum value of the front end. If the value is equal to the same period last year, the maximum value is indicated.
const RES_MAX_VALUE = 9999999

type ServiceDetail struct {
	ServiceName          string
	EndpointCount        int
	ServiceSize          int
	Endpoints            []*prom.EndpointMetrics
	Instances            []Instance
	LogData              []prom.Points // Data of log alarm times for 30min
	InfrastructureStatus string
	NetStatus            string
	K8sStatus            string
	Timestamp            int64
}

const (
	POD = iota
	CONTAINER
	VM
)

type Instance struct {
	InstanceName           string // instance name
	ContentKey             string // URL
	ConvertName            string
	SvcName                string // Name of the service to which the url belongs
	Pod                    string
	Namespace              string
	NodeName               string
	NodeIP                 string
	Count                  int
	InstanceType           int
	ContainerId            string
	Pid                    string
	IsLatencyDODExceeded   bool
	IsErrorRateDODExceeded bool
	IsTPSDODExceeded       bool
	IsLatencyWOWExceeded   bool
	IsErrorRateWOWExceeded bool
	IsTPSWOWExceeded       bool
	AvgLatency             *float64      // Average delay time within 30min
	LatencyDayOverDay      *float64      // Delay Day-over-Day Growth Rate
	LatencyWeekOverWeek    *float64      // Delay Week YoY
	LatencyData            []prom.Points // Data with 30min delay

	AvgErrorRate          *float64      // Average error rate over 30min
	ErrorRateDayOverDay   *float64      // Error Rate Day-over-Day Growth Rate
	ErrorRateWeekOverWeek *float64      // Error Rate Week-on-Week
	ErrorRateData         []prom.Points // error rate data for 30min

	AvgTPS          *float64      // Average TPS over 30min
	TPSDayOverDay   *float64      // TPS Day-over-Day Growth Rate
	TPSWeekOverWeek *float64      // TPS Week YoY
	TPSData         []prom.Points // TPS 30min

	AvgLog          *float64      // Number of log alarms in 30min
	LogDayOverDay   *float64      // Number of log alarms per day
	LogWeekOverWeek *float64      // Number of log alarms week-on-week
	LogData         []prom.Points // Data of log alarm times for 30min

	// used to calculate YoY under service
	LogNow       *float64 // Number of log errors during the query period
	LogYesterday *float64 // Query the number of log errors from yesterday during the time period
	LogLastWeek  *float64 // Number of log errors in the last week during the query period
}
