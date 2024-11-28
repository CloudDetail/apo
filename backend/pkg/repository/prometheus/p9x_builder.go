package prometheus

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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

func NewUnionP9xBuilder(value string, tableName string, labels []string, duration time.Duration) *UnionP9xBuilder {
	return &UnionP9xBuilder{
		value:           value,
		tableName:       tableName,
		labels:          labels,
		duration:        duration,
		count:           0,
		conditions:      make([]*P9xCondition, 0), // 保证有序
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

	// 默认时间
	return "1m"
}
