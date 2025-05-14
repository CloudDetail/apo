// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"os"
	"testing"

	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
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

	var grantFunc = func(ctx core.Context) error {
		return repo.GrantRoleWithUser(nil, ctx, 239077004960, []int{2})
	}

	var boomFunc = func(ctx core.Context) error {
		return errors.New("boom")
	}

	err = repo.Transaction(nil, grantFunc, boomFunc)
	exists, checkErr := repo.RoleGrantedToUser(nil, 239077004960, 2)
	if checkErr != nil {
		t.Error(err)
	}
	assert.NotNil(t, err)
	assert.False(t, exists)
}
