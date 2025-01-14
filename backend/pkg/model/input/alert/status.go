// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "strconv"

const (
	StatusFiring   = "firing"
	StatusResolved = "resolved"
)

func ConvertStatus(sourceType string, status string) string {
	switch sourceType {
	case "prometheus":
		return status
	case "zabbix":
		if code, err := strconv.Atoi(status); err == nil {
			switch code {
			case 0:
				return StatusResolved
			case 1:
				return StatusFiring
			}
		} else {
			switch status {
			case "OK":
				return StatusResolved
			case "PROBLEM":
				return StatusFiring
			}
		}
	}

	return status
}
