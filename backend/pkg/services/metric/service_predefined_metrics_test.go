package metric

import "testing"

func TestNormalizeMetricParamsUsesInstanceNameAsNodeAlias(t *testing.T) {
	params := normalizeMetricParams(map[string]string{
		"instance_name": "CloudInsightAgent-POC-2",
		"pid":           ".*",
	})

	if params["node"] != "CloudInsightAgent-POC-2" {
		t.Fatalf("expected node alias from instance_name, got %q", params["node"])
	}
	if params["pid"] != ".*" {
		t.Fatalf("expected pid to be preserved, got %q", params["pid"])
	}
}

func TestNormalizeMetricParamsPreservesExplicitNode(t *testing.T) {
	params := normalizeMetricParams(map[string]string{
		"node":          "explicit-node",
		"instance_name": "ignored-node",
	})

	if params["node"] != "explicit-node" {
		t.Fatalf("expected explicit node to be preserved, got %q", params["node"])
	}
}

func TestNormalizeMetricParamsHandlesNil(t *testing.T) {
	params := normalizeMetricParams(nil)

	if params == nil {
		t.Fatal("expected params map to be initialized")
	}
}
