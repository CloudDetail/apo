// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/CloudDetail/apo/backend/pkg/logger"
)

func NewTestRepo(t *testing.T) Repo {
	viper.SetConfigFile("testdata/config.yml")
	viper.ReadInConfig()

	address := viper.GetString("clickhouse.address")
	database := viper.GetString("clickhouse.database")
	username := viper.GetString("clickhouse.username")
	password := viper.GetString("clickhouse.password")

	zapLog := logger.NewLogger(logger.WithLevel("debug"))
	repo, err := New(zapLog, []string{address}, database, username, password)
	if err != nil {
		t.Fatalf("Error to connect clickhouse: %v", err)
	}
	return repo
}
