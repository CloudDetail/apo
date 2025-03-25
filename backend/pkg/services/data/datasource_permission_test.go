// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/stretchr/testify/assert"
)

var s *service

func init() {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	err := os.Chdir(projectRoot)
	if err != nil {
		panic(err)
	}
	zapLog := logger.NewLogger(logger.WithLevel("debug"))

	dbRepo, err := database.New(zapLog)
	if err != nil {
		panic(err)
	}

	cfg := config.Get()
	promeRepo, err := prometheus.New(zapLog, cfg.Promethues.Address, cfg.Promethues.Storage)
	if err != nil {
		panic(err)
	}

	k8sRepo, err := kubernetes.New(zapLog, cfg.Kubernetes.AuthType, cfg.Kubernetes.AuthFilePath, config.MetadataSettings{})
	if err != nil {
		panic(err)
	}
	s = &service{
		dbRepo:   dbRepo,
		promRepo: promeRepo,
		k8sRepo:  k8sRepo,
	}
}

func TestDataSourcePermission(t *testing.T) {
	if s == nil {
		t.Fatal("service is nil")
		return
	}

	namespaceList := []string{}

	anonymousUser, err := s.dbRepo.GetAnonymousUser()
	if err != nil {
		t.Fatal(err)
	}

	err = s.CheckDatasourcePermission(anonymousUser.UserID, 0, &namespaceList, nil, model.DATASOURCE_CATEGORY_APM)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "", namespaceList[len(namespaceList)-1])
}
