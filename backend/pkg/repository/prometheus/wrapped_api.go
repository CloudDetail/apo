package prometheus

import (
	"context"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

type WrappedApi struct {
	v1.API
	logger *zap.Logger
}

func (api *WrappedApi) Query(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		api.logger.Sugar().Debugf("Promethues Query: {query=%s, ts=%d}, cost: %d ms",
			query, ts.UnixNano(), endTime.UnixMilli()-startTime.UnixMilli())
	}()
	return api.API.Query(ctx, query, ts, opts...)
}

func (api *WrappedApi) QueryRange(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		api.logger.Sugar().Debugf("Promethues QueryRange: {query=%s, from=%d, to=%d, step=%d}, cost: %d ms",
			query, r.Start.UnixNano(), r.End.UnixNano(), int64(r.Step), endTime.UnixMilli()-startTime.UnixMilli())
	}()
	return api.API.QueryRange(ctx, query, r, opts...)
}
