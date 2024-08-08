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
