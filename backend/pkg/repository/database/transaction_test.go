// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	os.Setenv("APO_CONFIG", "../../../config/apo.yml")

	logCfg := config.Get().Logger
	accessLogger := logger.NewLogger(
		logger.WithConsole(logCfg.EnableConsole),
		logger.WithLevel(logCfg.Level),
		logger.WithTimeLayout(logger.CSTLayout),
		logger.WithFileRotationP(logCfg.EnableFile, logCfg.FilePath, logCfg.FileNum, logCfg.FileSize),
	)
	repo, err := New(accessLogger)
	if err != nil {
		return
	}

	var grantFunc = func(ctx context.Context) error {
		return repo.GrantRoleWithUser(nil, ctx, 239077004960, []int{2})
	}

	var boomFunc = func(ctx context.Context) error {
		return errors.New("boom")
	}

	err = repo.Transaction(nil, context.Background(), grantFunc, boomFunc)
	exists, checkErr := repo.RoleGrantedToUser(nil, 239077004960, 2)
	if checkErr != nil {
		t.Error(err)
	}
	assert.NotNil(t, err)
	assert.False(t, exists)
}
