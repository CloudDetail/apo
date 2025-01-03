// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetSQLMetricsRequest struct {
	StartTime int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Service   string `form:"service" binding:"required"`                   // 查询服务名
	Step      int64  `form:"step" binding:"min=1000000"`                   // 查询步长(us)

	SortBy     string `form:"sortBy"` // 排序方式,(latency,errorRate,tps) 默认按照latency排序
	*PageParam        // 分页参数
}
