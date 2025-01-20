// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "strconv"

func ConvertStatus(sourceType string, status string) string {
	switch sourceType {
	case PrometheusType:
		return status
	case ZabbixType:
		if code, err := strconv.Atoi(status); err == nil {
			switch code {
			case 0:
				return StatusResolved
			case 1:
				return StatusFiring
			}
		} else {
			switch status {
			case ZabbixStatusOK:
				return StatusResolved
			case ZabbixStatusProblem:
				return StatusFiring
			}
		}
	}

	return status
}
