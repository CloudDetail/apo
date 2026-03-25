// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	pmodel "github.com/prometheus/common/model"
	prometheus_model "github.com/prometheus/common/model"
)

type UnionP9xBuilder struct {
	value           string
	tableName       string
	labels          []string
	duration        time.Duration
	count           int
	conditions      []*P9xCondition
	extraConditions []string
}

var (
	MAX_CONDITIONS_PER_GROUP = 30
)

func (repo *promRepo) QueryRangeWithP9xBuilder(ctx core.Context, builder *UnionP9xBuilder, tRange v1.Range) (pmodel.Value, v1.Warnings, error) {
	if builder.count <= MAX_CONDITIONS_PER_GROUP {
		pql := builder.ToString()
		return repo.GetApi().QueryRange(ctx.GetContext(), pql, tRange)
	}

	var res []*pmodel.SampleStream
	var warnings v1.Warnings
	var errs []error

	for i := 0; i < builder.count; i += MAX_CONDITIONS_PER_GROUP {
		end := i + MAX_CONDITIONS_PER_GROUP
		if end > builder.count {
			end = builder.count
		}
		pql := builder.buildQueryWithCondRange(i, end)
		subRes, subWarnings, subErr := repo.GetApi().QueryRange(ctx.GetContext(), pql, tRange)
		if subErr != nil {
			errs = append(errs, subErr)
			continue
		}

		if subWarnings != nil {
			warnings = append(warnings, subWarnings...)
		}

		if subRes == nil {
			continue
		}

		subMatrix, ok := subRes.(prometheus_model.Matrix)
		if !ok {
			return nil, nil,
				core.Error(
					code.GetDescendantMetricsError,
					fmt.Sprintf("invalid query result, expected matrix but got %T, subQuery: %s", subRes, pql),
				)
		}
		if len(subMatrix) <= 0 {
			continue
		}
		res = append(res, subMatrix...)
	}

	return prometheus_model.Matrix(res), warnings, errors.Join(errs...)
}

func NewUnionP9xBuilder(value string, tableName string, labels []string, duration time.Duration) *UnionP9xBuilder {
	return &UnionP9xBuilder{
		value:           value,
		tableName:       tableName,
		labels:          labels,
		duration:        duration,
		count:           0,
		conditions:      make([]*P9xCondition, 0), // to ensure order
		extraConditions: make([]string, 0),
	}
}

func (p9x *UnionP9xBuilder) AddCondition(key string, values []string) error {
	if len(values) == 0 {
		return nil
	}

	if p9x.count == 0 {
		p9x.count = len(values)
	} else if p9x.count != len(values) {
		return fmt.Errorf("fail to addCondition. Expect %s.count = %d, but got %d", key, p9x.count, len(values))
	}
	p9x.conditions = append(p9x.conditions, &P9xCondition{
		Key:    key,
		Values: values,
	})
	return nil
}

func (p9x *UnionP9xBuilder) AddExtraCondition(condition string) {
	if condition != "" {
		p9x.extraConditions = append(p9x.extraConditions, condition)
	}
}

func (p9x *UnionP9xBuilder) ToString() string {
	var builder strings.Builder
	if p9x.count > 1 {
		builder.WriteString("union(")
	}

	for i := 0; i < p9x.count; i++ {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString("histogram_quantile(")
		builder.WriteString(p9x.value)
		builder.WriteString(", sum by(")
		for j, label := range p9x.labels {
			if j > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(label)
		}
		builder.WriteString(") (increase(")
		builder.WriteString(p9x.tableName)
		builder.WriteString("{")
		for k, condition := range p9x.conditions {
			if k > 0 {
				builder.WriteString(", ")
			}

			builder.WriteString(condition.Key)
			builder.WriteString("='")
			builder.WriteString(condition.Values[i])
			builder.WriteString("'")
		}
		for n, extraCondition := range p9x.extraConditions {
			if len(p9x.conditions) > 0 || n > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(extraCondition)
		}
		builder.WriteString("}[")
		builder.WriteString(getDurationFromStep(p9x.duration))
		builder.WriteString("])))")
	}
	if p9x.count > 1 {
		builder.WriteString(")")
	}

	return builder.String()
}

func (p9x *UnionP9xBuilder) buildQueryWithCondRange(start, end int) string {
	var builder strings.Builder

	lens := end - start
	if lens > 1 {
		builder.WriteString("union(")
	}

	for i := 0; i < lens; i++ {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString("histogram_quantile(")
		builder.WriteString(p9x.value)
		builder.WriteString(", sum by(")
		for j, label := range p9x.labels {
			if j > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(label)
		}
		builder.WriteString(") (increase(")
		builder.WriteString(p9x.tableName)
		builder.WriteString("{")
		for k, condition := range p9x.conditions {
			if k > 0 {
				builder.WriteString(", ")
			}

			builder.WriteString(condition.Key)
			builder.WriteString("='")
			builder.WriteString(condition.Values[start+i])
			builder.WriteString("'")
		}
		for n, extraCondition := range p9x.extraConditions {
			if len(p9x.conditions) > 0 || n > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(extraCondition)
		}
		builder.WriteString("}[")
		builder.WriteString(getDurationFromStep(p9x.duration))
		builder.WriteString("])))")
	}
	if lens > 1 {
		builder.WriteString(")")
	}

	return builder.String()
}

type P9xCondition struct {
	Key    string
	Values []string
}

func getDurationFromStep(step time.Duration) string {
	var stepNS = step.Nanoseconds()
	if stepNS > TIME_DAY && (stepNS%TIME_DAY == 0) {
		return strconv.FormatInt(stepNS/TIME_DAY, 10) + "d"
	}
	if stepNS > TIME_HOUR && (stepNS%TIME_HOUR == 0) {
		return strconv.FormatInt(stepNS/TIME_HOUR, 10) + "h"
	}
	if stepNS > TIME_MINUTE && (stepNS%TIME_MINUTE == 0) {
		return strconv.FormatInt(stepNS/TIME_MINUTE, 10) + "m"
	}
	if stepNS > TIME_SECOND && (stepNS%TIME_SECOND == 0) {
		return strconv.FormatInt(stepNS/TIME_SECOND, 10) + "s"
	}

	// Default time
	return "1m"
}
