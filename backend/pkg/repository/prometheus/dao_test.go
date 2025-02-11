// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/config"
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
	testGetMultiServicesInstanceList(t, repo)
}

func testGetActiveInstanceList(t *testing.T, repo Repo) {
	instances, err := repo.GetActiveInstanceList(1722914086000000, 1722935686000000, "ts-travel-plan-service", nil)
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

func testGetMultiServicesInstanceList(t *testing.T, repo Repo) {
	endTime := time.Now()
	startTime := endTime.Add(-time.Minute * 30)
	services := []string{"ts-basic-service", "ts-route-service", "${SPRING_APP_NAME:spring-reqtemplate-demo}"}
	ret, err := repo.GetMultiServicesInstanceList(startTime.UnixMicro(), endTime.UnixMicro(), services)
	if err != nil {
		t.Fatalf("GetMultiServicesInstanceList failed, err: %v", err)
	}
	retJson, err := json.Marshal(ret)
	if err != nil {
		t.Fatalf("Marshal result failed, err: %v", err)
	}
	t.Logf("GetMultiServicesInstanceList result: %s", retJson)
}
