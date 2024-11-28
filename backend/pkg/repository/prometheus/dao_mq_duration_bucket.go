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
	TEMPLATE_FILTER_MQ_ROLE   = `role!="%s"`
	TEMPLATE_FILTER_MQ_SVC    = `address=~"%s"`
	TEMPLATE_FILTER_MQ_URL    = `name=~"%s"`
	TEMPLATE_FILTER_MQ_SYSTEM = `system=~"%s"`
	TEMPLATE_HISTO_P90_MQ     = `histogram_quantile(0.9, sum by (%s,address,name) (increase(kindling_mq_duration_nanoseconds_bucket{%s}[%s])))`
)

// 基于服务列表、URL列表和时段、步长，查询P90曲线
func (repo *promRepo) QueryMqRangePercentile(startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error) {
	svcs, endpoints, systems := nodes.GetLabels(model.GROUP_MQ)
	if len(svcs) == 0 {
		return nil, nil
	}
	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}
	filters := []string{}
	filters = append(filters, fmt.Sprintf(TEMPLATE_FILTER_MQ_ROLE, "consumer"))
	filters = append(filters, fmt.Sprintf(TEMPLATE_FILTER_MQ_SVC, strings.Join(svcs, "|")))
	filters = append(filters, fmt.Sprintf(TEMPLATE_FILTER_MQ_URL, RegexMultipleValue(endpoints...)))
	filters = append(filters, fmt.Sprintf(TEMPLATE_FILTER_MQ_SYSTEM, strings.Join(systems, "|")))

	query := fmt.Sprintf(TEMPLATE_HISTO_P90_MQ,
		repo.GetRange(),
		strings.Join(filters, ","),
		getDurationFromStep(tRange.Step),
	)
	res, _, err := repo.GetApi().QueryRange(context.Background(), query, tRange)
	if err != nil {
		return nil, err
	}
	return getDescendantMetrics("address", "name", tRange, res), nil
}
