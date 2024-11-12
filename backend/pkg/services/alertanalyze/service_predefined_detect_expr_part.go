package alertanalyze

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// DetectDefectsOptions 提供快速指标模版
func (s *service) DetectDefectsOptions() (*response.GetPredefinedDetectExprResponse, error) {
	quickMutationMetric, err := s.dbRepo.ListQuickMutationMetric()
	if err != nil {
		return nil, err
	}

	// 把数据库里的记录调整成和前端约定的方式
	var predefinedMetricsList []model.PredefinedMetricExpr
	for _, metric := range quickMutationMetric {
		predefinedMetricsList = append(predefinedMetricsList, model.PredefinedMetricExpr{
			DetectExprPart: model.DetectExprPart{
				PredefinedMetric: model.PredefinedMetric{
					Name:   metric.Name,
					Metric: metric.MetricTemplate,
				},
				Modifier: model.Modifier{
					RangeVector: metric.RangeVector,
					Scale:       metric.Scale,
					Quantile:    metric.Quantile,
					DayOnDay:    metric.DayOnDay,
					WeekOnWeek:  metric.WeekOnWeek,
				},
				Group:        metric.Group,
				CompareGroup: metric.CompareGroup,
				CustomMetric: metric.CustomMetric,
			},
			Describe: metric.Describe,
		})
	}

	return &response.GetPredefinedDetectExprResponse{
		PredefinedMetrics: predefinedMetricsList,
	}, nil
}
