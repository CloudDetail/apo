// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "strconv"

func ConvertSeverity(sourceType string, severity string) string {
	switch sourceType {
	case PrometheusType:
		return severity
	case ZabbixType:
		// Try to determine the data type
		if level, err := strconv.Atoi(severity); err == nil {
			switch level {
			case 0:
				return SeverityUnknownLevel
			case 1:
				return SeverityInfoLevel
			case 2, 3:
				return SeverityWarnLevel
			case 4:
				return SeverityErrorLevel
			case 5:
				return SeverityCriticalLevel
			}
		} else {
			switch severity {
			case ZabbixSeverityDisaster:
				return SeverityCriticalLevel
			case ZabbixSeverityHigh:
				return SeverityErrorLevel
			case ZabbixSeverityAverage, ZabbixSeverityWarning:
				return SeverityWarnLevel
			case ZabbixSeverityInfo:
				return SeverityInfoLevel
			case ZabbixSeverityUnknown:
				return SeverityUnknownLevel
			}
		}
	default:
		return severity
	}
	return severity
}
