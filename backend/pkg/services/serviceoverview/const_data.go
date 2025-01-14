// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

// Error rate, delay, TPS sort weight
const ErrorCount = 1
const TPSCount = 1
const LatencyCount = 1

type SortType int

const (
	// Sort by Day-over-Day Growth Rate Threshold
	DODThreshold SortType = iota + 1
	// Sort by mutation
	MUTATIONSORT
)
