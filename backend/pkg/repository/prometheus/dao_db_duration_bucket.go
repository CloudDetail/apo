package prometheus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

const (
	TEMPLATE_FILTER_DB_SVC    = `db_url=~"%s"`
	TEMPLATE_FILTER_DB_URL    = `name=~"%s"`
	TEMPLATE_FILTER_DB_SYSTEM = `db_system=~"%s"`
	TEMPLATE_HISTO_P90_DB     = `histogram_quantile(0.9, sum by (%s,db_url,name) (increase(kindling_db_duration_nanoseconds_bucket{%s}[%s])))`
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

	filters := []string{}
	filters = append(filters, fmt.Sprintf(TEMPLATE_FILTER_DB_SVC, strings.Join(svcs, "|")))
	filters = append(filters, fmt.Sprintf(TEMPLATE_FILTER_DB_URL, RegexMultipleValue(endpoints...)))
	filters = append(filters, fmt.Sprintf(TEMPLATE_FILTER_DB_SYSTEM, strings.Join(systems, "|")))

	query := fmt.Sprintf(TEMPLATE_HISTO_P90_DB,
		repo.GetRange(),
		strings.Join(filters, ","),
		getDurationFromStep(tRange.Step),
	)
	res, _, err := repo.GetApi().QueryRange(context.Background(), query, tRange)
	if err != nil {
		return nil, err
	}
	return getDescendantMetrics("db_url", "name", tRange, res), nil
}
