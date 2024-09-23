package prometheus

import (
	"fmt"
	"strconv"
	"time"
)

const DefaultDepLatency int64 = -1

func (repo *promRepo) QueryAggMetricsWithFilter(pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, granularity Granularity, filterKVs ...string) ([]MetricResult, error) {
	if len(filterKVs)%2 != 0 {
		return nil, fmt.Errorf("size of filterKVs is not even: %d", len(filterKVs))
	}
	var filters []string
	for i := 0; i+1 < len(filterKVs); i += 2 {
		filters = append(filters, fmt.Sprintf("%s\"%s\"", filterKVs[i], filterKVs[i+1]))
	}
	vector := VecFromS2E(startTime, endTime)
	pql := pqlTemplate(vector, string(granularity), filters)
	return repo.QueryData(time.UnixMicro(endTime), pql)
}

// 计算指标日同比变化率
func DayOnDay(pqlTemplate AggPQLWithFilters) AggPQLWithFilters {
	return func(vector string, granularity string, filterKVs []string) string {
		nowPql := pqlTemplate(vector, granularity, filterKVs)

		return `(` + nowPql + `) / ((` + nowPql + `) offset 24h )`
	}
}

// 计算指标日同比变化率
func WeekOnWeek(pqlTemplate AggPQLWithFilters) AggPQLWithFilters {
	return func(vector string, granularity string, filterKVs []string) string {
		nowPql := pqlTemplate(vector, granularity, filterKVs)

		return `(` + nowPql + `) / ((` + nowPql + `) offset 7d )`
	}
}

func WithDefaultIFPolarisMetricExits(pqlTemplate AggPQLWithFilters, defaultValue int64) AggPQLWithFilters {
	return func(vector string, granularity string, filterKVs []string) string {
		pql := pqlTemplate(vector, granularity, filterKVs)
		checkPql := PQLIsPolarisMetricExitsWithFilters(vector, granularity, filterKVs)
		defaultV := strconv.FormatInt(defaultValue, 10)
		return `(` + pql + `) or ( ` + checkPql + ` * 0 + ` + defaultV + `)`
	}
}
