// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

/**
错误率，延时，TPS指标异常的排序权重
*/

const ErrorCount = 1
const TPSCount = 1
const LatencyCount = 1

/*
*
排序逻辑
*/

type SortType int

const (
	// DODThreshold 日同比阈值排序
	DODThreshold SortType = iota + 1
	// MUTATIONSORT 突变排序
	MUTATIONSORT
)
