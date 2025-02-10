package alertanalyze

import (
	"fmt"
	"log"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

func TestDetectExprPart_String(t *testing.T) {
	type fields struct {
		PredefinedMetric model.PredefinedMetric
		Modifier         model.Modifier
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "avg_latency",
			fields: fields{
				PredefinedMetric: model.PredefinedMetric{
					Name:   "latency",
					Metric: "sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_sum[%RangeVector%])) / sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_count[%RangeVector%])) / 1000000",
				},
				Modifier: model.Modifier{
					RangeVector: "1m",
					Scale:       "",
					Quantile:    "",
				},
			},
			want: "sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_sum[1m])) / sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_count[1m])) / 1000000",
		},
		{
			name: "histogram_latency",
			fields: fields{
				PredefinedMetric: model.PredefinedMetric{
					Name:   "latency",
					Metric: "histogram_quantile(%Quantile%, sum by (%BucketRange%,svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_bucket[%RangeVector%]))) ",
				},
				Modifier: model.Modifier{
					RangeVector: "1m",
					Scale:       "",
					Quantile:    "0.95",
				},
			},
			want: "histogram_quantile(0.95, sum by (%BucketRange%,svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_bucket[1m]))) ",
		},
	}

	fmt.Printf("%+v\n", map[string]string{
		"a": "a",
		"b": "b",
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &model.DetectExprPart{
				PredefinedMetric: tt.fields.PredefinedMetric,
				Modifier:         tt.fields.Modifier,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("DetectExprPart.String() = %v, want %v", got, tt.want)
			} else {
				log.Println(got)
			}
		})
	}
}
