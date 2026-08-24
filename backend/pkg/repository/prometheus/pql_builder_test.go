// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"strings"
	"testing"
)

func TestPQLAvgLogErrorCountCombineServiceInfoDoesNotDuplicateByEndpoint(t *testing.T) {
	query := PQLAvgLogErrorCountCombineServiceInfoWithPQLFilter(
		"15m",
		string(SVCGranularity),
		NewFilter().RegexMatch(ServiceNameKey, "order-api|order-worker").RegexMatch(ContentKeyKey, "/api/.*"),
		"",
	)

	for _, unexpected := range []string{
		"group_left (svc_name,content_key)",
		"sum by (svc_name, content_key)",
		"sum by (svc_name,content_key)",
	} {
		if strings.Contains(query, unexpected) {
			t.Fatalf("expected service log query not to duplicate logs by endpoint dimension %q, got %s", unexpected, query)
		}
	}
	for _, expected := range []string{
		"group_left (svc_name)",
		"on (container_id)",
		"on (node_name,pid)",
		`svc_name=~"order-api|order-worker"`,
		`content_key=~"/api/.*"`,
	} {
		if !strings.Contains(query, expected) {
			t.Fatalf("expected service log query to contain %q, got %s", expected, query)
		}
	}
}
