// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"strings"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func TestBuildRYGServiceInstanceFilterKeepsDataGroupScope(t *testing.T) {
	groupFilter := prometheus.NewFilter().
		Equal("cluster_id", "cluster-a").
		Equal("namespace", "orders")
	baseFilter := prometheus.NewFilter().
		RegexMatch(prometheus.ServiceNameKey, "order.*")

	filter := buildRYGServiceInstanceFilter(groupFilter, baseFilter, "order-api")
	got := filter.String()

	for _, want := range []string{
		`cluster_id="cluster-a"`,
		`namespace="orders"`,
		`svc_name=~"order.*"`,
		`svc_name="order-api"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected scoped service instance filter to contain %q, got %s", want, got)
		}
	}
}
