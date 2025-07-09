// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// PQLInstanceLog get the pql pod or vm of the instance-level log metric
func PQLInstanceLog(pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, granularity Granularity, podFilterKVs, vmFilterKVs []string) (string, error) {
	if len(podFilterKVs)%2 != 0 {
		return "", fmt.Errorf("size of podFilterKVs is not even: %d", len(podFilterKVs))
	}

	if len(vmFilterKVs)%2 != 0 {
		return "", fmt.Errorf("size of vmFilterKVs is not even: %d", len(vmFilterKVs))
	}
	var podFilters []string
	for i := 0; i+1 < len(podFilterKVs); i += 2 {
		podFilters = append(podFilters, fmt.Sprintf("%s\"%s\"", podFilterKVs[i], podFilterKVs[i+1]))
	}

	var vmFilters []string
	for i := 0; i+1 < len(vmFilterKVs); i += 2 {
		vmFilters = append(vmFilters, fmt.Sprintf("%s\"%s\"", vmFilterKVs[i], vmFilterKVs[i+1]))
	}

	vector := VecFromS2E(startTime, endTime)
	podPql := pqlTemplate(vector, string(granularity), podFilters)
	vmPql := pqlTemplate(vector, string(granularity), vmFilters)
	return `(` + podPql + `) or (` + vmPql + `)`, nil
}

func (repo *promRepo) QueryInstanceLogRangeData(ctx core.Context, pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, stepMicroS int64, granularity Granularity, podFilterKVs, vmFilterKVs []string) ([]MetricResult, error) {
	if len(podFilterKVs)%2 != 0 {
		return nil, fmt.Errorf("size of podFilterKVs is not even: %d", len(podFilterKVs))
	}

	if len(vmFilterKVs)%2 != 0 {
		return nil, fmt.Errorf("size of vmFilterKVs is not even: %d", len(vmFilterKVs))
	}
	var podFilters []string
	for i := 0; i+1 < len(podFilterKVs); i += 2 {
		podFilters = append(podFilters, fmt.Sprintf("%s\"%s\"", podFilterKVs[i], podFilterKVs[i+1]))
	}
	var vmFilters []string
	for i := 0; i+1 < len(vmFilterKVs); i += 2 {
		vmFilters = append(vmFilters, fmt.Sprintf("%s\"%s\"", vmFilterKVs[i], vmFilterKVs[i+1]))
	}

	step := time.Duration(stepMicroS) * time.Microsecond
	vector := VecFromDuration(step)
	podPql := pqlTemplate(vector, string(granularity), podFilters)
	vmPql := pqlTemplate(vector, string(granularity), vmFilters)
	pql := `(` + podPql + `) or (` + vmPql + `)`
	return repo.QueryRangeData(ctx, time.UnixMicro(startTime), time.UnixMicro(endTime), pql, step)
}
