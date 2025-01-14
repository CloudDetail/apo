// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "strconv"

const (
	SeverityCriticalLevel = "critical"
	SeverityErrorLevel    = "error"
	SeverityWarnLevel     = "warn"
	SeverityInfoLevel     = "info"
	SeverityUnknownLevel  = "unknown"
)

func ConvertSeverity(sourceType string, severity string) string {
	switch sourceType {
	case "prometheus":
		return severity
	case "zabbix":
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
			case "Disaster":
				return SeverityCriticalLevel
			case "High":
				return SeverityErrorLevel
			case "Average", "Warning":
				return SeverityWarnLevel
			case "Information":
				return SeverityInfoLevel
			case "Not classified":
				return SeverityUnknownLevel
			}
		}
	default:
		return severity
	}
	return severity
}
