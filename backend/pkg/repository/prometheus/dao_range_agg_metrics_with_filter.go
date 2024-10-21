package prometheus

import (
	"fmt"
	"time"
)

func (repo *promRepo) QueryRangeAggMetricsWithFilter(pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, stepMicroS int64, granularity Granularity, filterKVs ...string) ([]MetricResult, error) {
	if len(filterKVs)%2 != 0 {
		return nil, fmt.Errorf("size of filterKVs is not even: %d", len(filterKVs))
	}
	var filters []string
	for i := 0; i+1 < len(filterKVs); i += 2 {
		filters = append(filters, fmt.Sprintf("%s\"%s\"", filterKVs[i], filterKVs[i+1]))
	}

	step := time.Duration(stepMicroS) * time.Microsecond
	vector := VecFromDuration(step)
	pql := pqlTemplate(vector, string(granularity), filters)
	return repo.QueryRangeData(
		time.UnixMicro(startTime), time.UnixMicro(endTime),
		pql,
		step)
}

func (repo *promRepo) QueryInstanceLogRangeData(pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, stepMicroS int64, granularity Granularity, podFilterKVs, vmFilterKVs []string) ([]MetricResult, error) {
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
	return repo.QueryRangeData(time.UnixMicro(startTime), time.UnixMicro(endTime), pql, step)
}
