package prometheus

import (
	"context"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// 基于服务列表、URL列表和时段、步长，查询P90曲线
func (repo *promRepo) QueryDbRangePercentile(startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error) {
	svcs, endpoints, systems := nodes.GetLabels(model.GROUP_DB)
	if len(svcs) == 0 {
		return nil, nil
	}

	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}

	query := getDbP9xSql(repo.promRange, tRange.Step, svcs, endpoints, systems)
	res, _, err := repo.GetApi().QueryRange(context.Background(), query, tRange)
	if err != nil {
		return nil, err
	}
	return getDescendantMetrics("db_url", "name", tRange, res), nil
}

func getDbP9xSql(promRange string, step time.Duration, svcs []string, endpoints []string, systems []string) string {
	builder := NewUnionP9xBuilder(
		"0.9",
		"kindling_db_duration_nanoseconds_bucket",
		[]string{promRange, "db_url", "name"},
		step,
	)
	builder.AddCondition("db_url", svcs)
	builder.AddCondition("name", endpoints)
	builder.AddCondition("db_system", systems)
	return builder.ToString()
}
