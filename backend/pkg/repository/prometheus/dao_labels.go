// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func (repo *promRepo) LabelValues(ctx core.Context, expr string, label string, startTime, endTime int64) (model.LabelValues, error) {
	labelValues, _, err := repo.api.LabelValues(
		ctx.GetContext(),
		label,
		[]string{expr},
		time.UnixMicro(startTime),
		time.UnixMicro(endTime),
	)
	if err != nil {
		return nil, err
	}
	return labelValues, nil
}

func (repo *promRepo) QueryResult(ctx core.Context, expr string, regex string, startTime, endTime int64) ([]string, error) {
	value, _, err := repo.api.QueryRange(ctx.GetContext(), expr, v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(endTime-startTime) * time.Microsecond,
	})
	if err != nil {
		return nil, err
	}

	vector, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Vector", value)
	}

	var valueRegex *regexp.Regexp
	if len(regex) > 0 {
		pattern := cleanRegex(regex, "value")
		var err error
		valueRegex, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
	}

	var values []string
	for _, sample := range vector {
		res := sample.Metric.String()

		if valueRegex == nil {
			values = append(values, res)
			continue
		}

		subStrs := valueRegex.FindStringSubmatch(res)
		if len(subStrs) > 1 {
			values = append(values, subStrs[1])
		}
	}

	return values, nil
}

func cleanRegex(originalRegex string, keepGroup string) string {
	regex := strings.Trim(originalRegex, "/")
	if idx := strings.Index(regex, "/"); idx != -1 {
		regex = regex[:idx] // remove '/g'
	}

	parts := strings.Split(regex, "|")
	var cleanedParts []string

	for _, part := range parts {
		reNamed := regexp.MustCompile(`\(\?\<([a-zA-Z]+)\>([^\)]+)\)`)
		matches := reNamed.FindAllStringSubmatch(part, -1)

		if len(matches) > 0 {
			for _, match := range matches {
				fullMatch := match[0]
				name := match[1]
				content := match[2]

				if name == keepGroup {
					cleaned := fmt.Sprintf("(%s)", content)
					cleanedPart := strings.Replace(part, fullMatch, cleaned, 1)
					cleanedParts = append(cleanedParts, cleanedPart)
				}
			}
		} else {
			cleanedParts = append(cleanedParts, part)
		}
	}

	if len(cleanedParts) == 0 {
		return ""
	}
	return cleanedParts[0]
}
