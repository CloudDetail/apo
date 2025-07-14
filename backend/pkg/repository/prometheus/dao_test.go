// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"os"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func init() {
	os.Setenv("APO_CONFIG", "../../../config/apo.yml")
}

func TestRepo(t *testing.T) {
	cfg := config.Get().Promethues
	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	repo, err := New(zapLog, cfg.Address, cfg.Storage)
	if err != nil {
		t.Fatalf("Error to connect prometheus: %v", err)
	}

	testGetActiveInstanceList(t, repo)
}

func testGetActiveInstanceList(t *testing.T, repo Repo) {
	instances, err := repo.GetActiveInstanceList(core.EmptyCtx(), 1722914086000000, 1722935686000000, []string{"ts-travel-plan-service"})
	if err != nil {
		t.Errorf("Error to get active instance list: %v", err)
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
