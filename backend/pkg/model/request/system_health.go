// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetSystemHealthReportRequest struct {
	GroupID      int64  `form:"groupId"`
	RelativeTime string `form:"relativeTime"`
	StartTime    int64  `form:"startTime"`
	EndTime      int64  `form:"endTime"`
}
