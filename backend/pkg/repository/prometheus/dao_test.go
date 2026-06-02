// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/util"
	prommodel "github.com/prometheus/common/model"
)

func init() {
	os.Setenv("APO_CONFIG", "../../../config/apo.yml")
}

func TestRepo(t *testing.T) {
	if os.Getenv("APO_PROMETHEUS_INTEGRATION") != "1" {
		t.Skip("set APO_PROMETHEUS_INTEGRATION=1 to run prometheus integration tests")
	}

	cfg := config.Get().Promethues
	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	repo, err := New(zapLog, cfg.Address, cfg.Storage)
	if err != nil {
		t.Fatalf("Error to connect prometheus: %v", err)
	}

	testGetActiveInstanceList(t, repo)
}

func TestLabelsExtractIncludesContainerLabel(t *testing.T) {
	labels := Labels{}
	labels.Extract(prommodel.Metric{
		prommodel.LabelName("pod"):       prommodel.LabelValue("apisix-0"),
		prommodel.LabelName("container"): prommodel.LabelValue("apisix"),
	})

	if labels.POD != "apisix-0" {
		t.Fatalf("expected pod label to be extracted, got %q", labels.POD)
	}
	if labels.Container != "apisix" {
		t.Fatalf("expected container label to be extracted, got %q", labels.Container)
	}
}

func TestLabelsExtractPreservesArbitraryLabelsForLegendTemplates(t *testing.T) {
	labels := Labels{}
	labels.Extract(prommodel.Metric{
		prommodel.LabelName("src_pod"):      prommodel.LabelValue("frontend-0"),
		prommodel.LabelName("dst_ip"):       prommodel.LabelValue("10.0.0.9"),
		prommodel.LabelName("custom_label"): prommodel.LabelValue("custom-value"),
	})

	payload, err := json.Marshal(labels)
	if err != nil {
		t.Fatalf("marshal labels: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal labels: %v", err)
	}

	for key, want := range map[string]string{
		"src_pod":      "frontend-0",
		"dst_ip":       "10.0.0.9",
		"custom_label": "custom-value",
	} {
		if got[key] != want {
			t.Fatalf("expected label %q to be preserved for legend templates, got %q", key, got[key])
		}
	}
}

func testGetActiveInstanceList(t *testing.T, repo Repo) {
	instances, err := repo.GetActiveInstanceList(core.EmptyCtx(), 1722914086000000, 1722935686000000, "", []string{"ts-travel-plan-service"})
	if err != nil {
		t.Fatalf("Error to get active instance list: %v", err)
	}

	got := instances.GetInstanceIds()
	expect := []string{
		"ts-travel-plan-service-5dfc676467-km4zh",
		"ts-travel-plan-service-6659688b59-95t6c",
	}
	validator := util.NewValidator(t, "GetActiveInstanceList")
	validator.CheckIntValue("Instance Size", len(expect), len(got))
	for i, key := range expect {
		validator.CheckStringValue(fmt.Sprintf("Exist instance-%d", i), key, got[i])
	}
}
